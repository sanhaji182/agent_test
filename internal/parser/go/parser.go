package golang

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/gotest-agent/internal/parser/types"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

// Parser implements the Parser interface for Go
type Parser struct {
	sitterParser *sitter.Parser
}

// NewParser creates a new Go parser
func NewParser() *Parser {
	p := sitter.NewParser()
	p.SetLanguage(golang.GetLanguage())
	return &Parser{
		sitterParser: p,
	}
}

// Parse analyzes a Go codebase
func (p *Parser) Parse(ctx context.Context, rootDir string) (*types.Codebase, error) {
	codebase := &types.Codebase{
		Language: "go",
		RootDir:  rootDir,
	}

	// Detect framework
	framework, err := p.DetectFramework(rootDir)
	if err == nil {
		codebase.Framework = framework
	}

	// Find all Go files
	files, err := p.findSourceFiles(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to find source files: %w", err)
	}

	// Parse each file
	for _, file := range files {
		if err := p.parseFile(ctx, file, codebase); err != nil {
			// Log error but continue parsing other files
			fmt.Printf("Warning: failed to parse %s: %v\n", file, err)
		}
	}

	return codebase, nil
}

// SupportedLanguages returns the languages this parser supports
func (p *Parser) SupportedLanguages() []string {
	return []string{"go"}
}

// DetectFramework detects the Go framework used
func (p *Parser) DetectFramework(rootDir string) (string, error) {
	// Check go.mod
	modPath := filepath.Join(rootDir, "go.mod")
	content, err := os.ReadFile(modPath)
	if err != nil {
		return "", fmt.Errorf("failed to read go.mod: %w", err)
	}

	contentStr := string(content)

	// Check for common frameworks
	if strings.Contains(contentStr, "github.com/gin-gonic/gin") {
		return "gin", nil
	}
	if strings.Contains(contentStr, "github.com/labstack/echo") {
		return "echo", nil
	}
	if strings.Contains(contentStr, "github.com/go-chi/chi") {
		return "chi", nil
	}
	if strings.Contains(contentStr, "github.com/gofiber/fiber") {
		return "fiber", nil
	}
	if strings.Contains(contentStr, "github.com/gorilla/mux") {
		return "gorilla", nil
	}
	if strings.Contains(contentStr, "github.com/go-macaron/macaron") {
		return "macaron", nil
	}

	return "stdlib", nil
}

// findSourceFiles finds all Go source files
func (p *Parser) findSourceFiles(rootDir string) ([]string, error) {
	var files []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			// Skip common non-source directories
			name := info.Name()
			if name == "vendor" || name == ".git" || name == "bin" || name == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check file extension
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".go" {
			// Skip test files
			if !strings.HasSuffix(path, "_test.go") {
				files = append(files, path)
			}
		}

		return nil
	})

	return files, err
}

// parseFile parses a single Go file
func (p *Parser) parseFile(ctx context.Context, filePath string, codebase *types.Codebase) error {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse with tree-sitter
	tree, err := p.sitterParser.ParseCtx(ctx, content, nil)
	if err != nil {
		return fmt.Errorf("failed to parse: %w", err)
	}

	// Extract routes, models, handlers based on framework
	switch codebase.Framework {
	case "chi":
		routes := p.extractChiRoutes(tree.RootNode(), content, filePath)
		codebase.Routes = append(codebase.Routes, routes...)
	case "gin":
		routes := p.extractGinRoutes(tree.RootNode(), content, filePath)
		codebase.Routes = append(codebase.Routes, routes...)
	case "echo":
		routes := p.extractEchoRoutes(tree.RootNode(), content, filePath)
		codebase.Routes = append(codebase.Routes, routes...)
	case "fiber":
		routes := p.extractFiberRoutes(tree.RootNode(), content, filePath)
		codebase.Routes = append(codebase.Routes, routes...)
	case "gorilla":
		routes := p.extractGorillaRoutes(tree.RootNode(), content, filePath)
		codebase.Routes = append(codebase.Routes, routes...)
	}

	// Extract models (structs with json tags)
	models := p.extractModels(tree.RootNode(), content, filePath)
	codebase.Models = append(codebase.Models, models...)

	// Extract handlers
	handlers := p.extractHandlers(tree.RootNode(), content, filePath)
	codebase.Handlers = append(codebase.Handlers, handlers...)

	return nil
}

