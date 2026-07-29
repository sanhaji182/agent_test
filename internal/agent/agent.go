// Package agent berisi state machine utama untuk menjalankan test secara otomatis.
// Alur: idle → analyzing → planning → writing → running → fixing → done
package agent

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/events"
	"github.com/go-go-golems/gotest-agent/internal/execution"
)

// RunPersistence is a minimal interface for saving run state transitions.
// Defined here to avoid circular dependencies (db package imports agent).
type RunPersistence interface {
	CreateRun(ctx context.Context, run *TestRun) error
	UpdateRun(ctx context.Context, run *TestRun) error
}

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
	StateSimulated     State = "simulated"      // No real execution — synthetic result
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

// FeatureMap menyimpan fitur dan use case yang diturunkan dari PRD/requirements.
type FeatureMap struct {
	Source   string    `json:"source"`
	Features []Feature `json:"features"`
}

type Feature struct {
	Name     string   `json:"name"`
	UseCases []string `json:"use_cases"`
}

// TestFile adalah file test yang di-generate oleh LLM
type TestFile struct {
	Name    string `json:"name"`    // Nama file, misal "login.spec.ts"
	Content string `json:"content"` // Isi kode test
}

// RunResult adalah hasil eksekusi test
type RunResult struct {
	Passed     int       `json:"passed"`
	Failed     int       `json:"failed"`
	Total      int       `json:"total"`
	Failures   []Failure `json:"failures"`
	VideoPath  string    `json:"video_path,omitempty"`  // Path ke file video recording
	DurationMs int64     `json:"duration_ms,omitempty"` // Total execution time
	Healed     int       `json:"healed,omitempty"`      // Actions recovered by self-healing
	Retried    int       `json:"retried,omitempty"`     // Actions recovered by simple retry
}

// Failure menyimpan detail test yang gagal
type Failure struct {
	Test       string `json:"test"`
	Message    string `json:"message"`
	Screenshot string `json:"screenshot_url,omitempty"`
	DurationMs int64  `json:"duration_ms,omitempty"` // Time spent on this action before failure
}

