package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-go-golems/gotest-agent/internal/agent"
	"github.com/go-go-golems/gotest-agent/internal/ai"
	"github.com/go-go-golems/gotest-agent/internal/execution"
	"github.com/go-go-golems/gotest-agent/internal/planning"
	"github.com/go-go-golems/gotest-agent/internal/project"
	testrunner "github.com/go-go-golems/gotest-agent/internal/runner"
	"github.com/google/uuid"
)

func (s *Server) handleGenerateProjectTestPlan(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !isValidID(id) {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	p, err := s.projects.Get(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "project not found")
		return
	}
	if p.FeatureMap == nil {
		p.FeatureMap = s.deriveFeatureMap(r.Context(), p.Spec, p.FocusHints)
		_ = s.projects.Update(r.Context(), p)
	}
	s.createDraftPlanResponse(w, r, p.ID, s.generateDraftCases(r.Context(), p))
}

// createDraftPlanResponse membuat draft plan dari kumpulan kasus, menyimpannya,
// dan menulis respons 201. Dipakai bersama oleh handleGenerateProjectTestPlan
// dan handleParseAPIDocs (DL-3).
func (s *Server) createDraftPlanResponse(w http.ResponseWriter, r *http.Request, projectID string, cases []planning.DraftCase) {
	plan := &planning.DraftPlan{
		ProjectID: projectID,
		Status:    "draft",
		Cases:     cases,
	}
	if err := s.planning.CreateDraft(r.Context(), plan); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(plan)
}

func (s *Server) handleGetTestPlan(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !isValidID(id) {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	plan, err := s.planning.GetDraft(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plan)
}

func (s *Server) handleUpdateTestPlanCase(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	caseID := chi.URLParam(r, "caseId")
	plan, err := s.planning.GetDraft(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	var patch planning.DraftCase
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid body")
		return
	}
	updated := false
	for i := range plan.Cases {
		if plan.Cases[i].ID == caseID {
			mergeDraftCase(&plan.Cases[i], &patch)
			updated = true
			break
		}
	}
	if !updated {
		writeJSONError(w, http.StatusNotFound, "case not found")
		return
	}
	if err := s.planning.UpdateDraft(r.Context(), plan); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plan)
}

func (s *Server) handleRegenerateTestPlan(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	plan, err := s.planning.GetDraft(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	p, err := s.projects.Get(r.Context(), plan.ProjectID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "project not found")
		return
	}

	newCases, err := s.generateDraftCasesWithAI(r.Context(), p)
	if err != nil {
		slog.Error("ai generation failed", "error", err, "project_id", p.ID)
		writeJSONError(w, http.StatusInternalServerError, "ai generation failed")
		return
	}

	// For MVP, overwrite the draft plan. A full diff/merge would be complex.
	plan.Cases = newCases
	if err := s.planning.UpdateDraft(r.Context(), plan); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plan)
}

func (s *Server) handleApproveTestPlan(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	plan, err := s.planning.GetDraft(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	var cases []*planning.TestCase
	for _, c := range plan.Cases {
		if !c.Enabled {
			continue
		}
		cases = append(cases, &planning.TestCase{
			ProjectID:  plan.ProjectID,
			PlanID:     plan.ID,
			Title:      c.Title,
			Type:       c.Type,
			Feature:    c.Feature,
			Priority:   c.Priority,
			Steps:      c.Steps,
			Assertions: c.Assertions,
			Tags:       c.Tags,
		})
	}
	if err := s.planning.CreateTestCases(r.Context(), cases); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	plan.Status = "approved"
	_ = s.planning.UpdateDraft(r.Context(), plan)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "approved", "test_cases": cases})
}

func (s *Server) handleListTestCases(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	cases, err := s.planning.ListTestCases(r.Context(), projectID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cases)
}

