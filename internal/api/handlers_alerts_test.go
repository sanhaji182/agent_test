package api_test

import (
	"encoding/json"
	"testing"
)

// TestAlertsEndpoint memverifikasi endpoint /api/v1/alerts beserta action
// acknowledge/dismiss/read-all. Demo seed dipakai untuk mengisi data alert.
func TestAlertsEndpoint(t *testing.T) {
	srv := newTestServer()

	// Seed demo data → menciptakan notifications (failure, flake, degradation).
	w := postNoBody(srv, "/api/v1/demo/seed")
	assertStatus(t, w, 200)

	// List alerts: harus ada dan punya severity/category server-side.
	w = get(srv, "/api/v1/alerts")
	assertStatus(t, w, 200)
	var alerts []map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&alerts); err != nil {
		t.Fatalf("decode alerts: %v", err)
	}
	if len(alerts) == 0 {
		t.Fatal("expected seeded alerts")
	}
	seenSeverity := false
	for _, a := range alerts {
		if a["severity"] == nil || a["severity"] == "" {
			t.Fatalf("alert missing severity: %v", a)
		}
		if a["category"] == nil || a["category"] == "" {
			t.Fatalf("alert missing category: %v", a)
		}
		if a["severity"] == "critical" {
			seenSeverity = true
		}
	}
	if !seenSeverity {
		t.Fatal("expected at least one critical (failure) alert")
	}

	// Filter by type.
	w = get(srv, "/api/v1/alerts?type=failure")
	assertStatus(t, w, 200)
	var failures []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&failures)
	if len(failures) == 0 {
		t.Fatal("expected failure alerts")
	}
	for _, a := range failures {
		if a["type"] != "failure" {
			t.Fatalf("expected failure type, got %v", a["type"])
		}
	}

	// Acknowledge alert pertama (idempotent — item yang sama tetap 200).
	id := alerts[0]["id"].(string)
	w = postNoBody(srv, "/api/v1/alerts/"+id+"/acknowledge")
	assertStatus(t, w, 200)
	w = postNoBody(srv, "/api/v1/alerts/"+id+"/acknowledge")
	assertStatus(t, w, 200)

	// Acknowledge alert yang tidak ada → 404.
	w = postNoBody(srv, "/api/v1/alerts/missing/acknowledge")
	assertStatus(t, w, 404)

	// Dismiss alert kedua.
	id2 := alerts[1]["id"].(string)
	w = postNoBody(srv, "/api/v1/alerts/"+id2+"/dismiss")
	assertStatus(t, w, 200)

	// Dismissed alerts disembunyikan secara default.
	w = get(srv, "/api/v1/alerts")
	assertStatus(t, w, 200)
	var active []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&active)
	if len(active) != len(alerts)-1 {
		t.Fatalf("expected %d active alerts, got %d", len(alerts)-1, len(active))
	}

	// include_dismissed=true menampilkan kembali alert yang di-dismiss.
	w = get(srv, "/api/v1/alerts?include_dismissed=true")
	assertStatus(t, w, 200)
	var withDismissed []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&withDismissed)
	if len(withDismissed) != len(alerts) {
		t.Fatalf("expected %d alerts with dismissed, got %d", len(alerts), len(withDismissed))
	}

	// Mark all read → tidak ada lagi yang unacknowledged.
	w = postNoBody(srv, "/api/v1/alerts/read-all")
	assertStatus(t, w, 200)
	w = get(srv, "/api/v1/alerts")
	assertStatus(t, w, 200)
	var all []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&all)
	for _, a := range all {
		if a["acknowledged"] != true {
			t.Fatalf("expected all acknowledged after read-all, got %v", a["acknowledged"])
		}
	}
}
