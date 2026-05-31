package api_test

import (
	"encoding/json"
	"testing"
)

func TestMetricsRisk(t *testing.T) {
	srv := newTestServer()
	w := get(srv, "/api/v1/metrics/risk")
	assertStatus(t, w, 200)
}

func TestMetricsRecommendations(t *testing.T) {
	srv := newTestServer()
	w := get(srv, "/api/v1/metrics/recommendations")
	assertStatus(t, w, 200)
}

func TestSuiteSelection(t *testing.T) {
	srv := newTestServer()
	w := post(srv, "/api/v1/suite-selection", `{"mode":"all","all_tests":["login","checkout"]}`)
	assertStatus(t, w, 200)
	var sel map[string]interface{}
	json.NewDecoder(w.Body).Decode(&sel)
	if sel["mode"] != "all" {
		t.Fatalf("expected mode=all, got %v", sel["mode"])
	}
}

func TestReviewWorkflow(t *testing.T) {
	srv := newTestServer()

	// Create run first
	w := post(srv, "/api/v1/runs", `{"project_path":"/tmp/x"}`)
	var cr map[string]string
	json.NewDecoder(w.Body).Decode(&cr)
	runID := cr["run_id"]

	// Create review
	w = post(srv, "/api/v1/reviews", `{"run_id":"`+runID+`","type":"test_plan"}`)
	assertStatus(t, w, 201)
	var rev map[string]interface{}
	json.NewDecoder(w.Body).Decode(&rev)
	revID := rev["id"].(string)
	if rev["status"] != "pending" {
		t.Fatalf("expected pending, got %v", rev["status"])
	}

	// Get reviews for run
	w = get(srv, "/api/v1/runs/"+runID+"/reviews")
	assertStatus(t, w, 200)

	// Approve
	w = post(srv, "/api/v1/reviews/"+revID+"/approve", `{"reviewer":"alice","comment":"LGTM"}`)
	assertStatus(t, w, 200)
	json.NewDecoder(w.Body).Decode(&rev)
	if rev["status"] != "approved" {
		t.Fatalf("expected approved, got %v", rev["status"])
	}
}

func TestSuiteCRUD(t *testing.T) {
	srv := newTestServer()

	w := post(srv, "/api/v1/suites", `{"name":"smoke","tags":["critical","fast"],"pinned":true}`)
	assertStatus(t, w, 201)
	var suite map[string]interface{}
	json.NewDecoder(w.Body).Decode(&suite)
	id := suite["id"].(string)

	w = get(srv, "/api/v1/suites")
	assertStatus(t, w, 200)
	var list []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&list)
	if len(list) != 1 {
		t.Fatalf("expected 1 suite, got %d", len(list))
	}

	w = get(srv, "/api/v1/suites/"+id)
	assertStatus(t, w, 200)

	w = del(srv, "/api/v1/suites/"+id)
	assertStatus(t, w, 204)
}

func TestAlertRulesEvaluate(t *testing.T) {
	srv := newTestServer()
	w := post(srv, "/api/v1/alert-rules/evaluate", `{"rules":[{"id":"1","name":"low pass","condition":"pass_rate_drop","threshold":0.9,"enabled":true}]}`)
	assertStatus(t, w, 200)
}