// extractChiRoutes extracts Chi router routes
func (p *Parser) extractChiRoutes(node *sitter.Node, content []byte, filePath string) []types.Route {
	var routes []types.Route

	// Chi patterns:
	// r.Get("/path", handler)
	// r.Post("/path", handler)
	// r.Route("/prefix", func(r chi.Router) { ... })
	// r.Group(func(r chi.Router) { ... })

	p.traverseNode(node, func(n *sitter.Node) {
		if n.Type() == "call_expression" {
			route := p.parseChiRouteCall(n, content, filePath)
			if route != nil {
				routes = append(routes, *route)
			}
		}
	})

	return routes
}

// parseChiRouteCall parses a Chi route call
func (p *Parser) parseChiRouteCall(node *sitter.Node, content []byte, filePath string) *types.Route {
	functionNode := node.ChildByFieldName("function")
	if functionNode == nil {
		return nil
	}

	if functionNode.Type() != "selector_expression" {
		return nil
	}

	methodNode := functionNode.ChildByFieldName("field")
	if methodNode == nil {
		return nil
	}

	method := string(content[methodNode.StartByte():methodNode.EndByte()])

	httpMethods := map[string]string{
		"Get":     "GET",
		"Post":    "POST",
		"Put":     "PUT",
		"Delete":  "DELETE",
		"Patch":   "PATCH",
		"Head":    "HEAD",
		"Options": "OPTIONS",
		"Handle":  "HANDLE",
	}

	httpMethod, isRoute := httpMethods[method]
	if !isRoute {
		return nil
	}

	argsNode := node.ChildByFieldName("arguments")
	if argsNode == nil || argsNode.ChildCount() < 2 {
		return nil
	}

	// First argument is path
	pathNode := argsNode.Child(0)
	if pathNode.Type() != "interpreted_string_literal" {
		return nil
	}

	path := string(content[pathNode.StartByte():pathNode.EndByte()])
	path = strings.Trim(path, "\"")

	// Remaining arguments are middleware/handlers
	var middleware []string
	for i := 1; i < int(argsNode.ChildCount()); i++ {
		argNode := argsNode.Child(i)
		if argNode.Type() == "identifier" {
			name := string(content[argNode.StartByte():argNode.EndByte()])
			middleware = append(middleware, name)
		}
	}

	line := int(node.StartPoint().Row) + 1

	return &types.Route{
		Method:     httpMethod,
		Path:       path,
		Middleware: middleware,
		File:       filePath,
		Line:       line,
	}
}

// extractGinRoutes extracts Gin router routes
func (p *Parser) extractGinRoutes(node *sitter.Node, content []byte, filePath string) []types.Route {
	var routes []types.Route

	// Gin patterns:
	// r.GET("/path", handler)
	// r.POST("/path", handler)
	// r.PUT("/path", handler)

	p.traverseNode(node, func(n *sitter.Node) {
		if n.Type() == "call_expression" {
			route := p.parseGinRouteCall(n, content, filePath)
			if route != nil {
				routes = append(routes, *route)
			}
		}
	})

	return routes
}

// parseGinRouteCall parses a Gin route call
func (p *Parser) parseGinRouteCall(node *sitter.Node, content []byte, filePath string) *types.Route {
	functionNode := node.ChildByFieldName("function")
	if functionNode == nil {
		return nil
	}

	if functionNode.Type() != "selector_expression" {
		return nil
	}

	methodNode := functionNode.ChildByFieldName("field")
	if methodNode == nil {
		return nil
	}

	method := string(content[methodNode.StartByte():methodNode.EndByte()])

	// Gin uses uppercase methods
	httpMethods := map[string]string{
		"GET":     "GET",
		"POST":    "POST",
		"PUT":     "PUT",
		"DELETE":  "DELETE",
		"PATCH":   "PATCH",
		"HEAD":    "HEAD",
		"OPTIONS": "OPTIONS",
		"Any":     "ANY",
	}

	httpMethod, isRoute := httpMethods[method]
	if !isRoute {
		return nil
	}

	argsNode := node.ChildByFieldName("arguments")
	if argsNode == nil || argsNode.ChildCount() < 2 {
		return nil
	}

	pathNode := argsNode.Child(0)
	if pathNode.Type() != "interpreted_string_literal" {
		return nil
	}

	path := string(content[pathNode.StartByte():pathNode.EndByte()])
	path = strings.Trim(path, "\"")

	var middleware []string
	for i := 1; i < int(argsNode.ChildCount()); i++ {
		argNode := argsNode.Child(i)
		if argNode.Type() == "identifier" {
			name := string(content[argNode.StartByte():argNode.EndByte()])
			middleware = append(middleware, name)
		}
	}

	line := int(node.StartPoint().Row) + 1

	return &types.Route{
		Method:     httpMethod,
		Path:       path,
		Middleware: middleware,
		File:       filePath,
		Line:       line,
	}
}

