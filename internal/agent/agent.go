package agent

import (
	"context"
	"fmt"
	"time"
)

type State string

const (
	StateIdle          State = "idle"
	StateAnalyzing     State = "analyzing"
	StatePlanGenerated State = "plan_generated"
	StateWritingTests  State = "writing_tests"
	StateRunning       State = "running"
	StateFixing        State = "fixing"
	StateDone          State = "done"
	StateFailed        State = "failed"
)

type TestPlan struct {
	Summary   string     `json:"summary"`
	Scenarios []Scenario `json:"scenarios"`
}

type Scenario struct {
	Name     string   `json:"name"`
	Priority string   `json:"priority"`
	Steps    []string `json:"steps"`
}

type TestFile struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type RunResult struct {
	Passed   int       `json:"passed"`
	Failed   int       `json:"failed"`
	Total    int       `json:"total"`
	Failures []Failure `json:"failures"`
}

type Failure struct {
	Test       string `json:"test"`
	Message    string `json:"message"`
	Screenshot string `json:"screenshot_url,omitempty"`
}

type TestRun struct {
	ID           string     `json:"id"`
	ProjectPath  string     `json:"project_path"`
	Requirements string     `json:"requirements"`
	Mode         string     `json:"mode"` // "simple" or "advanced"
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

type LLM interface {
	AnalyzeCodebase(ctx context.Context, path string) (string, error)
	GenerateTestPlan(ctx context.Context, analysis, requirements string) (*TestPlan, error)
	GenerateTestScripts(ctx context.Context, plan *TestPlan, analysis string) ([]TestFile, error)
	SuggestFixes(ctx context.Context, failures []Failure, files []TestFile) ([]TestFile, error)
}

type Runner interface {
	Run(ctx context.Context, testFiles []TestFile, projectURL string) (*RunResult, error)
}

// ScreenshotCapturer captures screenshots during test failures.
type ScreenshotCapturer interface {
	Capture(ctx context.Context, runID string, label string) (string, error)
}

// AgentConfig holds optional dependencies for the agent.
type AgentConfig struct {
	Sidecar    *SidecarClient
	Screenshotter ScreenshotCapturer
}

type Agent struct {
	llm            LLM
	runner         Runner
	maxFixAttempts int
	sidecar        *SidecarClient
	screenshotter  ScreenshotCapturer
}

func New(llm LLM, runner Runner, maxFixes int) *Agent {
	return &Agent{llm: llm, runner: runner, maxFixAttempts: maxFixes}
}

func NewWithConfig(llm LLM, runner Runner, maxFixes int, cfg AgentConfig) *Agent {
	return &Agent{
		llm:            llm,
		runner:         runner,
		maxFixAttempts: maxFixes,
		sidecar:        cfg.Sidecar,
		screenshotter:  cfg.Screenshotter,
	}
}

func (a *Agent) Execute(ctx context.Context, run *TestRun) error {
	if run.Mode == "" {
		run.Mode = "simple"
	}

	// Advanced mode: delegate to LangGraph sidecar
	if run.Mode == "advanced" && a.sidecar != nil {
		return a.executeAdvanced(ctx, run)
	}

	return a.executeSimple(ctx, run)
}

func (a *Agent) executeAdvanced(ctx context.Context, run *TestRun) error {
	// First do analysis locally so sidecar has context
	run.State = StateAnalyzing
	run.UpdatedAt = time.Now()

	analysis, err := a.llm.AnalyzeCodebase(ctx, run.ProjectPath)
	if err != nil {
		return a.fail(run, fmt.Errorf("analyze: %w", err))
	}
	run.CodeAnalysis = analysis

	// Delegate to sidecar
	jobID, err := a.sidecar.StartRun(ctx, run, a.maxFixAttempts)
	if err != nil {
		return a.fail(run, fmt.Errorf("sidecar start: %w", err))
	}

	// Poll for completion
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

func (a *Agent) executeSimple(ctx context.Context, run *TestRun) error {
	run.State = StateAnalyzing
	run.UpdatedAt = time.Now()

	analysis, err := a.llm.AnalyzeCodebase(ctx, run.ProjectPath)
	if err != nil {
		return a.fail(run, fmt.Errorf("analyze: %w", err))
	}
	run.CodeAnalysis = analysis

	run.State = StatePlanGenerated
	plan, err := a.llm.GenerateTestPlan(ctx, analysis, run.Requirements)
	if err != nil {
		return a.fail(run, fmt.Errorf("plan: %w", err))
	}
	run.TestPlan = plan

	run.State = StateWritingTests
	files, err := a.llm.GenerateTestScripts(ctx, plan, analysis)
	if err != nil {
		return a.fail(run, fmt.Errorf("write: %w", err))
	}
	run.TestFiles = files

	for {
		run.State = StateRunning
		result, err := a.runner.Run(ctx, run.TestFiles, run.ProjectPath)
		if err != nil {
			return a.fail(run, fmt.Errorf("run: %w", err))
		}
		run.RunResult = result

		if result.Failed == 0 || run.FixAttempts >= a.maxFixAttempts {
			break
		}

		// Capture screenshots on failure
		a.captureFailureScreenshots(ctx, run, result)

		run.State = StateFixing
		run.FixAttempts++
		fixed, err := a.llm.SuggestFixes(ctx, result.Failures, run.TestFiles)
		if err != nil {
			return a.fail(run, fmt.Errorf("fix: %w", err))
		}
		run.TestFiles = fixed
	}

	run.State = StateDone
	now := time.Now()
	run.FinishedAt = &now
	run.UpdatedAt = now
	return nil
}

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
	}
}

func (a *Agent) fail(run *TestRun, err error) error {
	run.State = StateFailed
	run.Error = err.Error()
	run.UpdatedAt = time.Now()
	return err
}
