package javascript_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/parser/javascript"
)

func TestJavaScriptParser_ParseExpressRoutes(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "js-parser-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create package.json
	packageJSON := `{
		"name": "test-express-app",
		"dependencies": {
			"express": "^4.18.0"
		}
	}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSON), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	// Create Express app file
	expressApp := `
const express = require('express');
const app = express();

// Basic routes
app.get('/users', listUsers);
app.post('/users', authenticate, createUser);
app.get('/users/:id', getUser);
app.put('/users/:id', authenticate, updateUser);
app.delete('/users/:id', authenticate, authorize, deleteUser);

// Middleware
app.use('/api', apiRouter);

// Route with multiple middleware
app.post('/auth/login', validateInput, rateLimit, loginHandler);
`
	if err := os.WriteFile(filepath.Join(tmpDir, "app.js"), []byte(expressApp), 0644); err != nil {
		t.Fatalf("Failed to create app.js: %v", err)
	}

	// Parse codebase
	parser := javascript.NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify framework detection
	if codebase.Framework != "express" {
		t.Errorf("Expected framework express, got %s", codebase.Framework)
	}

	// Verify routes
	expectedRoutes := []struct {
		method     string
		path       string
		middleware []string
	}{
		{"GET", "/users", []string{"listUsers"}},
		{"POST", "/users", []string{"authenticate", "createUser"}},
		{"GET", "/users/:id", []string{"getUser"}},
		{"PUT", "/users/:id", []string{"authenticate", "updateUser"}},
		{"DELETE", "/users/:id", []string{"authenticate", "authorize", "deleteUser"}},
		{"USE", "/api", []string{"apiRouter"}},
		{"POST", "/auth/login", []string{"validateInput", "rateLimit", "loginHandler"}},
	}

	if len(codebase.Routes) != len(expectedRoutes) {
		t.Errorf("Expected %d routes, got %d", len(expectedRoutes), len(codebase.Routes))
	}

	for i, expected := range expectedRoutes {
		if i >= len(codebase.Routes) {
			break
		}

		route := codebase.Routes[i]

		if route.Method != expected.method {
			t.Errorf("Route %d: expected method %s, got %s", i, expected.method, route.Method)
		}

		if route.Path != expected.path {
			t.Errorf("Route %d: expected path %s, got %s", i, expected.path, route.Path)
		}

		if len(route.Middleware) != len(expected.middleware) {
			t.Errorf("Route %d: expected %d middleware, got %d", i, len(expected.middleware), len(route.Middleware))
			continue
		}

		for j, mw := range expected.middleware {
			if route.Middleware[j] != mw {
				t.Errorf("Route %d middleware %d: expected %s, got %s", i, j, mw, route.Middleware[j])
			}
		}
	}
}

func TestJavaScriptParser_ParseRouterPattern(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "js-parser-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create package.json
	packageJSON := `{"dependencies": {"express": "^4.18.0"}}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSON), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	// Create router file
	routerCode := `
const express = require('express');
const router = express.Router();

router.get('/', listItems);
router.post('/', createItem);
router.get('/:id', getItem);
router.put('/:id', updateItem);
router.delete('/:id', deleteItem);

module.exports = router;
`
	if err := os.WriteFile(filepath.Join(tmpDir, "routes.js"), []byte(routerCode), 0644); err != nil {
		t.Fatalf("Failed to create routes.js: %v", err)
	}

	// Parse codebase
	parser := javascript.NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify router routes
	if len(codebase.Routes) < 5 {
		t.Errorf("Expected at least 5 routes, got %d", len(codebase.Routes))
	}

	// Check for GET / route
	found := false
	for _, route := range codebase.Routes {
		if route.Method == "GET" && route.Path == "/" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find GET / route")
	}
}