// extractEchoRoutes extracts Echo router routes
func (p *Parser) extractEchoRoutes(node *sitter.Node, content []byte, filePath string) []types.Route {
	var routes []types.Route

	// Echo patterns:
	// e.GET("/path", handler)
	// e.POST("/path", handler)

	p.traverseNode(node, func(n *sitter.Node) {
		if n.Type() == "call_expression" {
			route := p.parseEchoRouteCall(n, content, filePath)
			if route != nil {
				routes = append(routes, *route)
			}
		}
	})

	return routes
}

// parseEchoRouteCall parses an Echo route call (similar to Gin)
func (p *Parser) parseEchoRouteCall(node *sitter.Node, content []byte, filePath string) *types.Route {
	// Echo uses same pattern as Gin
	return p.parseGinRouteCall(node, content, filePath)
}

// extractFiberRoutes extracts Fiber router routes
func (p *Parser) extractFiberRoutes(node *sitter.Node, content []byte, filePath string) []types.Route {
	var routes []types.Route

	// Fiber patterns:
	// app.Get("/path", handler)
	// app.Post("/path", handler)

	p.traverseNode(node, func(n *sitter.Node) {
		if n.Type() == "call_expression" {
			route := p.parseFiberRouteCall(n, content, filePath)
			if route != nil {
				routes = append(routes, *route)
			}
		}
	})

	return routes
}

// parseFiberRouteCall parses a Fiber route call
func (p *Parser) parseFiberRouteCall(node *sitter.Node, content []byte, filePath string) *types.Route {
	functionNode := node.ChildByFieldName("function")
	if functionNode == nil {
		return nil
	}

	if functionNode.Type() != "selector_expression" {
		return nil
	}

	methodNode := functionNode.ChildByFieldName("field")
	if methodNode == nil {
		return nil
	}

	method := string(content[methodNode.StartByte():methodNode.EndByte()])

	// Fiber uses PascalCase methods
	httpMethods := map[string]string{
		"Get":     "GET",
		"Post":    "POST",
		"Put":     "PUT",
		"Delete":  "DELETE",
		"Patch":   "PATCH",
		"Head":    "HEAD",
		"Options": "OPTIONS",
		"All":     "ALL",
	}

	httpMethod, isRoute := httpMethods[method]
	if !isRoute {
		return nil
	}

	argsNode := node.ChildByFieldName("arguments")
	if argsNode == nil || argsNode.ChildCount() < 2 {
		return nil
	}

	pathNode := argsNode.Child(0)
	if pathNode.Type() != "interpreted_string_literal" {
		return nil
	}

	path := string(content[pathNode.StartByte():pathNode.EndByte()])
	path = strings.Trim(path, "\"")

	var middleware []string
	for i := 1; i < int(argsNode.ChildCount()); i++ {
		argNode := argsNode.Child(i)
		if argNode.Type() == "identifier" {
			name := string(content[argNode.StartByte():argNode.EndByte()])
			middleware = append(middleware, name)
		}
	}

	line := int(node.StartPoint().Row) + 1

	return &types.Route{
		Method:     httpMethod,
		Path:       path,
		Middleware: middleware,
		File:       filePath,
		Line:       line,
	}
}

