package php

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/gotest-agent/internal/parser/types"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/php"
)

type Parser struct {
	treeSitterParser *sitter.Parser
}

func NewParser() *Parser {
	p := sitter.NewParser()
	p.SetLanguage(php.GetLanguage())
	return &Parser{
		treeSitterParser: p,
	}
}

func (p *Parser) SupportedLanguages() []string {
	return []string{"php"}
}

func (p *Parser) Parse(ctx context.Context, rootDir string) (*types.Codebase, error) {
	codebase := &types.Codebase{
		Language:  "php",
		Framework: p.detectFramework(rootDir),
		Routes:    make([]types.Route, 0),
		Models:    make([]types.Model, 0),
		Handlers:  make([]types.Handler, 0),
	}

	// Parse routes based on framework
	switch codebase.Framework {
	case "laravel":
		if err := p.parseLaravelRoutes(ctx, rootDir, codebase); err != nil {
			return nil, fmt.Errorf("failed to parse Laravel routes: %w", err)
		}
		if err := p.parseLaravelModels(ctx, rootDir, codebase); err != nil {
			return nil, fmt.Errorf("failed to parse Laravel models: %w", err)
		}
		if err := p.parseLaravelControllers(ctx, rootDir, codebase); err != nil {
			return nil, fmt.Errorf("failed to parse Laravel controllers: %w", err)
		}
	case "symfony":
		if err := p.parseSymfonyRoutes(ctx, rootDir, codebase); err != nil {
			return nil, fmt.Errorf("failed to parse Symfony routes: %w", err)
		}
	}

	return codebase, nil
}

// DetectFramework detects the PHP framework used
func (p *Parser) DetectFramework(rootDir string) (string, error) {
	return p.detectFramework(rootDir), nil
}

func (p *Parser) detectFramework(rootDir string) string {
	// Check for Laravel
	if _, err := os.Stat(filepath.Join(rootDir, "artisan")); err == nil {
		return "laravel"
	}

	// Check for Symfony
	if _, err := os.Stat(filepath.Join(rootDir, "bin/console")); err == nil {
		return "symfony"
	}

	return "php"
}

func (p *Parser) parseLaravelRoutes(ctx context.Context, rootDir string, codebase *types.Codebase) error {
	// Laravel routes are typically in routes/web.php and routes/api.php
	routeFiles := []string{
		filepath.Join(rootDir, "routes", "web.php"),
		filepath.Join(rootDir, "routes", "api.php"),
	}

	for _, routeFile := range routeFiles {
		if _, err := os.Stat(routeFile); os.IsNotExist(err) {
			continue
		}

		routes, err := p.extractRoutesFromFile(ctx, routeFile)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", routeFile, err)
		}

		codebase.Routes = append(codebase.Routes, routes...)
	}

	return nil
}

func (p *Parser) extractRoutesFromFile(ctx context.Context, filePath string) ([]types.Route, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	tree, err := p.treeSitterParser.ParseCtx(ctx, nil, content)
	if err != nil {
		return nil, err
	}

	routes := []types.Route{}
	p.traverseNode(tree.RootNode(), func(node *sitter.Node) {
		if node.Type() == "scoped_call_expression" {
			route := p.extractLaravelRoute(node, content, filePath)
			if route != nil {
				routes = append(routes, *route)
			}
		}
	})

	return routes, nil
}