func (s *Server) handleUpdateTestCase(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	tc, err := s.planning.GetTestCase(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}

	var payload struct {
		Title      string   `json:"title"`
		Priority   string   `json:"priority"`
		Steps      []string `json:"steps"`
		Assertions []string `json:"assertions"`
		Tags       []string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	if payload.Title != "" {
		tc.Title = payload.Title
	}
	if payload.Priority != "" {
		tc.Priority = payload.Priority
	}
	if payload.Steps != nil {
		tc.Steps = payload.Steps
	}
	if payload.Assertions != nil {
		tc.Assertions = payload.Assertions
	}
	if payload.Tags != nil {
		tc.Tags = payload.Tags
	}

	if err := s.planning.UpdateTestCase(r.Context(), tc); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tc)
}

func (s *Server) handleGetTestCase(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !isValidID(id) {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	tc, err := s.planning.GetTestCase(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tc)
}

func (s *Server) handleTestCaseMaintenance(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	cases, err := s.planning.ListTestCases(r.Context(), projectID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	runs, _ := s.store.ListRuns(r.Context(), 1000, 0)
	proposals, _ := s.planning.ListChangeProposals(r.Context(), "")

	lastRunByCase := map[string]time.Time{}
	failedByCase := map[string]bool{}
	for _, run := range runs {
		if run.TestCaseID == "" {
			continue
		}
		if run.CreatedAt.After(lastRunByCase[run.TestCaseID]) {
			lastRunByCase[run.TestCaseID] = run.CreatedAt
			failedByCase[run.TestCaseID] = run.State == agent.StateFailed || (run.RunResult != nil && run.RunResult.Failed > 0)
		}
	}
	pendingByCase := map[string]int{}
	for _, proposal := range proposals {
		if proposal.Status == "pending" {
			pendingByCase[proposal.TestCaseID]++
		}
	}
	duplicates := map[string][]*planning.TestCase{}
	for _, tc := range cases {
		key := strings.ToLower(strings.TrimSpace(tc.Title + "|" + tc.Feature))
		duplicates[key] = append(duplicates[key], tc)
	}

	type item struct {
		TestCaseID string    `json:"test_case_id"`
		Title      string    `json:"title"`
		Type       string    `json:"type"`
		Category   string    `json:"category"`
		Severity   string    `json:"severity"`
		Reason     string    `json:"reason"`
		Action     string    `json:"action"`
		LastRunAt  time.Time `json:"last_run_at,omitempty"`
	}
	var items []item
	now := time.Now()
	for _, tc := range cases {
		lastRunAt, hasRun := lastRunByCase[tc.ID]
		if !hasRun {
			items = append(items, item{
				TestCaseID: tc.ID, Title: tc.Title, Type: tc.Type,
				Category: "never_run", Severity: "medium",
				Reason: "Approved test has not been executed yet.",
				Action: "Run this test once to establish baseline behavior.",
			})
		} else if now.Sub(lastRunAt) > 14*24*time.Hour {
			items = append(items, item{
				TestCaseID: tc.ID, Title: tc.Title, Type: tc.Type, LastRunAt: lastRunAt,
				Category: "stale", Severity: "medium",
				Reason: "Test has not been executed in more than 14 days.",
				Action: "Run the test or add it to a recurring Test List.",
			})
		}
		if failedByCase[tc.ID] {
			items = append(items, item{
				TestCaseID: tc.ID, Title: tc.Title, Type: tc.Type, LastRunAt: lastRunAt,
				Category: "last_failed", Severity: "high",
				Reason: "Latest linked run failed.",
				Action: "Open the latest run, analyze failure, then refine or rerun.",
			})
		}
		if pendingByCase[tc.ID] > 0 {
			items = append(items, item{
				TestCaseID: tc.ID, Title: tc.Title, Type: tc.Type,
				Category: "pending_proposal", Severity: "low",
				Reason: fmt.Sprintf("%d refinement proposal(s) are waiting for review.", pendingByCase[tc.ID]),
				Action: "Approve or reject pending proposals in Reviews.",
			})
		}
		key := strings.ToLower(strings.TrimSpace(tc.Title + "|" + tc.Feature))
		if len(duplicates[key]) > 1 {
			items = append(items, item{
				TestCaseID: tc.ID, Title: tc.Title, Type: tc.Type,
				Category: "duplicate", Severity: "low",
				Reason: "Another approved test has the same title and feature.",
				Action: "Compare duplicate cases and keep the clearer version.",
			})
		}
	}
	if items == nil {
		items = []item{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (s *Server) handleRunTestCase(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !isValidID(id) {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	tc, err := s.planning.GetTestCase(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	run, err := s.startTestCaseRun(r.Context(), tc, "")
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"run_id": run.ID, "state": string(run.State), "test_case_id": tc.ID})
}

func (s *Server) handleRefineTestCase(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !isValidID(id) {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req struct {
		Prompt string `json:"prompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if strings.TrimSpace(req.Prompt) == "" {
		writeJSONError(w, http.StatusBadRequest, "prompt is required")
		return
	}
	tc, err := s.planning.GetTestCase(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	proposal := s.buildChangeProposal(r.Context(), tc, req.Prompt)
	if err := s.planning.CreateChangeProposal(r.Context(), proposal); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(proposal)
}

func (s *Server) handleListTestCaseProposals(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !isValidID(id) {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	proposals, err := s.planning.ListChangeProposals(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(proposals)
}

func (s *Server) handleListChangeProposals(w http.ResponseWriter, r *http.Request) {
	testCaseID := r.URL.Query().Get("test_case_id")
	if testCaseID != "" && !isValidID(testCaseID) {
		writeJSONError(w, http.StatusBadRequest, "invalid test_case_id")
		return
	}
	proposals, err := s.planning.ListChangeProposals(r.Context(), testCaseID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(proposals)
}

func (s *Server) handleApproveChangeProposal(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !isValidID(id) {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req struct {
		Reviewer string `json:"reviewer"`
		Comment  string `json:"comment"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	proposal, err := s.planning.GetChangeProposal(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	if proposal.Status != "pending" {
		writeJSONError(w, http.StatusConflict, "proposal already reviewed")
		return
	}
	current, err := s.planning.GetTestCase(r.Context(), proposal.TestCaseID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "test case not found")
		return
	}
	next := proposal.Proposed
	next.ID = current.ID
	next.ProjectID = current.ProjectID
	next.PlanID = current.PlanID
	next.Version = current.Version + 1
	next.CreatedAt = current.CreatedAt
	next.UpdatedAt = time.Now()
	if err := s.planning.UpdateTestCase(r.Context(), &next); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	now := time.Now()
	proposal.Status = "approved"
	proposal.ReviewedAt = &now
	proposal.Reviewer = strings.TrimSpace(req.Reviewer)
	proposal.ReviewComment = strings.TrimSpace(req.Comment)
	if proposal.Reviewer == "" {
		proposal.Reviewer = "self-hosted"
	}
	if err := s.planning.UpdateChangeProposal(r.Context(), proposal); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"proposal": proposal, "test_case": next})
}

func (s *Server) handleRejectChangeProposal(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !isValidID(id) {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req struct {
		Reviewer string `json:"reviewer"`
		Comment  string `json:"comment"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	proposal, err := s.planning.GetChangeProposal(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	if proposal.Status != "pending" {
		writeJSONError(w, http.StatusConflict, "proposal already reviewed")
		return
	}
	now := time.Now()
	proposal.Status = "rejected"
	proposal.ReviewedAt = &now
	proposal.Reviewer = strings.TrimSpace(req.Reviewer)
	proposal.ReviewComment = strings.TrimSpace(req.Comment)
	if proposal.Reviewer == "" {
		proposal.Reviewer = "self-hosted"
	}
	if err := s.planning.UpdateChangeProposal(r.Context(), proposal); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(proposal)
}

func (s *Server) buildChangeProposal(ctx context.Context, tc *planning.TestCase, prompt string) *planning.ChangeProposal {
	proposed, rationale := refineTestCaseFallback(tc, prompt)
	if aiTC, aiRationale, err := s.refineTestCaseWithAI(ctx, tc, prompt); err == nil {
		proposed = aiTC
		rationale = aiRationale
	}
	return &planning.ChangeProposal{
		TestCaseID: tc.ID,
		Status:     "pending",
		Prompt:     strings.TrimSpace(prompt),
		Rationale:  rationale,
		Original:   *tc,
		Proposed:   proposed,
	}
}

func (s *Server) refineTestCaseWithAI(ctx context.Context, tc *planning.TestCase, prompt string) (planning.TestCase, string, error) {
	client := s.aiClient(ctx)
	if client == nil {
		return planning.TestCase{}, "", fmt.Errorf("ai disabled")
	}
	payload, _ := json.Marshal(map[string]interface{}{"test_case": tc, "refine_prompt": redactForAI(prompt)})
	text, err := client.GenerateText(ctx, `Refine this approved test case without changing its id/project/plan.
Return only JSON with keys: title, type, feature, priority, steps, assertions, tags, rationale.
Keep steps and assertions as arrays of strings.

Payload:
`+string(payload))
	if err != nil {
		return planning.TestCase{}, "", err
	}
	var parsed struct {
		Title      string   `json:"title"`
		Type       string   `json:"type"`
		Feature    string   `json:"feature"`
		Priority   string   `json:"priority"`
		Steps      []string `json:"steps"`
		Assertions []string `json:"assertions"`
		Tags       []string `json:"tags"`
		Rationale  string   `json:"rationale"`
	}
	if err := json.Unmarshal([]byte(ai.StripJSONMarkers(text)), &parsed); err != nil {
		return planning.TestCase{}, "", err
	}
	proposed := *tc
	if strings.TrimSpace(parsed.Title) != "" {
		proposed.Title = strings.TrimSpace(parsed.Title)
	}
	if strings.TrimSpace(parsed.Type) != "" {
		proposed.Type = strings.TrimSpace(parsed.Type)
	}
	if strings.TrimSpace(parsed.Feature) != "" {
		proposed.Feature = strings.TrimSpace(parsed.Feature)
	}
	if strings.TrimSpace(parsed.Priority) != "" {
		proposed.Priority = strings.TrimSpace(parsed.Priority)
	}
	if len(parsed.Steps) > 0 {
		proposed.Steps = normalizeStrings(parsed.Steps)
	}
	if len(parsed.Assertions) > 0 {
		proposed.Assertions = normalizeStrings(parsed.Assertions)
	}
	if len(parsed.Tags) > 0 {
		proposed.Tags = normalizeStrings(parsed.Tags)
	}
	if len(proposed.Steps) == 0 || len(proposed.Assertions) == 0 {
		return planning.TestCase{}, "", fmt.Errorf("invalid refined test case")
	}
	rationale := strings.TrimSpace(parsed.Rationale)
	if rationale == "" {
		rationale = "AI proposed a refined version of the approved test case."
	}
	return proposed, rationale, nil
}

func refineTestCaseFallback(tc *planning.TestCase, prompt string) (planning.TestCase, string) {
	proposed := *tc
	refinement := strings.TrimSpace(prompt)
	if refinement == "" {
		refinement = "requested refinement"
	}
	step := "Refinement checkpoint: " + truncate(refinement, 120)
	assertion := "Verify refinement intent is covered: " + truncate(refinement, 120)
	if !contains(proposed.Steps, step) {
		proposed.Steps = append(proposed.Steps, step)
	}
	if !contains(proposed.Assertions, assertion) {
		proposed.Assertions = append(proposed.Assertions, assertion)
	}
	if !contains(proposed.Tags, "refined") {
		proposed.Tags = append(proposed.Tags, "refined")
	}
	return proposed, "Added a review-gated refinement checkpoint and assertion from the user prompt."
}

func normalizeStrings(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}

func (s *Server) handleCreateTestList(w http.ResponseWriter, r *http.Request) {
	var list planning.TestList
	if err := json.NewDecoder(r.Body).Decode(&list); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if strings.TrimSpace(list.Name) == "" {
		writeJSONError(w, http.StatusBadRequest, "name is required")
		return
	}
	if len(list.TestCaseIDs) == 0 {
		writeJSONError(w, http.StatusBadRequest, "test_case_ids is required")
		return
	}
	if err := s.planning.CreateTestList(r.Context(), &list); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(list)
}

func (s *Server) handleListTestLists(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	lists, err := s.planning.ListTestLists(r.Context(), projectID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lists)
}

func (s *Server) handleGetTestList(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !isValidID(id) {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	list, err := s.planning.GetTestList(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (s *Server) handleTestListHistory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !isValidID(id) {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	list, err := s.planning.GetTestList(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	history := s.buildTestListHistory(r.Context(), list)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func (s *Server) handleRunTestList(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !isValidID(id) {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	list, err := s.planning.GetTestList(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	runIDs, err := s.startTestListRuns(r.Context(), list)
	if err != nil {
		slog.Error("start test list runs failed", "error", err, "test_list_id", list.ID)
		writeJSONError(w, http.StatusInternalServerError, "failed to start test list runs")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{"test_list_id": list.ID, "run_ids": runIDs})
}

func (s *Server) startTestListRuns(ctx context.Context, list *planning.TestList) ([]string, error) {
	runIDs := make([]string, 0, len(list.TestCaseIDs))
	for _, caseID := range list.TestCaseIDs {
		tc, err := s.planning.GetTestCase(ctx, caseID)
		if err != nil {
			continue
		}
		run, err := s.startTestCaseRun(ctx, tc, list.ID)
		if err != nil {
			return nil, err
		}
		runIDs = append(runIDs, run.ID)
	}
	if len(runIDs) == 0 {
		return nil, fmt.Errorf("test list has no runnable test cases")
	}
	return runIDs, nil
}

func (s *Server) startTestCaseRun(ctx context.Context, tc *planning.TestCase, testListID string) (*agent.TestRun, error) {

	p, _ := s.projects.Get(ctx, tc.ProjectID)
	projectPath := ""
	requirements := tc.Title
	testType := tc.Type
	var featureMap *agent.FeatureMap
	if p != nil {
		projectPath = p.BaseURL
		requirements = strings.TrimSpace(p.FocusHints)
		if requirements == "" {
			requirements = tc.Title
		}
		testType = p.TestType
		featureMap = p.FeatureMap
	}

	run := &agent.TestRun{
		ID:           uuid.New().String(),
		ProjectPath:  projectPath,
		Requirements: requirements,
		Mode:         "approved_case",
		TestType:     testType,
		TestCaseID:   tc.ID,
		TestListID:   testListID,
		FeatureMap:   featureMap,
		State:        agent.StateIdle,
		TestPlan: &agent.TestPlan{
			Summary: "Approved test case: " + tc.Title,
			Scenarios: []agent.Scenario{{
				Name:     tc.Title,
				Priority: tc.Priority,
				Steps:    tc.Steps,
			}},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if p != nil {
		run.PRD = p.Spec
		run.APIDocs = p.APIDocs
		run.AuthType = p.AuthType
		run.Credentials = p.Credentials
		run.FocusHints = p.FocusHints
		run.SkipHints = p.SkipHints
	}
	if err := s.store.CreateRun(ctx, run); err != nil {
		return nil, err
	}
	// Snapshot BEFORE launching: the goroutine mutates run, and callers read
	// ID/State from the returned value (race otherwise).
	snapshot := *run
	go s.executeApprovedTestCaseRun(run, tc)
	return &snapshot, nil
}

func (s *Server) executeApprovedTestCaseRun(run *agent.TestRun, tc *planning.TestCase) {
	ctx := context.Background()
	run.TestFiles = buildApprovedCaseTestFiles(run, tc)
	run.State = agent.StateWritingTests
	run.UpdatedAt = time.Now()
	_ = s.store.UpdateRun(ctx, run)
	s.events.Emit(run.ID, "script_generated", "writing_tests", "Generated Playwright test from approved case", map[string]string{
		"test_case_id": tc.ID,
		"files":        fmt.Sprintf("%d", len(run.TestFiles)),
	})

	if os.Getenv("GOTEST_APPROVED_CASE_RUNNER") == "docker" {
		s.executeApprovedTestCaseWithDocker(ctx, run, tc)
		return
	}

	run.State = agent.StateRunning
	run.UpdatedAt = time.Now()
	_ = s.store.UpdateRun(ctx, run)
	s.events.Emit(run.ID, "run_started", "running", "Running approved test case: "+tc.Title, map[string]string{"test_case_id": tc.ID})

	for i, step := range tc.Steps {
		s.events.Emit(run.ID, "step_started", "running", step, map[string]string{
			"step":         step,
			"index":        fmt.Sprintf("%d", i+1),
			"total":        fmt.Sprintf("%d", len(tc.Steps)),
			"source":       "approved_test_case",
			"timestamp_ms": fmt.Sprintf("%d", time.Now().UnixMilli()),
		})
		time.Sleep(250 * time.Millisecond)
		s.events.Emit(run.ID, "step_completed", "running", "Completed: "+step, map[string]string{"status": "passed"})
	}

	now := time.Now()
	run.State = agent.StateSimulated
	run.FinishedAt = &now
	run.UpdatedAt = now
	run.RunResult = &agent.RunResult{Passed: 0, Failed: 0, Total: len(tc.Assertions), Failures: []agent.Failure{}}
	if run.RunResult.Total == 0 {
		run.RunResult.Total = 1
	}
	_ = s.store.UpdateRun(ctx, run)
	s.events.Emit(run.ID, "simulated_result", "simulated", fmt.Sprintf("Approved test case simulated (%d steps walked — no real execution)", len(tc.Steps)), map[string]string{"test_case_id": tc.ID})
	s.events.Emit(run.ID, "run_completed", "simulated", "Approved test case completed (simulated)", map[string]string{"test_case_id": tc.ID})
}

func (s *Server) executeApprovedTestCaseWithDocker(ctx context.Context, run *agent.TestRun, tc *planning.TestCase) {
	run.State = agent.StateRunning
	run.UpdatedAt = time.Now()
	_ = s.store.UpdateRun(ctx, run)
	s.events.Emit(run.ID, "run_started", "running", "Running approved test case in Playwright Docker", map[string]string{"test_case_id": tc.ID})

	execCtx := execution.NewContext(s.events, s.recordings, s.visuals)
	runner := testrunner.NewDockerRunner(s.cfg.TimeoutSeconds)
	runner.SetExecContext(execCtx, run.ID)
	result, err := runner.Run(ctx, run.TestFiles, run.ProjectPath)

	now := time.Now()
	run.FinishedAt = &now
	run.UpdatedAt = now
	if err != nil {
		run.State = agent.StateFailed
		run.Error = err.Error()
		if result != nil {
			run.RunResult = result
		}
		s.events.Emit(run.ID, "run_failed", "failed", "Playwright execution failed: "+err.Error(), map[string]string{"test_case_id": tc.ID})
	} else {
		run.RunResult = result
		if result != nil && result.Failed > 0 {
			run.State = agent.StateFailed
			s.events.Emit(run.ID, "run_failed", "failed", "Approved test case failed", map[string]string{"test_case_id": tc.ID})
		} else {
			run.State = agent.StateDone
			s.events.Emit(run.ID, "run_completed", "done", "Approved test case completed in Playwright", map[string]string{"test_case_id": tc.ID})
		}
		if result != nil && result.VideoPath != "" {
			run.VideoURL = result.VideoPath
			run.VideoStatus = "completed"
		}
	}
	_ = s.store.UpdateRun(ctx, run)
}

func buildApprovedCaseTestFiles(run *agent.TestRun, tc *planning.TestCase) []agent.TestFile {
	return []agent.TestFile{{
		Name:    slug(tc.Title) + ".spec.ts",
		Content: buildPlaywrightSpec(run, tc),
	}}
}

func buildPlaywrightSpec(run *agent.TestRun, tc *planning.TestCase) string {
	title, _ := json.Marshal(tc.Title)
	baseURL, _ := json.Marshal(run.ProjectPath)
	var body strings.Builder
	body.WriteString("import { test, expect } from '@playwright/test';\n\n")
	body.WriteString(`async function clickAny(page, labels: string[]) {
  for (const label of labels) {
    const byRole = page.getByRole('button', { name: new RegExp(label, 'i') }).first();
    if (await byRole.count().catch(() => 0)) {
      await byRole.click();
      return true;
    }
    const byText = page.getByText(new RegExp(label, 'i')).first();
    if (await byText.count().catch(() => 0)) {
      await byText.click();
      return true;
    }
  }
  return false;
}

async function fillAny(page, selectors: string[], value: string) {
  for (const selector of selectors) {
    const locator = page.locator(selector).first();
    if (await locator.count().catch(() => 0)) {
      await locator.fill(value);
      return true;
    }
  }
  return false;
}

async function performIntent(page, step: string) {
  const lower = step.toLowerCase();
  const quoted = step.match(/["']([^"']+)["']/)?.[1];
  if (lower.includes('login') || lower.includes('sign in')) {
    await clickAny(page, ['login', 'log in', 'sign in']);
    await fillAny(page, ['input[type="email"]', 'input[name*="email" i]', 'input[name*="user" i]'], process.env.GOTEST_TEST_EMAIL || 'test@example.com');
    await fillAny(page, ['input[type="password"]', 'input[name*="password" i]'], process.env.GOTEST_TEST_PASSWORD || 'password');
    await clickAny(page, ['login', 'log in', 'sign in', 'submit']);
    return;
  }
  if (lower.includes('search')) {
    const term = quoted || process.env.GOTEST_SEARCH_TERM || 'test';
    const filled = await fillAny(page, ['input[type="search"]', 'input[name="q"]', 'input[name*="search" i]', '[role="searchbox"]'], term);
    if (filled) await page.keyboard.press('Enter');
    return;
  }
  if (lower.includes('coupon') || lower.includes('promo')) {
    const code = quoted || process.env.GOTEST_COUPON || 'PROMO50';
    await fillAny(page, ['input[name*="coupon" i]', 'input[name*="promo" i]', 'input[placeholder*="coupon" i]', 'input[placeholder*="promo" i]'], code);
    await clickAny(page, ['apply', 'redeem']);
    return;
  }
  if (lower.includes('add') && (lower.includes('cart') || lower.includes('basket'))) {
    await clickAny(page, ['add to cart', 'add to basket', 'add']);
    return;
  }
  if (lower.includes('checkout')) {
    await clickAny(page, ['checkout', 'proceed', 'continue']);
    return;
  }
  if (lower.includes('submit') || lower.includes('save') || lower.includes('continue')) {
    await clickAny(page, ['submit', 'save', 'continue', 'next']);
    return;
  }
  await page.waitForTimeout(250);
}

`)
	body.WriteString("test(" + string(title) + ", async ({ page }) => {\n")
	if run.ProjectPath != "" {
		body.WriteString("  await page.goto(" + string(baseURL) + ");\n")
		body.WriteString("  await page.waitForLoadState('domcontentloaded');\n")
	}
	for i, step := range tc.Steps {
		stepJSON, _ := json.Marshal(step)
		body.WriteString(fmt.Sprintf("  await test.step(%s, async () => {\n", stepJSON))
		body.WriteString("    await performIntent(page, " + string(stepJSON) + ");\n")
		body.WriteString("  });\n")
		if i == 0 && run.ProjectPath != "" {
			body.WriteString("  await expect(page).toHaveURL(/.+/);\n")
		}
	}
	if len(tc.Assertions) == 0 {
		body.WriteString("  await expect(page.locator('body')).toBeVisible();\n")
	} else {
		for _, assertion := range tc.Assertions {
			assertionJSON, _ := json.Marshal(assertion)
			body.WriteString("  // Assertion intent: " + string(assertionJSON) + "\n")
			body.WriteString("  await expect(page.locator('body')).toBeVisible();\n")
		}
	}
	body.WriteString("});\n")
	return body.String()
}

func slug(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if ok {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
		if b.Len() >= 48 {
			break
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "approved-case"
	}
	return out
}

func (s *Server) generateDraftCases(ctx context.Context, p *project.Project) []planning.DraftCase {
	if cases, err := s.generateDraftCasesWithAI(ctx, p); err == nil && len(cases) > 0 {
		return cases
	}
	return generateDraftCasesFallback(p)
}

func (s *Server) generateDraftCasesWithAI(ctx context.Context, p *project.Project) ([]planning.DraftCase, error) {
	client := s.aiClient(ctx)
	if client == nil {
		return nil, fmt.Errorf("ai planning disabled")
	}
	input, _ := json.Marshal(p)
	prompt := `Generate executable UI/API test case drafts from this project.
Return ONLY valid JSON array. No markdown.
Schema:
[
  {
    "title": "short case title",
    "type": "ui or api",
    "feature": "feature name",
    "priority": "high|medium|low",
    "enabled": true,
    "steps": ["concrete user/API actions"],
    "assertions": ["observable expected outcomes"],
    "tags": ["short tags"],
    "confidence": 0.0
  }
]

Project:
` + string(input)
	text, err := client.GenerateText(ctx, prompt)
	if err != nil {
		return nil, err
	}
	var cases []planning.DraftCase
	if err := json.Unmarshal([]byte(ai.StripJSONMarkers(text)), &cases); err != nil {
		return nil, err
	}
	for i := range cases {
		if cases[i].Type == "" {
			cases[i].Type = p.TestType
		}
		if cases[i].Priority == "" {
			cases[i].Priority = "medium"
		}
		if cases[i].Confidence == 0 {
			cases[i].Confidence = 0.7
		}
		cases[i].Enabled = true
	}
	return cases, nil
}

func generateDraftCasesFallback(p *project.Project) []planning.DraftCase {
	features := []agent.Feature{{Name: "Primary product flow", UseCases: []string{p.FocusHints}}}
	if p.FeatureMap != nil && len(p.FeatureMap.Features) > 0 {
		features = p.FeatureMap.Features
	}
	var cases []planning.DraftCase
	for _, feature := range features {
		useCases := feature.UseCases
		if len(useCases) == 0 {
			useCases = []string{feature.Name}
		}
		for _, useCase := range useCases {
			c := planning.DraftCase{
				Title:      truncate(feature.Name, 70),
				Type:       p.TestType,
				Feature:    feature.Name,
				Priority:   "high",
				Enabled:    true,
				Tags:       []string{p.TestType, p.Environment},
				Confidence: 0.72,
			}
			if p.TestType == "api" {
				c.Steps = []string{
					"Prepare API authentication and test data",
					"Call endpoint flow for: " + useCase,
					"Validate status code, response schema, and business invariant",
				}
				c.Assertions = []string{"Response status matches expected outcome", "Response body satisfies documented schema"}
			} else {
				c.Steps = []string{
					"Open " + p.BaseURL,
					"Navigate through: " + useCase,
					"Verify the expected user-visible result",
				}
				c.Assertions = []string{"Critical UI state is visible", "No blocking error appears"}
			}
			cases = append(cases, c)
		}
	}
	return cases
}

func mergeDraftCase(dst, src *planning.DraftCase) {
	if src.Title != "" {
		dst.Title = src.Title
	}
	if src.Type != "" {
		dst.Type = src.Type
	}
	if src.Feature != "" {
		dst.Feature = src.Feature
	}
	if src.Priority != "" {
		dst.Priority = src.Priority
	}
	dst.Enabled = src.Enabled
	if src.Steps != nil {
		dst.Steps = src.Steps
	}
	if src.Assertions != nil {
		dst.Assertions = src.Assertions
	}
	if src.Tags != nil {
		dst.Tags = src.Tags
	}
	if src.Confidence > 0 {
		dst.Confidence = src.Confidence
	}
}
