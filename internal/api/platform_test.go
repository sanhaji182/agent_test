package api_test

import (
	"encoding/json"
	"testing"
)

func TestScheduleCRUD(t *testing.T) {
	srv := newTestServer()

	// Create
	w := post(srv, "/api/v1/schedules", `{"name":"nightly","project_path":"/tmp/app","frequency":"daily","enabled":true,"environment":"staging","base_url":"http://staging.example.com"}`)
	assertStatus(t, w, 201)
	var sch map[string]interface{}
	json.NewDecoder(w.Body).Decode(&sch)
	id := sch["id"].(string)
	if id == "" {
		t.Fatal("expected schedule id")
	}
	if sch["frequency"] != "daily" {
		t.Fatalf("expected daily, got %v", sch["frequency"])
	}

	// List
	w = get(srv, "/api/v1/schedules")
	assertStatus(t, w, 200)
	var list []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&list)
	if len(list) != 1 {
		t.Fatalf("expected 1 schedule, got %d", len(list))
	}

	// Get
	w = get(srv, "/api/v1/schedules/"+id)
	assertStatus(t, w, 200)

	// Update
	w = patch(srv, "/api/v1/schedules/"+id, `{"enabled":false}`)
	assertStatus(t, w, 200)
	var updated map[string]interface{}
	json.NewDecoder(w.Body).Decode(&updated)
	if updated["enabled"] != false {
		t.Fatal("expected enabled=false")
	}

	// Run now
	w = postNoBody(srv, "/api/v1/schedules/"+id+"/run-now")
	assertStatus(t, w, 202)
	var runResp map[string]string
	json.NewDecoder(w.Body).Decode(&runResp)
	if runResp["run_id"] == "" {
		t.Fatal("expected run_id from run-now")
	}

	// Delete
	w = del(srv, "/api/v1/schedules/"+id)
	assertStatus(t, w, 204)
}

func TestReleaseCRUD(t *testing.T) {
	srv := newTestServer()

	w := post(srv, "/api/v1/releases", `{"name":"v1.0","version":"1.0.0","project_id":"proj-1"}`)
	assertStatus(t, w, 201)
	var rel map[string]interface{}
	json.NewDecoder(w.Body).Decode(&rel)
	id := rel["id"].(string)

	w = get(srv, "/api/v1/releases")
	assertStatus(t, w, 200)

	w = get(srv, "/api/v1/releases/"+id)
	assertStatus(t, w, 200)

	w = patch(srv, "/api/v1/releases/"+id, `{"status":"completed"}`)
	assertStatus(t, w, 200)

	w = get(srv, "/api/v1/releases/"+id+"/summary")
	assertStatus(t, w, 200)
}

func TestNotifications(t *testing.T) {
	srv := newTestServer()
	w := get(srv, "/api/v1/notifications")
	assertStatus(t, w, 200)
}

func TestMetrics(t *testing.T) {
	srv := newTestServer()

	w := get(srv, "/api/v1/metrics/summary")
	assertStatus(t, w, 200)

	w = get(srv, "/api/v1/metrics/hotspots")
	assertStatus(t, w, 200)

	w = get(srv, "/api/v1/metrics/flaky")
	assertStatus(t, w, 200)

	w = get(srv, "/api/v1/metrics/trend")
	assertStatus(t, w, 200)
}
