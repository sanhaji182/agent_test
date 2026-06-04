package api_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/schedule"
)

func TestGenerateApprovePlanLifecycle(t *testing.T) {
	srv := newTestServer()

	w := post(srv, "/api/v1/projects", `{
		"name":"Portal",
		"test_type":"ui",
		"base_url":"https://app.example.com",
		"spec":"Feature: Login\nUse case: Returning users can sign in"
	}`)
	assertStatus(t, w, 201)
	var proj map[string]interface{}
	json.NewDecoder(w.Body).Decode(&proj)
	projectID := proj["id"].(string)

	w = postNoBody(srv, "/api/v1/projects/"+projectID+"/test-plan")
	assertStatus(t, w, 201)
	var plan map[string]interface{}
	json.NewDecoder(w.Body).Decode(&plan)
	planID := plan["id"].(string)
	cases := plan["cases"].([]interface{})
	if len(cases) == 0 {
		t.Fatal("expected draft cases")
	}

	firstCase := cases[0].(map[string]interface{})
	caseID := firstCase["id"].(string)
	w = patch(srv, "/api/v1/test-plans/"+planID+"/cases/"+caseID, `{"title":"Edited login flow","enabled":true}`)
	assertStatus(t, w, 200)

	w = postNoBody(srv, "/api/v1/test-plans/"+planID+"/approve")
	assertStatus(t, w, 200)

	w = get(srv, "/api/v1/test-cases?project_id="+projectID)
	assertStatus(t, w, 200)
	var testCases []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&testCases)
	if len(testCases) == 0 {
		t.Fatal("expected approved test cases")
	}
	foundEdited := false
	for _, tc := range testCases {
		if tc["title"] == "Edited login flow" {
			foundEdited = true
			break
		}
	}
	if !foundEdited {
		t.Fatalf("expected edited title in approved cases, got %v", testCases)
	}

	runCaseID := testCases[0]["id"].(string)
	w = get(srv, "/api/v1/test-cases/maintenance?project_id="+projectID)
	assertStatus(t, w, 200)
	var maintenance []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&maintenance)
	if len(maintenance) == 0 {
		t.Fatal("expected maintenance recommendations")
	}

	w = postNoBody(srv, "/api/v1/test-cases/"+runCaseID+"/run")
	assertStatus(t, w, 202)
	var runResp map[string]string
	json.NewDecoder(w.Body).Decode(&runResp)
	if runResp["run_id"] == "" {
		t.Fatal("expected run_id")
	}

	time.Sleep(1200 * time.Millisecond)
	w = get(srv, "/api/v1/runs/"+runResp["run_id"])
	assertStatus(t, w, 200)
	var run map[string]interface{}
	json.NewDecoder(w.Body).Decode(&run)
	if run["test_plan"] == nil {
		t.Fatal("expected run test_plan")
	}
	if run["test_case_id"] != runCaseID {
		t.Fatalf("expected test_case_id %s, got %v", runCaseID, run["test_case_id"])
	}
	files, ok := run["test_files"].([]interface{})
	if !ok || len(files) == 0 {
		t.Fatalf("expected generated test_files, got %v", run["test_files"])
	}
	firstFile := files[0].(map[string]interface{})
	content := firstFile["content"].(string)
	if !strings.Contains(content, "performIntent") {
		t.Fatal("expected generated Playwright intent helper")
	}

	w = post(srv, "/api/v1/test-cases/"+runCaseID+"/refine", `{"prompt":"cover invalid password error messaging"}`)
	assertStatus(t, w, 201)
	var proposal map[string]interface{}
	json.NewDecoder(w.Body).Decode(&proposal)
	proposalID := proposal["id"].(string)
	if proposal["status"] != "pending" {
		t.Fatalf("expected pending proposal, got %v", proposal["status"])
	}

	w = get(srv, "/api/v1/test-cases/maintenance?project_id="+projectID)
	assertStatus(t, w, 200)
	json.NewDecoder(w.Body).Decode(&maintenance)
	foundPendingProposal := false
	for _, item := range maintenance {
		if item["category"] == "pending_proposal" {
			foundPendingProposal = true
			break
		}
	}
	if !foundPendingProposal {
		t.Fatalf("expected pending proposal maintenance item, got %v", maintenance)
	}

	w = get(srv, "/api/v1/test-cases/"+runCaseID+"/proposals")
	assertStatus(t, w, 200)
	var proposals []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&proposals)
	if len(proposals) != 1 {
		t.Fatalf("expected 1 proposal, got %d", len(proposals))
	}

	w = get(srv, "/api/v1/change-proposals")
	assertStatus(t, w, 200)
	var allProposals []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&allProposals)
	if len(allProposals) != 1 {
		t.Fatalf("expected 1 global proposal, got %d", len(allProposals))
	}

	w = post(srv, "/api/v1/test-cases/"+runCaseID+"/refine", `{"prompt":"reject this alternate proposal"}`)
	assertStatus(t, w, 201)
	var rejectable map[string]interface{}
	json.NewDecoder(w.Body).Decode(&rejectable)
	rejectableID := rejectable["id"].(string)
	w = post(srv, "/api/v1/change-proposals/"+rejectableID+"/reject", `{"reviewer":"qa","comment":"not needed"}`)
	assertStatus(t, w, 200)
	var rejected map[string]interface{}
	json.NewDecoder(w.Body).Decode(&rejected)
	if rejected["status"] != "rejected" {
		t.Fatalf("expected rejected proposal, got %v", rejected["status"])
	}

	w = post(srv, "/api/v1/change-proposals/"+proposalID+"/approve", `{"reviewer":"qa","comment":"looks good"}`)
	assertStatus(t, w, 200)
	var approved map[string]interface{}
	json.NewDecoder(w.Body).Decode(&approved)
	updatedCase := approved["test_case"].(map[string]interface{})
	if updatedCase["version"].(float64) <= testCases[0]["version"].(float64) {
		t.Fatalf("expected test case version bump, got %v", updatedCase["version"])
	}

	w = post(srv, "/api/v1/test-lists", `{"name":"Smoke","project_id":"`+projectID+`","test_case_ids":["`+runCaseID+`"],"tags":["smoke"],"pinned":true}`)
	assertStatus(t, w, 201)
	var list map[string]interface{}
	json.NewDecoder(w.Body).Decode(&list)
	listID := list["id"].(string)

	w = get(srv, "/api/v1/test-lists?project_id="+projectID)
	assertStatus(t, w, 200)
	var lists []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&lists)
	if len(lists) != 1 {
		t.Fatalf("expected 1 test list, got %d", len(lists))
	}

	w = postNoBody(srv, "/api/v1/test-lists/"+listID+"/run")
	assertStatus(t, w, 202)
	var listRun map[string]interface{}
	json.NewDecoder(w.Body).Decode(&listRun)
	runIDs := listRun["run_ids"].([]interface{})
	if len(runIDs) != 1 {
		t.Fatalf("expected 1 list run, got %d", len(runIDs))
	}
	time.Sleep(1200 * time.Millisecond)
	w = get(srv, "/api/v1/runs/"+runIDs[0].(string))
	assertStatus(t, w, 200)
	var linkedRun map[string]interface{}
	json.NewDecoder(w.Body).Decode(&linkedRun)
	if linkedRun["test_list_id"] != listID {
		t.Fatalf("expected test_list_id %s, got %v", listID, linkedRun["test_list_id"])
	}

	w = get(srv, "/api/v1/test-lists/"+listID+"/history")
	assertStatus(t, w, 200)
	var history map[string]interface{}
	json.NewDecoder(w.Body).Decode(&history)
	if history["test_list_id"] != listID {
		t.Fatalf("expected history test_list_id %s, got %v", listID, history["test_list_id"])
	}
	if history["latest"] == nil || history["counts"] == nil || history["runs"] == nil {
		t.Fatalf("expected history latest/counts/runs, got %v", history)
	}
	if _, ok := history["newly_failed"].([]interface{}); !ok {
		t.Fatalf("expected newly_failed list, got %v", history["newly_failed"])
	}
	if _, ok := history["recovered"].([]interface{}); !ok {
		t.Fatalf("expected recovered list, got %v", history["recovered"])
	}
	if _, ok := history["stable_failed"].([]interface{}); !ok {
		t.Fatalf("expected stable_failed list, got %v", history["stable_failed"])
	}

	w = post(srv, "/api/v1/schedules", `{"name":"Nightly Smoke","test_list_id":"`+listID+`","frequency":"daily","enabled":true}`)
	assertStatus(t, w, 201)
	var sch map[string]interface{}
	json.NewDecoder(w.Body).Decode(&sch)
	scheduleID := sch["id"].(string)
	if sch["test_list_id"] != listID {
		t.Fatalf("expected schedule test_list_id %s, got %v", listID, sch["test_list_id"])
	}

	w = postNoBody(srv, "/api/v1/schedules/"+scheduleID+"/run-now")
	assertStatus(t, w, 202)
	var scheduleRun map[string]interface{}
	json.NewDecoder(w.Body).Decode(&scheduleRun)
	scheduleRunIDs := scheduleRun["run_ids"].([]interface{})
	if len(scheduleRunIDs) != 1 || scheduleRun["run_id"] == "" {
		t.Fatalf("expected schedule run ids, got %v", scheduleRun)
	}
	dueAt := time.Now().Add(-time.Minute)
	srv.Schedules().Update(scheduleID, func(sc *schedule.Schedule) {
		sc.Enabled = true
		sc.NextRunAt = dueAt
	})
	if processed := srv.ProcessDueSchedules(t.Context(), time.Now()); processed == 0 {
		t.Fatal("expected due test list schedule to be processed")
	}

	w = get(srv, "/api/v1/monitoring/summary")
	assertStatus(t, w, 200)
	var mon map[string]interface{}
	json.NewDecoder(w.Body).Decode(&mon)
	if mon["summary"] == nil || mon["lists"] == nil || mon["recent_runs"] == nil {
		t.Fatalf("expected monitoring summary, got %v", mon)
	}
}
