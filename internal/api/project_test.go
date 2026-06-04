package api_test

import (
	"encoding/json"
	"testing"
)

func TestProjectIntakeLifecycle(t *testing.T) {
	srv := newTestServer()

	w := post(srv, "/api/v1/projects", `{
		"name":"Customer Portal",
		"test_type":"ui",
		"base_url":"https://app.example.com",
		"spec":"Feature: Login\nUse case: Returning users can sign in\nFeature: Billing\nUse case: Admins can update plans",
		"auth_type":"login",
		"credentials":"use secret reference",
		"focus_hints":"login and billing",
		"skip_hints":"do not charge cards"
	}`)
	assertStatus(t, w, 201)

	var project map[string]interface{}
	json.NewDecoder(w.Body).Decode(&project)
	id := project["id"].(string)
	if id == "" {
		t.Fatal("expected project id")
	}
	if project["feature_map"] == nil {
		t.Fatal("expected feature_map")
	}

	w = get(srv, "/api/v1/projects")
	assertStatus(t, w, 200)
	var list []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&list)
	if len(list) != 1 {
		t.Fatalf("expected 1 project, got %d", len(list))
	}

	w = get(srv, "/api/v1/projects/"+id)
	assertStatus(t, w, 200)

	w = patch(srv, "/api/v1/projects/"+id, `{"focus_hints":"billing only"}`)
	assertStatus(t, w, 200)

	w = postNoBody(srv, "/api/v1/projects/"+id+"/extract-features")
	assertStatus(t, w, 200)
	var featureMap map[string]interface{}
	json.NewDecoder(w.Body).Decode(&featureMap)
	features := featureMap["features"].([]interface{})
	if len(features) == 0 {
		t.Fatal("expected extracted features")
	}
}

func TestCreateProjectRequiresName(t *testing.T) {
	srv := newTestServer()
	w := post(srv, "/api/v1/projects", `{"base_url":"https://app.example.com"}`)
	assertStatus(t, w, 400)
}