// extractGorillaRoutes extracts Gorilla Mux routes
func (p *Parser) extractGorillaRoutes(node *sitter.Node, content []byte, filePath string) []types.Route {
	var routes []types.Route

	// Gorilla Mux patterns:
	// r.HandleFunc("/path", handler).Methods("GET")
	// r.Path("/path").HandlerFunc(handler)

	// Simplified implementation - full version would handle method chaining
	p.traverseNode(node, func(n *sitter.Node) {
		if n.Type() == "call_expression" {
			route := p.parseGorillaRouteCall(n, content, filePath)
			if route != nil {
				routes = append(routes, *route)
			}
		}
	})

	return routes
}

// parseGorillaRouteCall parses a Gorilla Mux route call
func (p *Parser) parseGorillaRouteCall(node *sitter.Node, content []byte, filePath string) *types.Route {
	functionNode := node.ChildByFieldName("function")
	if functionNode == nil {
		return nil
	}

	if functionNode.Type() != "selector_expression" {
		return nil
	}

	methodNode := functionNode.ChildByFieldName("field")
	if methodNode == nil {
		return nil
	}

	method := string(content[methodNode.StartByte():methodNode.EndByte()])

	if method != "HandleFunc" && method != "Handle" {
		return nil
	}

	argsNode := node.ChildByFieldName("arguments")
	if argsNode == nil || argsNode.ChildCount() < 2 {
		return nil
	}

	pathNode := argsNode.Child(0)
	if pathNode.Type() != "interpreted_string_literal" {
		return nil
	}

	path := string(content[pathNode.StartByte():pathNode.EndByte()])
	path = strings.Trim(path, "\"")

	line := int(node.StartPoint().Row) + 1

	return &types.Route{
		Method: "HANDLE", // Would need to parse .Methods() call for actual method
		Path:   path,
		File:   filePath,
		Line:   line,
	}
}

// extractModels extracts Go struct definitions as models
func (p *Parser) extractModels(node *sitter.Node, content []byte, filePath string) []types.Model {
	var models []types.Model

	// Look for struct type definitions with json tags
	p.traverseNode(node, func(n *sitter.Node) {
		if n.Type() == "type_declaration" {
			model := p.parseStructModel(n, content, filePath)
			if model != nil {
				models = append(models, *model)
			}
		}
	})

	return models
}

// parseStructModel parses a struct definition into a model
func (p *Parser) parseStructModel(node *sitter.Node, content []byte, filePath string) *types.Model {
	// This is a simplified implementation
	// Full version would extract fields, json tags, validation tags, etc.

	// Get struct name
	specNode := node.Child(0)
	if specNode == nil || specNode.Type() != "type_spec" {
		return nil
	}

	nameNode := specNode.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	name := string(content[nameNode.StartByte():nameNode.EndByte()])

	// Check if it's a struct
	typeNode := specNode.ChildByFieldName("type")
	if typeNode == nil || typeNode.Type() != "struct_type" {
		return nil
	}

	line := int(node.StartPoint().Row) + 1

	return &types.Model{
		Name: name,
		File: filePath,
		Line: line,
	}
}

// extractHandlers extracts handler functions
func (p *Parser) extractHandlers(node *sitter.Node, content []byte, filePath string) []types.Handler {
	var handlers []types.Handler

	// Look for function declarations
	p.traverseNode(node, func(n *sitter.Node) {
		if n.Type() == "function_declaration" {
			handler := p.parseHandlerFunction(n, content, filePath)
			if handler != nil {
				handlers = append(handlers, *handler)
			}
		}
	})

	return handlers
}

// parseHandlerFunction parses a function declaration as a handler
func (p *Parser) parseHandlerFunction(node *sitter.Node, content []byte, filePath string) *types.Handler {
	// This is a simplified implementation
	// Full version would extract parameters, return types, etc.

	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	name := string(content[nameNode.StartByte():nameNode.EndByte()])

	// Skip main function
	if name == "main" {
		return nil
	}

	line := int(node.StartPoint().Row) + 1

	return &types.Handler{
		Name: name,
		File: filePath,
		Line: line,
	}
}

// traverseNode recursively traverses the AST
func (p *Parser) traverseNode(node *sitter.Node, callback func(*sitter.Node)) {
	callback(node)

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		p.traverseNode(child, callback)
	}
}
