package golang

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestGoParser_DetectFramework(t *testing.T) {
	tests := []struct {
		name     string
		goMod    string
		expected string
	}{
		{
			name: "Chi framework",
			goMod: `module example.com/myapp

go 1.21

require (
	github.com/go-chi/chi/v5 v5.0.0
)
`,
			expected: "chi",
		},
		{
			name: "Gin framework",
			goMod: `module example.com/myapp

go 1.21

require (
	github.com/gin-gonic/gin v1.9.0
)
`,
			expected: "gin",
		},
		{
			name: "Echo framework",
			goMod: `module example.com/myapp

go 1.21

require (
	github.com/labstack/echo/v4 v4.10.0
)
`,
			expected: "echo",
		},
		{
			name: "Fiber framework",
			goMod: `module example.com/myapp

go 1.21

require (
	github.com/gofiber/fiber/v2 v2.40.0
)
`,
			expected: "fiber",
		},
		{
			name: "Gorilla Mux",
			goMod: `module example.com/myapp

go 1.21

require (
	github.com/gorilla/mux v1.8.0
)
`,
			expected: "gorilla",
		},
		{
			name: "Standard library",
			goMod: `module example.com/myapp

go 1.21
`,
			expected: "stdlib",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir := t.TempDir()

			// Write go.mod
			err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(tt.goMod), 0644)
			if err != nil {
				t.Fatalf("Failed to write go.mod: %v", err)
			}

			parser := NewParser()
			framework, err := parser.DetectFramework(tmpDir)
			if err != nil {
				t.Fatalf("DetectFramework failed: %v", err)
			}

			if framework != tt.expected {
				t.Errorf("Expected framework %q, got %q", tt.expected, framework)
			}
		})
	}
}

func TestGoParser_ExtractChiRoutes(t *testing.T) {
	tmpDir := t.TempDir()

	// Write go.mod
	goMod := `module example.com/myapp

go 1.21

require github.com/go-chi/chi/v5 v5.0.0
`
	err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Write main.go with Chi routes
	mainGo := `package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func main() {
	r := chi.NewRouter()

	r.Get("/users", listUsers)
	r.Post("/users", createUser)
	r.Put("/users/{id}", updateUser)
	r.Delete("/users/{id}", deleteUser)

	http.ListenAndServe(":8080", r)
}

func listUsers(w http.ResponseWriter, r *http.Request) {}
func createUser(w http.ResponseWriter, r *http.Request) {}
func updateUser(w http.ResponseWriter, r *http.Request) {}
func deleteUser(w http.ResponseWriter, r *http.Request) {}
`
	err = os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGo), 0644)
	if err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if codebase.Framework != "chi" {
		t.Errorf("Expected framework chi, got %q", codebase.Framework)
	}

	// Should extract 4 routes
	if len(codebase.Routes) != 4 {
		t.Errorf("Expected 4 routes, got %d", len(codebase.Routes))
	}

	// Verify route methods and paths
	expectedRoutes := map[string]string{
		"GET":    "/users",
		"POST":   "/users",
		"PUT":    "/users/{id}",
		"DELETE": "/users/{id}",
	}

	for _, route := range codebase.Routes {
		expectedPath, ok := expectedRoutes[route.Method]
		if !ok {
			t.Errorf("Unexpected route method: %s", route.Method)
			continue
		}
		if route.Path != expectedPath {
			t.Errorf("Expected path %q for method %s, got %q", expectedPath, route.Method, route.Path)
		}
	}
}

func TestGoParser_ExtractGinRoutes(t *testing.T) {
	tmpDir := t.TempDir()

	// Write go.mod
	goMod := `module example.com/myapp

go 1.21

require github.com/gin-gonic/gin v1.9.0
`
	err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Write main.go with Gin routes
	mainGo := `package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.Default()

	r.GET("/users", listUsers)
	r.POST("/users", createUser)
	r.PUT("/users/:id", updateUser)
	r.DELETE("/users/:id", deleteUser)

	r.Run(":8080")
}

func listUsers(c *gin.Context) {}
func createUser(c *gin.Context) {}
func updateUser(c *gin.Context) {}
func deleteUser(c *gin.Context) {}
`
	err = os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGo), 0644)
	if err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if codebase.Framework != "gin" {
		t.Errorf("Expected framework gin, got %q", codebase.Framework)
	}

	// Should extract 4 routes
	if len(codebase.Routes) != 4 {
		t.Errorf("Expected 4 routes, got %d", len(codebase.Routes))
	}
}