func (p *Parser) extractLaravelRoute(node *sitter.Node, content []byte, filePath string) *types.Route {
	// Laravel route pattern: Route::get('/path', [Controller::class, 'method'])
	// or Route::get('/path', function() { ... })
	// In tree-sitter-php this is a scoped_call_expression with fields:
	// scope ("Route"), name ("get"), arguments.

	scopeNode := node.ChildByFieldName("scope")
	if scopeNode == nil || scopeNode.Type() != "name" {
		return nil
	}

	scopeName := string(content[scopeNode.StartByte():scopeNode.EndByte()])
	if scopeName != "Route" {
		return nil
	}

	methodNode := node.ChildByFieldName("name")
	if methodNode == nil || methodNode.Type() != "name" {
		return nil
	}

	method := string(content[methodNode.StartByte():methodNode.EndByte()])

	// Map Laravel methods to HTTP methods
	httpMethod := ""
	switch strings.ToLower(method) {
	case "get":
		httpMethod = "GET"
	case "post":
		httpMethod = "POST"
	case "put":
		httpMethod = "PUT"
	case "patch":
		httpMethod = "PATCH"
	case "delete":
		httpMethod = "DELETE"
	case "options":
		httpMethod = "OPTIONS"
	default:
		return nil
	}

	// Extract arguments
	argumentsNode := node.ChildByFieldName("arguments")
	if argumentsNode == nil {
		return nil
	}

	// First argument should be the path (string)
	pathArg := argumentsNode.NamedChild(0)
	if pathArg == nil {
		return nil
	}

	// Unwrap argument node
	if pathArg.Type() == "argument" {
		pathArg = pathArg.NamedChild(0)
	}
	if pathArg == nil {
		return nil
	}

	path := ""
	if pathArg.Type() == "string" || pathArg.Type() == "encapsed_string" {
		path = string(content[pathArg.StartByte():pathArg.EndByte()])
		// Remove quotes
		path = strings.Trim(path, "'\"")
	}

	if path == "" {
		return nil
	}

	// Extract handler (second argument)
	handler := ""
	if argumentsNode.NamedChildCount() > 1 {
		handlerArg := argumentsNode.NamedChild(1)
		if handlerArg.Type() == "argument" && handlerArg.NamedChildCount() > 0 {
			handlerArg = handlerArg.NamedChild(0)
		}
		handler = string(content[handlerArg.StartByte():handlerArg.EndByte()])
	}

	// Extract middleware if present (->middleware())
	middleware := []string{}
	parent := node.Parent()
	if parent != nil && parent.Type() == "member_call_expression" {
		// Check if there's a middleware call
		p.traverseNode(parent, func(n *sitter.Node) {
			if n.Type() == "member_call_expression" {
				methodName := p.getNodeContent(n.ChildByFieldName("name"), content)
				if methodName == "middleware" {
					// Extract middleware names from arguments
					args := n.ChildByFieldName("arguments")
					if args != nil {
						for i := 0; i < int(args.NamedChildCount()); i++ {
							arg := args.NamedChild(i)
							if arg.Type() != "argument" {
								continue
							}
							value := arg.NamedChild(0)
							if value == nil {
								continue
							}
							if value.Type() == "array_creation_expression" {
								// ->middleware(['auth', 'verified'])
								for j := 0; j < int(value.NamedChildCount()); j++ {
									elem := value.NamedChild(j)
									if elem.Type() != "array_element_initializer" {
										continue
									}
									elemValue := elem.NamedChild(0)
									if elemValue == nil {
										continue
									}
									mwName := strings.Trim(p.getNodeContent(elemValue, content), "'\"")
									if mwName != "" {
										middleware = append(middleware, mwName)
									}
								}
							} else {
								// ->middleware('auth')
								mwName := strings.Trim(p.getNodeContent(value, content), "'\"")
								if mwName != "" {
									middleware = append(middleware, mwName)
								}
							}
						}
					}
				}
			}
		})
	}

	line := p.getNodeLine(node, content)

	return &types.Route{
		Method:     httpMethod,
		Path:       path,
		Handler:    handler,
		Middleware: middleware,
		File:       filePath,
		Line:       line,
	}
}

func (p *Parser) parseLaravelModels(ctx context.Context, rootDir string, codebase *types.Codebase) error {
	modelsDir := filepath.Join(rootDir, "app", "Models")
	if _, err := os.Stat(modelsDir); os.IsNotExist(err) {
		return nil
	}

	return filepath.Walk(modelsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".php") {
			return nil
		}

		models, err := p.extractModelsFromFile(ctx, path)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		codebase.Models = append(codebase.Models, models...)
		return nil
	})
}

func (p *Parser) extractModelsFromFile(ctx context.Context, filePath string) ([]types.Model, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	tree, err := p.treeSitterParser.ParseCtx(ctx, nil, content)
	if err != nil {
		return nil, err
	}

	models := []types.Model{}
	p.traverseNode(tree.RootNode(), func(node *sitter.Node) {
		if node.Type() == "class_declaration" {
			model := p.extractLaravelModel(node, content, filePath)
			if model != nil {
				models = append(models, *model)
			}
		}
	})

	return models, nil
}

func (p *Parser) extractLaravelModel(node *sitter.Node, content []byte, filePath string) *types.Model {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	name := string(content[nameNode.StartByte():nameNode.EndByte()])

	// Check if it extends Model
	bodyNode := node.ChildByFieldName("body")
	if bodyNode == nil {
		return nil
	}

	// Extract table name if specified
	// tree-sitter-php: property_declaration > property_element >
	//   variable_name ("$table") + property_initializer ("= 'users'")
	table := ""
	if value := p.findPropertyValue(bodyNode, "table", content); value != nil {
		table = strings.Trim(p.getNodeContent(value, content), "'\"")
	}

	// Extract fillable fields
	fillable := []types.Field{}
	if value := p.findPropertyValue(bodyNode, "fillable", content); value != nil && value.Type() == "array_creation_expression" {
		for i := 0; i < int(value.NamedChildCount()); i++ {
			elem := value.NamedChild(i)
			if elem.Type() != "array_element_initializer" {
				continue
			}
			fieldValue := elem.NamedChild(0)
			if fieldValue == nil {
				continue
			}
			fieldName := strings.Trim(p.getNodeContent(fieldValue, content), "'\"")
			if fieldName != "" {
				fillable = append(fillable, types.Field{Name: fieldName})
			}
		}
	}

	line := p.getNodeLine(node, content)

	return &types.Model{
		Name:     name,
		Table:    table,
		Fields:   fillable,
		File:     filePath,
		Line:     line,
	}
}

