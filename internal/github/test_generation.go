package github

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/go-go-golems/gotest-agent/internal/ai"
	"github.com/go-go-golems/gotest-agent/internal/parser"
	"github.com/google/uuid"
)

// TestGenerationService orchestrates test generation from GitHub webhook events
type TestGenerationService struct {
	integration *Integration
	parser      *parser.Registry
	llm         ai.Client
	agent       *agent.Agent
	store       agent.RunPersistence
}

// NewTestGenerationService creates a new test generation service
func NewTestGenerationService(
	integration *Integration,
	parserRegistry *parser.Registry,
	llmClient ai.Client,
	testAgent *agent.Agent,
	store agent.RunPersistence,
) *TestGenerationService {
	return &TestGenerationService{
		integration: integration,
		parser:      parserRegistry,
		llm:         llmClient,
		agent:       testAgent,
		store:       store,
	}
}

// ProcessPushEvent handles push webhook events and triggers test generation
func (s *TestGenerationService) ProcessPushEvent(ctx context.Context, pushEvent *PushEvent) error {
	log.Printf("Processing push event for %s (ref: %s)", pushEvent.Repository.FullName, pushEvent.Ref)

	// Clone or pull repository
	repoDir := filepath.Join(s.integration.cloneDir, pushEvent.Repository.FullName)
	branch := pushEvent.Ref
	if err := s.integration.ensureRepository(ctx, pushEvent.Repository.FullName, repoDir, branch); err != nil {
		return fmt.Errorf("failed to ensure repository: %w", err)
	}

	// Collect changed files from commits
	changedFiles := make([]string, 0)
	for _, commit := range pushEvent.Commits {
		changedFiles = append(changedFiles, commit.Added...)
		changedFiles = append(changedFiles, commit.Modified...)
	}

	if len(changedFiles) == 0 {
		log.Println("No changed files in push event, skipping test generation")
		return nil
	}

	log.Printf("Detected %d changed files, triggering test generation", len(changedFiles))

	// Parse changed files
	codebase, err := s.parser.Parse(ctx, repoDir)
	if err != nil {
		return fmt.Errorf("failed to parse codebase: %w", err)
	}

	// Generate test plan using AI synthesis
	synthesisService := ai.NewSynthesisService(s.llm)
	testPlan, err := synthesisService.GenerateTestPlan(ctx, codebase)
	if err != nil {
		return fmt.Errorf("failed to generate test plan: %w", err)
	}

	// Create a TestRun for this push event
	testRun := &agent.TestRun{
		ID:           uuid.New().String(),
		ProjectPath:  repoDir,
		Requirements: fmt.Sprintf("Test changes from push to %s", pushEvent.Ref),
		Mode:         "simple",
		TestType:     "ui",
		State:        agent.StateIdle,
		TestPlan: &agent.TestPlan{
			Summary: fmt.Sprintf("Generated test plan for %d test cases", len(testPlan.Tests)),
			Scenarios: s.convertTestCasesToScenarios(testPlan.Tests),
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save the test run
	if s.store != nil {
		if err := s.store.CreateRun(ctx, testRun); err != nil {
			return fmt.Errorf("failed to create test run: %w", err)
		}
	}

	// Launch the test execution
	log.Printf("Launching test execution for run %s", testRun.ID)
	s.agent.Launch(testRun)

	return nil
}

// ProcessPullRequestEvent handles pull request webhook events and triggers test generation
func (s *TestGenerationService) ProcessPullRequestEvent(ctx context.Context, prEvent *PullRequestEvent) error {
	log.Printf("Processing pull request event for %s #%d (action: %s)",
		prEvent.Repository.FullName, prEvent.Number, prEvent.Action)

	// Only process opened, synchronize, and reopened actions
	switch prEvent.Action {
	case "opened", "synchronize", "reopened":
		// Clone repository and checkout PR branch
		repoDir := filepath.Join(s.integration.cloneDir, prEvent.Repository.FullName)
		if err := s.integration.ensureRepository(ctx, prEvent.Repository.FullName, repoDir, prEvent.PullRequest.Head.Ref); err != nil {
			return fmt.Errorf("failed to ensure repository: %w", err)
		}

		// Parse the codebase
		codebase, err := s.parser.Parse(ctx, repoDir)
		if err != nil {
			return fmt.Errorf("failed to parse codebase: %w", err)
		}

		// Generate test plan
		synthesisService := ai.NewSynthesisService(s.llm)
		testPlan, err := synthesisService.GenerateTestPlan(ctx, codebase)
		if err != nil {
			return fmt.Errorf("failed to generate test plan: %w", err)
		}

		// Create a TestRun for this PR
		testRun := &agent.TestRun{
			ID:           uuid.New().String(),
			ProjectPath:  repoDir,
			Requirements: fmt.Sprintf("Test changes in PR #%d: %s", prEvent.Number, prEvent.PullRequest.Title),
			Mode:         "simple",
			TestType:     "ui",
			State:        agent.StateIdle,
			TestPlan: &agent.TestPlan{
				Summary: fmt.Sprintf("Generated test plan for %d test cases", len(testPlan.Tests)),
				Scenarios: s.convertTestCasesToScenarios(testPlan.Tests),
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Save the test run
		if s.store != nil {
			if err := s.store.CreateRun(ctx, testRun); err != nil {
				return fmt.Errorf("failed to create test run: %w", err)
			}
		}

		// Launch the test execution
		log.Printf("Launching test execution for PR #%d, run %s", prEvent.Number, testRun.ID)
		s.agent.Launch(testRun)

	case "closed":
		if prEvent.PullRequest.Merged {
			log.Printf("PR #%d was merged", prEvent.Number)
			// TODO: Could trigger full test suite after merge
		} else {
			log.Printf("PR #%d was closed without merging", prEvent.Number)
		}

	default:
		log.Printf("Ignoring PR action: %s", prEvent.Action)
	}

	return nil
}

// convertTestCasesToScenarios converts AI test cases to agent scenarios
func (s *TestGenerationService) convertTestCasesToScenarios(testCases []ai.TestCase) []agent.Scenario {
	scenarios := make([]agent.Scenario, len(testCases))
	for i, tc := range testCases {
		scenarios[i] = agent.Scenario{
			Name:     tc.Name,
			Priority: tc.Priority,
			Steps:    []string{tc.Description}, // Convert description to steps
		}
	}
	return scenarios
}