func TestGoParser_ExtractEchoRoutes(t *testing.T) {
	tmpDir := t.TempDir()

	// Write go.mod
	goMod := `module example.com/myapp

go 1.21

require github.com/labstack/echo/v4 v4.10.0
`
	err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Write main.go with Echo routes
	mainGo := `package main

import "github.com/labstack/echo/v4"

func main() {
	e := echo.New()

	e.GET("/users", listUsers)
	e.POST("/users", createUser)
	e.PUT("/users/:id", updateUser)
	e.DELETE("/users/:id", deleteUser)

	e.Start(":8080")
}

func listUsers(c echo.Context) error { return nil }
func createUser(c echo.Context) error { return nil }
func updateUser(c echo.Context) error { return nil }
func deleteUser(c echo.Context) error { return nil }
`
	err = os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGo), 0644)
	if err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if codebase.Framework != "echo" {
		t.Errorf("Expected framework echo, got %q", codebase.Framework)
	}

	// Should extract 4 routes
	if len(codebase.Routes) != 4 {
		t.Errorf("Expected 4 routes, got %d", len(codebase.Routes))
	}
}

func TestGoParser_ExtractModels(t *testing.T) {
	tmpDir := t.TempDir()

	// Write go.mod
	goMod := `module example.com/myapp

go 1.21
`
	err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Write models.go
	modelsGo := `package main

type User struct {
	ID    int    ` + "`json:\"id\"`" + `
	Name  string ` + "`json:\"name\"`" + `
	Email string ` + "`json:\"email\"`" + `
}

type Product struct {
	ID    int     ` + "`json:\"id\"`" + `
	Name  string  ` + "`json:\"name\"`" + `
	Price float64 ` + "`json:\"price\"`" + `
}
`
	err = os.WriteFile(filepath.Join(tmpDir, "models.go"), []byte(modelsGo), 0644)
	if err != nil {
		t.Fatalf("Failed to write models.go: %v", err)
	}

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should extract 2 models
	if len(codebase.Models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(codebase.Models))
	}

	// Verify model names
	modelNames := make(map[string]bool)
	for _, model := range codebase.Models {
		modelNames[model.Name] = true
	}

	if !modelNames["User"] {
		t.Error("Expected to find User model")
	}
	if !modelNames["Product"] {
		t.Error("Expected to find Product model")
	}
}

func TestGoParser_ExtractHandlers(t *testing.T) {
	tmpDir := t.TempDir()

	// Write go.mod
	goMod := `module example.com/myapp

go 1.21
`
	err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Write handlers.go
	handlersGo := `package main

import "net/http"

func listUsers(w http.ResponseWriter, r *http.Request) {}
func createUser(w http.ResponseWriter, r *http.Request) {}
func updateUser(w http.ResponseWriter, r *http.Request) {}
func deleteUser(w http.ResponseWriter, r *http.Request) {}

// Helper functions should also be extracted
func validateUser(user *User) error { return nil }
func authenticateUser(token string) (*User, error) { return nil, nil }
`
	err = os.WriteFile(filepath.Join(tmpDir, "handlers.go"), []byte(handlersGo), 0644)
	if err != nil {
		t.Fatalf("Failed to write handlers.go: %v", err)
	}

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should extract 6 handler functions
	if len(codebase.Handlers) != 6 {
		t.Errorf("Expected 6 handlers, got %d", len(codebase.Handlers))
	}

	// Verify handler names
	handlerNames := make(map[string]bool)
	for _, handler := range codebase.Handlers {
		handlerNames[handler.Name] = true
	}

	expectedHandlers := []string{"listUsers", "createUser", "updateUser", "deleteUser", "validateUser", "authenticateUser"}
	for _, name := range expectedHandlers {
		if !handlerNames[name] {
			t.Errorf("Expected to find handler %q", name)
		}
	}
}

func TestGoParser_SkipTestFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Write go.mod
	goMod := `module example.com/myapp

go 1.21

require github.com/go-chi/chi/v5 v5.0.0
`
	err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Write main.go with routes
	mainGo := `package main

import "github.com/go-chi/chi/v5"

func main() {
	r := chi.NewRouter()
	r.Get("/users", listUsers)
}

func listUsers() {}
`
	err = os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGo), 0644)
	if err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}

	// Write test file (should be skipped)
	testGo := `package main

import "testing"

func TestListUsers(t *testing.T) {
	// This should not be parsed as a handler
}

func TestCreateUser(t *testing.T) {
	// This should not be parsed as a handler
}
`
	err = os.WriteFile(filepath.Join(tmpDir, "main_test.go"), []byte(testGo), 0644)
	if err != nil {
		t.Fatalf("Failed to write main_test.go: %v", err)
	}

	parser := NewParser()
	codebase, err := parser.Parse(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should only extract 1 route from main.go, not from test file
	if len(codebase.Routes) != 1 {
		t.Errorf("Expected 1 route (test files should be skipped), got %d", len(codebase.Routes))
	}

	// Should only extract 1 handler from main.go
	if len(codebase.Handlers) != 1 {
		t.Errorf("Expected 1 handler (test functions should be skipped), got %d", len(codebase.Handlers))
	}
}