// findPropertyValue searches a class body for a property named propName
// (without the leading $) and returns the value node of its initializer.
func (p *Parser) findPropertyValue(bodyNode *sitter.Node, propName string, content []byte) *sitter.Node {
	var result *sitter.Node
	p.traverseNode(bodyNode, func(n *sitter.Node) {
		if result != nil || n.Type() != "property_element" {
			return
		}
		var varName *sitter.Node
		var initializer *sitter.Node
		for i := 0; i < int(n.NamedChildCount()); i++ {
			child := n.NamedChild(i)
			switch child.Type() {
			case "variable_name":
				varName = child
			case "property_initializer":
				initializer = child
			}
		}
		if varName == nil || initializer == nil {
			return
		}
		name := strings.TrimPrefix(p.getNodeContent(varName, content), "$")
		if name != propName {
			return
		}
		result = initializer.NamedChild(0)
	})
	return result
}

func (p *Parser) parseLaravelControllers(ctx context.Context, rootDir string, codebase *types.Codebase) error {
	controllersDir := filepath.Join(rootDir, "app", "Http", "Controllers")
	if _, err := os.Stat(controllersDir); os.IsNotExist(err) {
		return nil
	}

	return filepath.Walk(controllersDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".php") {
			return nil
		}

		handlers, err := p.extractHandlersFromFile(ctx, path)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		codebase.Handlers = append(codebase.Handlers, handlers...)
		return nil
	})
}

func (p *Parser) extractHandlersFromFile(ctx context.Context, filePath string) ([]types.Handler, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	tree, err := p.treeSitterParser.ParseCtx(ctx, nil, content)
	if err != nil {
		return nil, err
	}

	handlers := []types.Handler{}

	// Extract controller name
	controllerName := ""
	p.traverseNode(tree.RootNode(), func(node *sitter.Node) {
		if node.Type() == "class_declaration" {
			nameNode := node.ChildByFieldName("name")
			if nameNode != nil {
				controllerName = string(content[nameNode.StartByte():nameNode.EndByte()])
			}
		}
	})

	// Extract methods
	p.traverseNode(tree.RootNode(), func(node *sitter.Node) {
		if node.Type() == "method_declaration" {
			handler := p.extractControllerMethod(node, content, filePath, controllerName)
			if handler != nil {
				handlers = append(handlers, *handler)
			}
		}
	})

	return handlers, nil
}

func (p *Parser) extractControllerMethod(node *sitter.Node, content []byte, filePath, controllerName string) *types.Handler {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	name := string(content[nameNode.StartByte():nameNode.EndByte()])

	// Skip constructor and private methods
	if name == "__construct" || name == "__destruct" {
		return nil
	}

	// Check visibility
	visibility := "public"
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "visibility_modifier" {
			visibility = string(content[child.StartByte():child.EndByte()])
			break
		}
	}

	if visibility != "public" {
		return nil
	}

	line := p.getNodeLine(node, content)

	return &types.Handler{
		Name:       name,
		Controller: controllerName,
		File:       filePath,
		Line:       line,
	}
}

func (p *Parser) parseSymfonyRoutes(ctx context.Context, rootDir string, codebase *types.Codebase) error {
	// Symfony routes are typically defined in controllers with annotations/attributes
	// This is a simplified implementation
	return nil
}

func (p *Parser) traverseNode(node *sitter.Node, visit func(*sitter.Node)) {
	if node == nil {
		return
	}
	visit(node)
	for i := 0; i < int(node.ChildCount()); i++ {
		p.traverseNode(node.Child(i), visit)
	}
}

func (p *Parser) getNodeContent(node *sitter.Node, content []byte) string {
	if node == nil {
		return ""
	}
	return string(content[node.StartByte():node.EndByte()])
}

func (p *Parser) getNodeLine(node *sitter.Node, content []byte) int {
	if node == nil {
		return 0
	}
	return int(node.StartPoint().Row) + 1
}