// TestRun adalah objek utama yang merepresentasikan satu sesi pengujian
type TestRun struct {
	ID           string      `json:"id"`
	ProjectPath  string      `json:"project_path"`
	Requirements string      `json:"requirements"`
	Mode         string      `json:"mode"`                // "simple" atau "advanced"
	TestType     string      `json:"test_type,omitempty"` // "ui" atau "api"
	TestCaseID   string      `json:"test_case_id,omitempty"`
	TestListID   string      `json:"test_list_id,omitempty"`
	PRD          string      `json:"prd,omitempty"`
	APIDocs      string      `json:"api_docs,omitempty"`
	AuthType     string      `json:"auth_type,omitempty"`
	Credentials  string      `json:"credentials,omitempty"`
	FocusHints   string      `json:"focus_hints,omitempty"`
	SkipHints    string      `json:"skip_hints,omitempty"`
	FeatureMap   *FeatureMap `json:"feature_map,omitempty"`
	// Execution options (Phase 1+)
	Browser      string            `json:"browser,omitempty"`   // "chromium" (default), "firefox", "webkit"
	Viewport     string            `json:"viewport,omitempty"`  // viewport preset name (e.g. "iphone-14", "desktop-hd")
	Parallel     bool              `json:"parallel,omitempty"`  // execute test files concurrently
	TestData     map[string]string `json:"test_data,omitempty"` // parameterized test data ({{key}} expansion)
	State        State      `json:"state"`
	CodeAnalysis string     `json:"code_analysis,omitempty"`
	TestPlan     *TestPlan  `json:"test_plan,omitempty"`
	TestFiles    []TestFile `json:"test_files,omitempty"`
	RunResult    *RunResult `json:"run_result,omitempty"`
	Screenshots  []string   `json:"screenshots,omitempty"`
	FixAttempts  int        `json:"fix_attempts"`
	Error        string     `json:"error,omitempty"`
	// Video recording fields
	VideoURL             string  `json:"video_url,omitempty"`
	VideoStatus          string  `json:"video_status,omitempty"` // "recording", "ready", "failed", "none"
	VideoDuration        float64 `json:"video_duration,omitempty"`
	VideoSize            int64   `json:"video_size,omitempty"`
	VideoFailureMarkerAt float64 `json:"video_failure_marker_at,omitempty"` // timestamp in seconds where failure occurred
	// Timestamps
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

// LLM adalah interface untuk semua operasi yang membutuhkan AI/LLM
type LLM interface {
	AnalyzeCodebase(ctx context.Context, path string) (string, error)
	GenerateTestPlan(ctx context.Context, analysis, requirements string) (*TestPlan, error)
	GenerateTestScripts(ctx context.Context, plan *TestPlan, analysis string) ([]TestFile, error)
	SuggestFixes(ctx context.Context, failures []Failure, files []TestFile) ([]TestFile, error)
	HealAction(ctx context.Context, action string, domSnapshot string, errorMsg string) (string, error)
	HealActionWithVision(ctx context.Context, action string, domSnapshot string, errorMsg string, imageBase64 string) (string, error)
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
	Store         RunPersistence // Optional: auto-saves state transitions + panic-safe completion
}

// Agent adalah orchestrator utama yang menjalankan seluruh alur testing
type Agent struct {
	llm            LLM
	runner         Runner
	maxFixAttempts int
	sidecar        *SidecarClient
	screenshotter  ScreenshotCapturer
	exec           *execution.Context
	store          RunPersistence
}

// New membuat Agent baru dengan konfigurasi minimal
func New(llm LLM, runner Runner, maxFixes int) *Agent {
	return &Agent{llm: llm, runner: runner, maxFixAttempts: maxFixes}
}

// NewWithConfig membuat Agent dengan konfigurasi lengkap (sidecar + screenshot + events + store)
func NewWithConfig(llm LLM, runner Runner, maxFixes int, cfg AgentConfig) *Agent {
	return &Agent{
		llm:            llm,
		runner:         runner,
		maxFixAttempts: maxFixes,
		sidecar:        cfg.Sidecar,
		screenshotter:  cfg.Screenshotter,
		exec:           cfg.Exec,
		store:          cfg.Store,
	}
}

// Execute menjalankan test run. Jika mode=advanced, delegasi ke sidecar.
// When a RunPersistence store is configured, state transitions are automatically
// persisted and panics are recovered gracefully.
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

// Launch runs Execute in a background goroutine with panic recovery and automatic
// persistence. This replaces the server.launchRun wrapper — all 5 execution trigger
// paths (web API, webhook, MCP, schedule run-now, schedule due) can call Launch
// directly on a configured Agent (ADR-001).
func (a *Agent) Launch(run *TestRun) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("run execution panicked", "run_id", run.ID, "panic", r)
				run.State = StateFailed
				run.Error = fmt.Sprintf("execution panic: %v", r)
				run.FinishedAt = func() *time.Time { t := time.Now(); return &t }()
				run.UpdatedAt = time.Now()
				if a.store != nil {
					_ = a.store.UpdateRun(context.Background(), run)
				}
				if a.exec != nil && a.exec.Events != nil {
					a.exec.Events.Emit(run.ID, "run_failed", "failed", fmt.Sprintf("Run panicked: %v", r), nil)
				}
			}
		}()
		// Save final state on exit regardless of error
		err := a.Execute(context.Background(), run)
		if a.store != nil {
			_ = a.store.UpdateRun(context.Background(), run)
		}
		if err != nil {
			slog.Error("run execution failed", "run_id", run.ID, "error", err)
		}
	}()
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
	a.save(run)
	a.emit(run.ID, "run_started", "idle", "Run started", nil)

	// Langkah 1: Analisis kode project
	run.State = StateAnalyzing
	run.UpdatedAt = time.Now()
	a.save(run)
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
	a.save(run)
	a.emit(run.ID, "plan_generated", "plan_generated", fmt.Sprintf("Generated %d scenarios", len(plan.Scenarios)), nil)

	// Langkah 3: Generate file test Playwright
	run.State = StateWritingTests
	files, err := a.llm.GenerateTestScripts(ctx, plan, analysis)
	if err != nil {
		a.emit(run.ID, "run_failed", "writing_tests", err.Error(), nil)
		return a.fail(run, fmt.Errorf("write: %w", err))
	}
	run.TestFiles = files
	a.save(run)
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
		a.save(run)
		a.emit(run.ID, "test_started", "running", "Executing tests", nil)

		var result *RunResult
		var err error
		if run.TestType == "api" {
			apiRunner := NewAPIRunner()
			result, err = apiRunner.Run(ctx, run.TestFiles, run.ProjectPath)
		} else {
			result, err = a.runner.Run(ctx, run.TestFiles, run.ProjectPath)
		}

		if err != nil {
			a.emit(run.ID, "run_failed", "running", err.Error(), nil)
			return a.fail(run, fmt.Errorf("run: %w", err))
		}
		run.RunResult = result
		a.save(run)

		// Populate video fields jika runner menghasilkan video
		if result.VideoPath != "" {
			run.VideoURL = result.VideoPath
			run.VideoStatus = "ready"
			// Cari timestamp failure dari events (precise dari Playwright report)
			if result.Failed > 0 && a.exec != nil {
				for _, evt := range a.exec.Events.GetEvents(run.ID) {
					if string(evt.Type) == "step_completed" && evt.Metadata["status"] == "failed" {
						if ts := evt.Metadata["timestamp_ms"]; ts != "" {
							// Convert ms to seconds
							ms := 0
							for _, c := range ts {
								ms = ms*10 + int(c-'0')
							}
							run.VideoFailureMarkerAt = float64(ms) / 1000.0
							break
						}
					}
				}
			}
		}

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
		a.save(run)
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
	a.save(run)
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

// save persists the run's current state if a store is configured.
func (a *Agent) save(run *TestRun) {
	if a.store != nil {
		_ = a.store.UpdateRun(context.Background(), run)
	}
}

// fail menandai run sebagai gagal dan menyimpan pesan error
func (a *Agent) fail(run *TestRun, err error) error {
	run.State = StateFailed
	run.Error = err.Error()
	run.FinishedAt = func() *time.Time { t := time.Now(); return &t }()
	run.UpdatedAt = time.Now()
	a.save(run)
	return err
}
