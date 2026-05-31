// Package agent berisi state machine utama untuk menjalankan test secara otomatis.
// Alur: idle → analyzing → planning → writing → running → fixing → done
package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/events"
	"github.com/go-go-golems/gotest-agent/internal/execution"
)

// State merepresentasikan status dari sebuah test run
type State string

const (
	StateIdle          State = "idle"           // Belum dimulai
	StateAnalyzing     State = "analyzing"      // Sedang menganalisis kode
	StatePlanGenerated State = "plan_generated" // Test plan sudah dibuat
	StateWritingTests  State = "writing_tests"  // Sedang menulis file test
	StateRunning       State = "running"        // Sedang menjalankan test
	StateFixing        State = "fixing"         // Sedang memperbaiki test yang gagal
	StateDone          State = "done"           // Selesai
	StateFailed        State = "failed"         // Gagal
)

// TestPlan adalah rencana pengujian yang dihasilkan oleh LLM
type TestPlan struct {
	Summary   string     `json:"summary"`
	Scenarios []Scenario `json:"scenarios"`
}

// Scenario adalah satu skenario test dalam test plan
type Scenario struct {
	Name     string   `json:"name"`
	Priority string   `json:"priority"` // high, medium, low
	Steps    []string `json:"steps"`
}

// TestFile adalah file test yang di-generate oleh LLM
type TestFile struct {
	Name    string `json:"name"`    // Nama file, misal "login.spec.ts"
	Content string `json:"content"` // Isi kode test
}

// RunResult adalah hasil eksekusi test
type RunResult struct {
	Passed   int       `json:"passed"`
	Failed   int       `json:"failed"`
	Total    int       `json:"total"`
	Failures []Failure `json:"failures"`
}

// Failure menyimpan detail test yang gagal
type Failure struct {
	Test       string `json:"test"`
	Message    string `json:"message"`
	Screenshot string `json:"screenshot_url,omitempty"`
}

