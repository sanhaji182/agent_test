package javascript

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/gotest-agent/internal/parser/types"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
)

// Parser implements the Parser interface for JavaScript/Node.js
type Parser struct {
	sitterParser *sitter.Parser
}

// NewParser creates a new JavaScript parser
func NewParser() *Parser {
	p := sitter.NewParser()
	p.SetLanguage(javascript.GetLanguage())
	return &Parser{
		sitterParser: p,
	}
}

// Parse analyzes a JavaScript codebase
func (p *Parser) Parse(ctx context.Context, rootDir string) (*types.Codebase, error) {
	codebase := &types.Codebase{
		Language: "javascript",
		RootDir:  rootDir,
	}

	// Detect framework
	framework, err := p.DetectFramework(rootDir)
	if err == nil {
		codebase.Framework = framework
	}

	// Find all JavaScript/TypeScript files
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
	return []string{"javascript", "typescript"}
}

// DetectFramework detects the JavaScript framework used
func (p *Parser) DetectFramework(rootDir string) (string, error) {
	// Check package.json
	packagePath := filepath.Join(rootDir, "package.json")
	content, err := os.ReadFile(packagePath)
	if err != nil {
		return "", fmt.Errorf("failed to read package.json: %w", err)
	}

	contentStr := string(content)

	// Check for common frameworks
	if contains(contentStr, "express") {
		return "express", nil
	}
	if contains(contentStr, "next") {
		return "nextjs", nil
	}
	if contains(contentStr, "nuxt") {
		return "nuxtjs", nil
	}
	if contains(contentStr, "react") {
		return "react", nil
	}
	if contains(contentStr, "vue") {
		return "vue", nil
	}
	if contains(contentStr, "angular") {
		return "angular", nil
	}
	if contains(contentStr, "fastify") {
		return "fastify", nil
	}
	if contains(contentStr, "hapi") {
		return "hapi", nil
	}

	return "nodejs", nil
}

// findSourceFiles finds all JavaScript and TypeScript files
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
			if name == "node_modules" || name == ".git" || name == "dist" || name == "build" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check file extension
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".js" || ext == ".jsx" || ext == ".ts" || ext == ".tsx" {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// parseFile parses a single JavaScript file
func (p *Parser) parseFile(ctx context.Context, filePath string, codebase *types.Codebase) error {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse with tree-sitter
	tree, err := p.sitterParser.ParseCtx(ctx, nil, content)
	if err != nil {
		return fmt.Errorf("failed to parse: %w", err)
	}

	// Extract routes, models, handlers
	routes := p.extractRoutes(tree.RootNode(), content, filePath)
	models := p.extractModels(tree.RootNode(), content, filePath)
	handlers := p.extractHandlers(tree.RootNode(), content, filePath)

	// Add to codebase
	codebase.Routes = append(codebase.Routes, routes...)
	codebase.Models = append(codebase.Models, models...)
	codebase.Handlers = append(codebase.Handlers, handlers...)

	return nil
}

// extractRoutes extracts Express.js routes from the AST
func (p *Parser) extractRoutes(node *sitter.Node, content []byte, filePath string) []types.Route {
	var routes []types.Route

	// Look for Express route patterns:
	// app.get('/path', handler)
	// router.post('/path', middleware, handler)
	// app.use('/path', router)

	// Traverse AST looking for call expressions
	p.traverseNode(node, func(n *sitter.Node) {
		if n.Type() == "call_expression" {
			route := p.parseRouteCall(n, content, filePath)
			if route != nil {
				routes = append(routes, *route)
			}
		}
	})

	return routes
}

// parseRouteCall attempts to parse a call expression as an Express route
func (p *Parser) parseRouteCall(node *sitter.Node, content []byte, filePath string) *types.Route {
	// Get function name (should be get, post, put, delete, use, etc.)
	functionNode := node.ChildByFieldName("function")
	if functionNode == nil {
		return nil
	}

	// Check if it's a member expression (app.get, router.post)
	if functionNode.Type() != "member_expression" {
		return nil
	}

	propertyNode := functionNode.ChildByFieldName("property")
	if propertyNode == nil {
		return nil
	}

	method := string(content[propertyNode.StartByte():propertyNode.EndByte()])

	// Check if it's an HTTP method
	httpMethods := map[string]string{
		"get":    "GET",
		"post":   "POST",
		"put":    "PUT",
		"delete": "DELETE",
		"patch":  "PATCH",
		"head":   "HEAD",
		"options": "OPTIONS",
		"use":    "USE", // Middleware mounting
	}

	httpMethod, isRoute := httpMethods[method]
	if !isRoute {
		return nil
	}

	// Get arguments
	argsNode := node.ChildByFieldName("arguments")
	if argsNode == nil || argsNode.NamedChildCount() < 2 {
		return nil
	}

	// First argument should be the path (string)
	pathNode := argsNode.NamedChild(0)
	if pathNode.Type() != "string" {
		return nil
	}

	path := string(content[pathNode.StartByte():pathNode.EndByte()])
	// Remove quotes
	path = strings.Trim(path, "'\"`")

	// Remaining arguments are middleware/handlers
	var middleware []string
	for i := 1; i < int(argsNode.NamedChildCount()); i++ {
		argNode := argsNode.NamedChild(i)
		if argNode.Type() == "identifier" {
			name := string(content[argNode.StartByte():argNode.EndByte()])
			middleware = append(middleware, name)
		}
	}

	// Get line number
	line := int(node.StartPoint().Row) + 1

	return &types.Route{
		Method:     httpMethod,
		Path:       path,
		Middleware: middleware,
		File:       filePath,
		Line:       line,
	}
}

// extractModels extracts Mongoose schema definitions
func (p *Parser) extractModels(node *sitter.Node, content []byte, filePath string) []types.Model {
	var models []types.Model

	// Look for Mongoose schema patterns:
	// const schema = new mongoose.Schema({...})
	// const User = mongoose.model('User', schema)

	// This is a simplified implementation
	// Full implementation would extract schema fields, types, validations

	return models
}

// extractHandlers extracts handler functions
func (p *Parser) extractHandlers(node *sitter.Node, content []byte, filePath string) []types.Handler {
	var handlers []types.Handler

	// Look for exported functions that could be handlers:
	// exports.handler = function(req, res) {...}
	// module.exports = { handler: function(req, res) {...} }

	// This is a simplified implementation
	// Full implementation would extract function signatures, parameters, return types

	return handlers
}

// traverseNode recursively traverses the AST
func (p *Parser) traverseNode(node *sitter.Node, callback func(*sitter.Node)) {
	callback(node)

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		p.traverseNode(child, callback)
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

// findSubstring finds a substring in a string
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