func TestJavaScriptParser_ParseTypeScript(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "js-parser-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create package.json
	packageJSON := `{"dependencies": {"express": "^4.18.0", "typescript": "^5.0.0"}}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSON), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	// Create TypeScript file
	tsCode := `
import express from 'express';
const app = express();

app.get('/api/users', (req, res) => {
  res.json({ users: [] });
});

app.post('/api/users', async (req, res) => {
  const user = await createUser(req.body);
  res.json(user);
});
`
	if err := os.WriteFile(filepath.Join(tmpDir, "app.ts"), []byte(tsCode), 0644); err != nil {
		t.Fatalf("Failed to create app.ts: %v", err)
	}

	// Parse codebase
	parser := javascript.NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify TypeScript routes
	if len(codebase.Routes) < 2 {
		t.Errorf("Expected at least 2 routes, got %d", len(codebase.Routes))
	}
}

func TestJavaScriptParser_ParseNextJS(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "js-parser-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create package.json
	packageJSON := `{"dependencies": {"next": "^14.0.0", "react": "^18.0.0"}}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSON), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	// Parse codebase
	parser := javascript.NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify framework detection
	if codebase.Framework != "nextjs" {
		t.Errorf("Expected framework nextjs, got %s", codebase.Framework)
	}
}

func TestJavaScriptParser_ParseReact(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "js-parser-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create package.json
	packageJSON := `{"dependencies": {"react": "^18.0.0", "react-dom": "^18.0.0"}}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSON), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	// Parse codebase
	parser := javascript.NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify framework detection
	if codebase.Framework != "react" {
		t.Errorf("Expected framework react, got %s", codebase.Framework)
	}
}

func TestJavaScriptParser_SkipNodeModules(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "js-parser-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create package.json
	packageJSON := `{"dependencies": {"express": "^4.18.0"}}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSON), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	// Create app.js
	appCode := `app.get('/test', handler);`
	if err := os.WriteFile(filepath.Join(tmpDir, "app.js"), []byte(appCode), 0644); err != nil {
		t.Fatalf("Failed to create app.js: %v", err)
	}

	// Create node_modules directory with files
	nodeModulesDir := filepath.Join(tmpDir, "node_modules", "express")
	if err := os.MkdirAll(nodeModulesDir, 0755); err != nil {
		t.Fatalf("Failed to create node_modules: %v", err)
	}

	// Create file in node_modules (should be skipped)
	expressCode := `app.get('/should-not-be-parsed', handler);`
	if err := os.WriteFile(filepath.Join(nodeModulesDir, "index.js"), []byte(expressCode), 0644); err != nil {
		t.Fatalf("Failed to create express index.js: %v", err)
	}

	// Parse codebase
	parser := javascript.NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify only app.js route was parsed
	if len(codebase.Routes) != 1 {
		t.Errorf("Expected 1 route (node_modules should be skipped), got %d", len(codebase.Routes))
	}

	if codebase.Routes[0].Path != "/test" {
		t.Errorf("Expected path /test, got %s", codebase.Routes[0].Path)
	}
}

func TestJavaScriptParser_EmptyProject(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "js-parser-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create package.json only
	packageJSON := `{"name": "empty-project"}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSON), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	// Parse codebase
	parser := javascript.NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify empty results
	if len(codebase.Routes) != 0 {
		t.Errorf("Expected 0 routes, got %d", len(codebase.Routes))
	}

	if len(codebase.Models) != 0 {
		t.Errorf("Expected 0 models, got %d", len(codebase.Models))
	}

	if len(codebase.Handlers) != 0 {
		t.Errorf("Expected 0 handlers, got %d", len(codebase.Handlers))
	}
}

func TestJavaScriptParser_MultipleFiles(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "js-parser-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create package.json
	packageJSON := `{"dependencies": {"express": "^4.18.0"}}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSON), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	// Create multiple route files
	files := map[string]string{
		"routes/users.js": `app.get('/users', listUsers);`,
		"routes/posts.js": `app.get('/posts', listPosts);`,
		"routes/comments.js": `app.get('/comments', listComments);`,
	}

	for filename, content := range files {
		filePath := filepath.Join(tmpDir, filename)
		// Create directory if needed
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create %s: %v", filename, err)
		}
	}

	// Parse codebase
	parser := javascript.NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify all routes from all files
	if len(codebase.Routes) < 3 {
		t.Errorf("Expected at least 3 routes from 3 files, got %d", len(codebase.Routes))
	}
}
