package rust

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/parser/types"
)

func TestDetectFrameworkFromCargoToml(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Cargo.toml", `[package]
name = "sample"
version = "0.1.0"

[dependencies]
actix-web = "4"
serde = { version = "1", features = ["derive"] }
`)

	p := NewParser()
	framework, err := p.DetectFramework(root)
	if err != nil {
		t.Fatalf("DetectFramework returned error: %v", err)
	}
	if framework != "actix-web" {
		t.Fatalf("expected actix-web, got %q", framework)
	}
}

func TestParseActixMacroRoutesModelsAndHandlers(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Cargo.toml", `[package]
name = "sample"
version = "0.1.0"

[dependencies]
actix-web = "4"
serde = { version = "1", features = ["derive"] }
`)
	writeFile(t, root, "src/main.rs", `use actix_web::{get, post, web, HttpResponse};
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct User {
    pub id: uuid::Uuid,
    pub email: String,
    pub display_name: Option<String>,
}

#[get("/users/{id}")]
async fn get_user(path: web::Path<uuid::Uuid>) -> HttpResponse {
    let user = db.find(path.into_inner()).await;
    HttpResponse::Ok().json(user)
}

#[post("/users")]
async fn create_user(payload: web::Json<User>) -> HttpResponse {
    payload.validate().unwrap();
    db.insert(payload.into_inner()).await;
    HttpResponse::Created().finish()
}
`)

	codebase, err := NewParser().Parse(context.Background(), root)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if codebase.Language != "rust" || codebase.Framework != "actix-web" {
		t.Fatalf("unexpected codebase metadata: %+v", codebase)
	}
	if codebase.FileCount != 1 {
		t.Fatalf("expected 1 Rust file, got %d", codebase.FileCount)
	}
	if len(codebase.Routes) != 2 {
		t.Fatalf("expected 2 routes, got %+v", codebase.Routes)
	}
	assertRoute(t, codebase.Routes, "GET", "/users/{id}", "get_user")
	assertRoute(t, codebase.Routes, "POST", "/users", "create_user")
	if got := codebase.Routes[0].Params["id"]; got != "string" {
		t.Fatalf("expected id path param, got %+v", codebase.Routes[0].Params)
	}

	if len(codebase.Models) != 1 {
		t.Fatalf("expected 1 model, got %+v", codebase.Models)
	}
	model := codebase.Models[0]
	if model.Name != "User" || model.Table != "users" || len(model.Fields) != 3 {
		t.Fatalf("unexpected model: %+v", model)
	}
	if model.Fields[2].Name != "display_name" || model.Fields[2].Required {
		t.Fatalf("Option<T> field should not be required: %+v", model.Fields[2])
	}

	if len(codebase.Handlers) != 2 {
		t.Fatalf("expected 2 handlers, got %+v", codebase.Handlers)
	}
	create := findHandler(codebase.Handlers, "create_user")
	if create == nil {
		t.Fatal("missing create_user handler")
	}
	if !create.HasValidation {
		t.Fatalf("expected validation detection on create_user: %+v", create)
	}
	if len(create.DatabaseOps) == 0 || create.DatabaseOps[0] != "create" {
		t.Fatalf("expected create database op, got %+v", create.DatabaseOps)
	}
}

func TestParseAxumRoutes(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "Cargo.toml", `[package]
name = "sample"
version = "0.1.0"

[dependencies]
axum = "0.7"
`)
	writeFile(t, root, "src/main.rs", `use axum::{routing::{get, post}, Router};

async fn list_users() -> &'static str { "ok" }
async fn create_user() -> &'static str { "created" }
async fn get_user() -> &'static str { "ok" }
async fn update_user() -> &'static str { "ok" }

pub fn app() -> Router {
    Router::new()
        .route("/users", get(list_users).post(create_user))
        .route("/users/:id", get(get_user).put(update_user))
}
`)

	codebase, err := NewParser().Parse(context.Background(), root)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if codebase.Framework != "axum" {
		t.Fatalf("expected axum framework, got %q", codebase.Framework)
	}
	if len(codebase.Routes) != 4 {
		t.Fatalf("expected 4 axum routes, got %+v", codebase.Routes)
	}
	assertRoute(t, codebase.Routes, "GET", "/users", "list_users")
	assertRoute(t, codebase.Routes, "POST", "/users", "create_user")
	assertRoute(t, codebase.Routes, "GET", "/users/:id", "get_user")
	assertRoute(t, codebase.Routes, "PUT", "/users/:id", "update_user")
}

func writeFile(t *testing.T, root, relativePath, content string) {
	t.Helper()
	path := filepath.Join(root, relativePath)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create dir for %s: %v", relativePath, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write %s: %v", relativePath, err)
	}
}

func findHandler(handlers []types.Handler, name string) *types.Handler {
	for i := range handlers {
		if handlers[i].Name == name {
			return &handlers[i]
		}
	}
	return nil
}

func assertRoute(t *testing.T, routes []types.Route, method, path, handler string) {
	t.Helper()
	for _, route := range routes {
		if route.Method == method && route.Path == path && route.Handler == handler {
			return
		}
	}
	t.Fatalf("missing route %s %s -> %s in %+v", method, path, handler, routes)
}
