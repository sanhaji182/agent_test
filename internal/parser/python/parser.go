package python

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/gotest-agent/internal/parser/types"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"
)

type Parser struct {
	treeSitterParser *sitter.Parser
}

func NewParser() *Parser {
	p := sitter.NewParser()
	p.SetLanguage(python.GetLanguage())
	return &Parser{
		treeSitterParser: p,
	}
}

func (p *Parser) SupportedLanguages() []string {
	return []string{"python"}
}

func (p *Parser) Parse(ctx context.Context, rootDir string) (*types.Codebase, error) {
	codebase := &types.Codebase{
		Language:  "python",
		Framework: p.detectFramework(rootDir),
		Routes:    make([]types.Route, 0),
		Models:    make([]types.Model, 0),
		Handlers:  make([]types.Handler, 0),
	}

	// Parse routes based on framework
	switch codebase.Framework {
	case "django":
		if err := p.parseDjangoRoutes(ctx, rootDir, codebase); err != nil {
			return nil, fmt.Errorf("failed to parse Django routes: %w", err)
		}
		if err := p.parseDjangoModels(ctx, rootDir, codebase); err != nil {
			return nil, fmt.Errorf("failed to parse Django models: %w", err)
		}
		if err := p.parseDjangoViews(ctx, rootDir, codebase); err != nil {
			return nil, fmt.Errorf("failed to parse Django views: %w", err)
		}
	case "flask":
		if err := p.parseFlaskRoutes(ctx, rootDir, codebase); err != nil {
			return nil, fmt.Errorf("failed to parse Flask routes: %w", err)
		}
	case "fastapi":
		if err := p.parseFastAPIRoutes(ctx, rootDir, codebase); err != nil {
			return nil, fmt.Errorf("failed to parse FastAPI routes: %w", err)
		}
	}

	return codebase, nil
}

func (p *Parser) detectFramework(rootDir string) string {
	// Check for Django (manage.py exists)
	if _, err := os.Stat(filepath.Join(rootDir, "manage.py")); err == nil {
		return "django"
	}

	// Check requirements.txt or pyproject.toml for Flask/FastAPI
	reqFiles := []string{
		filepath.Join(rootDir, "requirements.txt"),
		filepath.Join(rootDir, "Pipfile"),
	}

	for _, reqFile := range reqFiles {
		content, err := os.ReadFile(reqFile)
		if err != nil {
			continue
		}

		contentStr := string(content)

		if strings.Contains(contentStr, "fastapi") {
			return "fastapi"
		}
		if strings.Contains(contentStr, "flask") || strings.Contains(contentStr, "Flask") {
			return "flask"
		}
	}

	// Check pyproject.toml
	pyprojectPath := filepath.Join(rootDir, "pyproject.toml")
	if content, err := os.ReadFile(pyprojectPath); err == nil {
		contentStr := string(content)
		if strings.Contains(contentStr, "fastapi") {
			return "fastapi"
		}
		if strings.Contains(contentStr, "flask") || strings.Contains(contentStr, "Flask") {
			return "flask"
		}
	}

	return "python"
}

func (p *Parser) parseDjangoRoutes(ctx context.Context, rootDir string, codebase *types.Codebase) error {
	// Django routes are typically in urls.py files
	// Look for urls.py in all apps
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if info.Name() == "urls.py" {
			routes, err := p.extractDjangoRoutesFromFile(ctx, path)
			if err != nil {
				return fmt.Errorf("failed to parse %s: %w", path, err)
			}
			codebase.Routes = append(codebase.Routes, routes...)
		}

		return nil
	})

	return err
}

func (p *Parser) extractDjangoRoutesFromFile(ctx context.Context, filePath string) ([]types.Route, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	tree, err := p.treeSitterParser.ParseCtx(ctx, content, nil)
	if err != nil {
		return nil, err
	}

	routes := []types.Route{}

	// Look for path() and re_path() calls in urlpatterns
	p.traverseNode(tree.RootNode(), func(node *sitter.Node) {
		if node.Type() == "call" {
			route := p.extractDjangoRoute(node, content, filePath)
			if route != nil {
				routes = append(routes, *route)
			}
		}
	})

	return routes, nil
}

func (p *Parser) extractDjangoRoute(node *sitter.Node, content []byte, filePath string) *types.Route {
	// Django route pattern: path('url/', view_function, name='route_name')
	// or re_path(r'^url/$', view_function)

	functionNode := node.ChildByFieldName("function")
	if functionNode == nil {
		return nil
	}

	functionName := string(content[functionNode.StartByte():functionNode.EndByte()])

	if functionName != "path" && functionName != "re_path" {
		return nil
	}

	// Extract arguments
	argumentsNode := node.ChildByFieldName("arguments")
	if argumentsNode == nil {
		return nil
	}

	// First argument should be the path (string)
	pathArg := argumentsNode.Child(0)
	if pathArg == nil {
		return nil
	}

	path := ""
	if pathArg.Type() == "string" {
		path = string(content[pathArg.StartByte():pathArg.EndByte()])
		// Remove quotes and r prefix
		path = strings.Trim(path, "'\"")
		path = strings.TrimPrefix(path, "r")
	}

	if path == "" {
		return nil
	}

	// Extract handler (second argument)
	handler := ""
	if argumentsNode.ChildCount() > 1 {
		handlerArg := argumentsNode.Child(1)
		handler = string(content[handlerArg.StartByte():handlerArg.EndByte()])
	}

	line := p.getNodeLine(node, content)

	return &types.Route{
		Method:  "ANY", // Django doesn't specify method in path()
		Path:    path,
		Handler: handler,
		File:    filePath,
		Line:    line,
	}
}