// TestRun adalah objek utama yang merepresentasikan satu sesi pengujian
type TestRun struct {
	ID           string     `json:"id"`
	ProjectPath  string     `json:"project_path"`
	Requirements string     `json:"requirements"`
	Mode         string     `json:"mode"` // "simple" atau "advanced"
	State        State      `json:"state"`
	CodeAnalysis string     `json:"code_analysis,omitempty"`
	TestPlan     *TestPlan  `json:"test_plan,omitempty"`
	TestFiles    []TestFile `json:"test_files,omitempty"`
	RunResult    *RunResult `json:"run_result,omitempty"`
	Screenshots  []string   `json:"screenshots,omitempty"`
	FixAttempts  int        `json:"fix_attempts"`
	Error        string     `json:"error,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
}

// LLM adalah interface untuk semua operasi yang membutuhkan AI/LLM
type LLM interface {
	AnalyzeCodebase(ctx context.Context, path string) (string, error)
	GenerateTestPlan(ctx context.Context, analysis, requirements string) (*TestPlan, error)
	GenerateTestScripts(ctx context.Context, plan *TestPlan, analysis string) ([]TestFile, error)
	SuggestFixes(ctx context.Context, failures []Failure, files []TestFile) ([]TestFile, error)
}

// Runner adalah interface untuk menjalankan test (Docker atau Steel)
type Runner interface {
	Run(ctx context.Context, testFiles []TestFile, projectURL string) (*RunResult, error)
}

// ScreenshotCapturer adalah interface untuk mengambil screenshot saat test gagal
type ScreenshotCapturer interface {
	Capture(ctx context.Context, runID string, label string) (string, error)
}

// AgentConfig menyimpan dependensi opsional untuk agent
type AgentConfig struct {
	Sidecar       *SidecarClient
	Screenshotter ScreenshotCapturer
	Exec          *execution.Context
}

// Agent adalah orchestrator utama yang menjalankan seluruh alur testing
type Agent struct {
	llm            LLM
	runner         Runner
	maxFixAttempts int
	sidecar        *SidecarClient
	screenshotter  ScreenshotCapturer
	exec           *execution.Context
}

// New membuat Agent baru dengan konfigurasi minimal
func New(llm LLM, runner Runner, maxFixes int) *Agent {
	return &Agent{llm: llm, runner: runner, maxFixAttempts: maxFixes}
}

// NewWithConfig membuat Agent dengan konfigurasi lengkap (sidecar + screenshot + events)
func NewWithConfig(llm LLM, runner Runner, maxFixes int, cfg AgentConfig) *Agent {
	return &Agent{
		llm:            llm,
		runner:         runner,
		maxFixAttempts: maxFixes,
		sidecar:        cfg.Sidecar,
		screenshotter:  cfg.Screenshotter,
		exec:           cfg.Exec,
	}
}

// Execute menjalankan test run. Jika mode=advanced, delegasi ke sidecar.
func (a *Agent) Execute(ctx context.Context, run *TestRun) error {
	if run.Mode == "" {
		run.Mode = "simple"
	}

	// Mode advanced: delegasi ke LangGraph sidecar (multi-agent)
	if run.Mode == "advanced" && a.sidecar != nil {
		return a.executeAdvanced(ctx, run)
	}

	return a.executeSimple(ctx, run)
}

// executeAdvanced mendelegasikan eksekusi ke Python LangGraph sidecar
func (a *Agent) executeAdvanced(ctx context.Context, run *TestRun) error {
	// Analisis dulu secara lokal supaya sidecar punya konteks
	run.State = StateAnalyzing
	run.UpdatedAt = time.Now()

	analysis, err := a.llm.AnalyzeCodebase(ctx, run.ProjectPath)
	if err != nil {
		return a.fail(run, fmt.Errorf("analyze: %w", err))
	}
	run.CodeAnalysis = analysis

	// Kirim ke sidecar untuk diproses multi-agent
	jobID, err := a.sidecar.StartRun(ctx, run, a.maxFixAttempts)
	if err != nil {
		return a.fail(run, fmt.Errorf("sidecar start: %w", err))
	}

	// Polling status sampai selesai
	run.State = StateRunning
	for {
		select {
		case <-ctx.Done():
			return a.fail(run, ctx.Err())
		case <-time.After(2 * time.Second):
		}

		status, err := a.sidecar.GetStatus(ctx, jobID)
		if err != nil {
			return a.fail(run, fmt.Errorf("sidecar status: %w", err))
		}

		switch status.Status {
		case "completed":
			run.State = StateDone
			now := time.Now()
			run.FinishedAt = &now
			run.UpdatedAt = now
			return nil
		case "failed":
			return a.fail(run, fmt.Errorf("sidecar failed: %s", status.Error))
		}
	}
}

// executeSimple menjalankan alur testing langsung di Go (tanpa sidecar)
func (a *Agent) executeSimple(ctx context.Context, run *TestRun) error {
	a.emit(run.ID, "run_started", "idle", "Run started", nil)

	// Langkah 1: Analisis kode project
	run.State = StateAnalyzing
	run.UpdatedAt = time.Now()
	a.emit(run.ID, "analysis_started", "analyzing", "Analyzing codebase", nil)

	analysis, err := a.llm.AnalyzeCodebase(ctx, run.ProjectPath)
	if err != nil {
		a.emit(run.ID, "run_failed", "analyzing", err.Error(), nil)
		return a.fail(run, fmt.Errorf("analyze: %w", err))
	}
	run.CodeAnalysis = analysis
	a.emit(run.ID, "analysis_completed", "analyzing", "Analysis complete", nil)

	// Langkah 2: Buat test plan dari hasil analisis
	run.State = StatePlanGenerated
	plan, err := a.llm.GenerateTestPlan(ctx, analysis, run.Requirements)
	if err != nil {
		a.emit(run.ID, "run_failed", "plan_generated", err.Error(), nil)
		return a.fail(run, fmt.Errorf("plan: %w", err))
	}
	run.TestPlan = plan
	a.emit(run.ID, "plan_generated", "plan_generated", fmt.Sprintf("Generated %d scenarios", len(plan.Scenarios)), nil)

	// Langkah 3: Generate file test Playwright
	run.State = StateWritingTests
	files, err := a.llm.GenerateTestScripts(ctx, plan, analysis)
	if err != nil {
		a.emit(run.ID, "run_failed", "writing_tests", err.Error(), nil)
		return a.fail(run, fmt.Errorf("write: %w", err))
	}
	run.TestFiles = files
	a.emit(run.ID, "script_generated", "writing_tests", fmt.Sprintf("Generated %d test files", len(files)), nil)

	// Langkah 4: Jalankan test + fix loop (maks 3x percobaan)
	// Set execution context pada runner jika didukung
	type execSetter interface {
		SetExecContext(exec *execution.Context, runID string)
	}
	if es, ok := a.runner.(execSetter); ok && a.exec != nil {
		es.SetExecContext(a.exec, run.ID)
	}

	for {
		run.State = StateRunning
		a.emit(run.ID, "test_started", "running", "Executing tests", nil)

		result, err := a.runner.Run(ctx, run.TestFiles, run.ProjectPath)
		if err != nil {
			a.emit(run.ID, "run_failed", "running", err.Error(), nil)
			return a.fail(run, fmt.Errorf("run: %w", err))
		}
		run.RunResult = result

		// Emit per-assertion events
		for _, f := range result.Failures {
			a.emit(run.ID, "assertion_failed", "running", f.Message, map[string]string{"test": f.Test})
		}
		if result.Passed > 0 {
			a.emit(run.ID, "assertion_passed", "running", fmt.Sprintf("%d tests passed", result.Passed), nil)
		}

		// Jika semua pass atau sudah melebihi batas fix, selesai
		if result.Failed == 0 || run.FixAttempts >= a.maxFixAttempts {
			break
		}

		// Ambil screenshot untuk setiap test yang gagal
		a.captureFailureScreenshots(ctx, run, result)

		// Minta LLM untuk memperbaiki test yang gagal
		run.State = StateFixing
		run.FixAttempts++
		a.emit(run.ID, "fix_attempt_started", "fixing", fmt.Sprintf("Fix attempt %d", run.FixAttempts), nil)

		fixed, err := a.llm.SuggestFixes(ctx, result.Failures, run.TestFiles)
		if err != nil {
			a.emit(run.ID, "run_failed", "fixing", err.Error(), nil)
			return a.fail(run, fmt.Errorf("fix: %w", err))
		}
		run.TestFiles = fixed
		a.emit(run.ID, "fix_attempt_completed", "fixing", "Fix applied, re-running", nil)
	}

	run.State = StateDone
	now := time.Now()
	run.FinishedAt = &now
	run.UpdatedAt = now
	a.emit(run.ID, "run_completed", "done", "Run completed", nil)
	return nil
}

// captureFailureScreenshots mengambil screenshot dan otomatis membuat recording + visual artifact
func (a *Agent) captureFailureScreenshots(ctx context.Context, run *TestRun, result *RunResult) {
	if a.screenshotter == nil {
		return
	}
	for i, f := range result.Failures {
		label := fmt.Sprintf("failure-%d-%s", run.FixAttempts, f.Test)
		url, err := a.screenshotter.Capture(ctx, run.ID, label)
		if err != nil {
			continue
		}
		result.Failures[i].Screenshot = url
		run.Screenshots = append(run.Screenshots, url)
		// RecordScreenshot membuat recording + visual artifact + emit event
		if a.exec != nil {
			a.exec.RecordScreenshot(run.ID, f.Test, label, url)
		}
	}
}

// emit mengirim event jika execution context dikonfigurasi
func (a *Agent) emit(runID, eventType, phase, message string, metadata map[string]string) {
	if a.exec != nil && a.exec.Events != nil {
		a.exec.Events.Emit(runID, events.EventType(eventType), phase, message, metadata)
	}
}

// fail menandai run sebagai gagal dan menyimpan pesan error
func (a *Agent) fail(run *TestRun, err error) error {
	run.State = StateFailed
	run.Error = err.Error()
	run.UpdatedAt = time.Now()
	return err
}