func (p *Parser) parseDjangoModels(ctx context.Context, rootDir string, codebase *types.Codebase) error {
	// Django models are typically in models.py files
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if info.Name() == "models.py" {
			models, err := p.extractDjangoModelsFromFile(ctx, path)
			if err != nil {
				return fmt.Errorf("failed to parse %s: %w", path, err)
			}
			codebase.Models = append(codebase.Models, models...)
		}

		return nil
	})

	return err
}

func (p *Parser) extractDjangoModelsFromFile(ctx context.Context, filePath string) ([]types.Model, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	tree, err := p.treeSitterParser.ParseCtx(ctx, content, nil)
	if err != nil {
		return nil, err
	}

	models := []types.Model{}

	// Look for class definitions that extend models.Model
	p.traverseNode(tree.RootNode(), func(node *sitter.Node) {
		if node.Type() == "class_definition" {
			model := p.extractDjangoModel(node, content, filePath)
			if model != nil {
				models = append(models, *model)
			}
		}
	})

	return models, nil
}

func (p *Parser) extractDjangoModel(node *sitter.Node, content []byte, filePath string) *types.Model {
	// Django model pattern: class ModelName(models.Model):
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	name := string(content[nameNode.StartByte():nameNode.EndByte()])

	// Check if it extends models.Model
	superclassesNode := node.ChildByFieldName("superclasses")
	if superclassesNode == nil {
		return nil
	}

	isModel := false
	for i := 0; i < superclassesNode.ChildCount(); i++ {
		superclass := superclassesNode.Child(i)
		superclassName := string(content[superclass.StartByte():superclass.EndByte()])
		if superclassName == "models.Model" || superclassName == "Model" {
			isModel = true
			break
		}
	}

	if !isModel {
		return nil
	}

	// Extract fields from class body
	bodyNode := node.ChildByFieldName("body")
	if bodyNode == nil {
		return nil
	}

	fields := []string{}
	p.traverseNode(bodyNode, func(n *sitter.Node) {
		if n.Type() == "expression_statement" {
			assignment := n.Child(0)
			if assignment != nil && assignment.Type() == "assignment" {
				left := assignment.ChildByFieldName("left")
				if left != nil {
					fieldName := string(content[left.StartByte():left.EndByte()])
					fields = append(fields, fieldName)
				}
			}
		}
	})

	line := p.getNodeLine(node, content)

	return &types.Model{
		Name:   name,
		Fields: fields,
		File:   filePath,
		Line:   line,
	}
}

func (p *Parser) parseDjangoViews(ctx context.Context, rootDir string, codebase *types.Codebase) error {
	// Django views are typically in views.py files
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if info.Name() == "views.py" {
			handlers, err := p.extractDjangoViewsFromFile(ctx, path)
			if err != nil {
				return fmt.Errorf("failed to parse %s: %w", path, err)
			}
			codebase.Handlers = append(codebase.Handlers, handlers...)
		}

		return nil
	})

	return err
}

func (p *Parser) extractDjangoViewsFromFile(ctx context.Context, filePath string) ([]types.Handler, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	tree, err := p.treeSitterParser.ParseCtx(ctx, content, nil)
	if err != nil {
		return nil, err
	}

	handlers := []types.Handler{}

	// Look for function definitions
	p.traverseNode(tree.RootNode(), func(node *sitter.Node) {
		if node.Type() == "function_definition" {
			handler := p.extractDjangoView(node, content, filePath)
			if handler != nil {
				handlers = append(handlers, *handler)
			}
		}
	})

	return handlers, nil
}

func (p *Parser) extractDjangoView(node *sitter.Node, content []byte, filePath string) *types.Handler {
	// Django view pattern: def view_name(request, ...):
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}

	name := string(content[nameNode.StartByte():nameNode.EndByte()])

	// Check if first parameter is 'request'
	parametersNode := node.ChildByFieldName("parameters")
	if parametersNode == nil || parametersNode.ChildCount() == 0 {
		return nil
	}

	firstParam := parametersNode.Child(0)
	if firstParam.Type() == "identifier" {
		paramName := string(content[firstParam.StartByte():firstParam.EndByte()])
		if paramName != "request" && paramName != "self" {
			return nil
		}
	}

	line := p.getNodeLine(node, content)

	return &types.Handler{
		Name: name,
		File: filePath,
		Line: line,
	}
}

func (p *Parser) parseFlaskRoutes(ctx context.Context, rootDir string, codebase *types.Codebase) error {
	// Flask routes are typically in app.py or main.py with @app.route decorators
	// This is a stub implementation
	return nil
}

func (p *Parser) parseFastAPIRoutes(ctx context.Context, rootDir string, codebase *types.Codebase) error {
	// FastAPI routes use @app.get, @app.post decorators
	// This is a stub implementation
	return nil
}

func (p *Parser) traverseNode(node *sitter.Node, visit func(*sitter.Node)) {
	if node == nil {
		return
	}
	visit(node)
	for i := 0; i < node.ChildCount(); i++ {
		p.traverseNode(node.Child(i), visit)
	}
}

func (p *Parser) getNodeLine(node *sitter.Node, content []byte) int {
	if node == nil {
		return 0
	}
	return int(node.StartPoint().Row) + 1
}
