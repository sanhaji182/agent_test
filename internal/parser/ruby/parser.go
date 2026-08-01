package ruby

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-go-golems/gotest-agent/internal/parser/types"
)

// Parser is a regex-based parser for Ruby on Rails projects.
// It does not require tree-sitter and instead scans files with regular expressions
// to extract routes, models, and controller handlers.
type Parser struct{}

// NewParser creates a new Ruby parser instance.
func NewParser() *Parser {
	return &Parser{}
}

// SupportedLanguages returns the list of languages this parser handles.
func (p *Parser) SupportedLanguages() []string {
	return []string{"ruby"}
}

// DetectFramework detects the Ruby web framework used in the project.
func (p *Parser) DetectFramework(rootDir string) (string, error) {
	return p.detectFramework(rootDir), nil
}

// detectFramework checks for Rails markers: Gemfile with 'rails' gem,
// config/routes.rb, or config/application.rb.
func (p *Parser) detectFramework(rootDir string) string {
	gemfilePath := filepath.Join(rootDir, "Gemfile")
	if content, err := os.ReadFile(gemfilePath); err == nil {
		gemfile := string(content)
		if strings.Contains(gemfile, "gem 'rails'") || strings.Contains(gemfile, `gem "rails"`) {
			return "rails"
		}
	}

	routesPath := filepath.Join(rootDir, "config", "routes.rb")
	if _, err := os.Stat(routesPath); err == nil {
		return "rails"
	}

	appConfig := filepath.Join(rootDir, "config", "application.rb")
	if _, err := os.Stat(appConfig); err == nil {
		return "rails"
	}

	return "ruby"
}

// Parse analyzes a Ruby/Rails codebase and extracts routes, models, and handlers.
func (p *Parser) Parse(ctx context.Context, rootDir string) (*types.Codebase, error) {
	codebase := &types.Codebase{
		Language:  "ruby",
		Framework: p.detectFramework(rootDir),
		Routes:    make([]types.Route, 0),
		Models:    make([]types.Model, 0),
		Handlers:  make([]types.Handler, 0),
	}

	if codebase.Framework == "rails" {
		if err := p.parseRailsRoutes(rootDir, codebase); err != nil {
			return nil, fmt.Errorf("failed to parse Rails routes: %w", err)
		}
		if err := p.parseRailsModels(rootDir, codebase); err != nil {
			return nil, fmt.Errorf("failed to parse Rails models: %w", err)
		}
		if err := p.parseRailsControllers(rootDir, codebase); err != nil {
			return nil, fmt.Errorf("failed to parse Rails controllers: %w", err)
		}
	}

	return codebase, nil
}

// ---------------------------------------------------------------------------
// Regex patterns for route parsing
// ---------------------------------------------------------------------------

var (
	// Matches HTTP verb routes: get '/path', to: 'controller#action'
	routeVerbRe = regexp.MustCompile(`^\s*(get|post|put|patch|delete|match)\s+['"]([^'"]+)['"]`)

	// Extracts the to: option value
	routeToRe = regexp.MustCompile(`to:\s*['"]([^'"]+)['"]`)

	// Extracts via: array for match routes
	routeViaRe = regexp.MustCompile(`via:\s*\[([^\]]+)\]`)

	// Matches resources :model_name
	resourcesRe = regexp.MustCompile(`^\s*resources\s+:(\w+)`)

	// Extracts only: [:index, :show] from resources
	resourcesOnlyRe = regexp.MustCompile(`only:\s*\[([^\]]+)\]`)

	// Matches namespace :name do
	namespaceRe = regexp.MustCompile(`^\s*namespace\s+:(\w+)\s+do\b`)

	// Matches scope '/prefix' do
	scopePathRe = regexp.MustCompile(`^\s*scope\s+['"]([^'"]+)['"]\s+do\b`)
)

// ---------------------------------------------------------------------------
// Route parsing
// ---------------------------------------------------------------------------

// parseRailsRoutes reads config/routes.rb and extracts route definitions.
func (p *Parser) parseRailsRoutes(rootDir string, codebase *types.Codebase) error {
	routesPath := filepath.Join(rootDir, "config", "routes.rb")
	content, err := os.ReadFile(routesPath)
	if err != nil {
		return nil // Missing routes file is not an error
	}

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	var prefixes []string // stack of namespace/scope path prefixes
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip blank lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Handle end of block
		if strings.HasPrefix(trimmed, "end") {
			if len(prefixes) > 0 {
				prefixes = prefixes[:len(prefixes)-1]
			}
			continue
		}

		// Namespace: push prefix onto stack
		if m := namespaceRe.FindStringSubmatch(line); m != nil {
			prefixes = append(prefixes, "/"+m[1])
			continue
		}

		// Scope with path: push prefix onto stack
		if m := scopePathRe.FindStringSubmatch(line); m != nil {
			prefixes = append(prefixes, m[1])
			continue
		}

		// Build the cumulative path prefix
		prefix := strings.Join(prefixes, "")

		// Resources: expand to standard REST routes
		if m := resourcesRe.FindStringSubmatch(line); m != nil {
			resourceName := m[1]
			only := extractOnlyOption(line)

			for _, route := range expandResourceRoutes(prefix, resourceName, only) {
				route.File = routesPath
				route.Line = lineNum
				codebase.Routes = append(codebase.Routes, route)
			}
			continue
		}

		// HTTP verb routes: get/post/put/patch/delete/match
		if m := routeVerbRe.FindStringSubmatch(line); m != nil {
			verb := m[1]
			path := m[2]

			// Determine HTTP method(s)
			httpMethods := []string{verbToHTTPMethod(verb)}
			if verb == "match" {
				if v := routeViaRe.FindStringSubmatch(line); v != nil {
					httpMethods = parseViaOption(v[1])
				}
			}

			// Extract handler from to: option
			handler := ""
			if h := routeToRe.FindStringSubmatch(line); h != nil {
				handler = h[1]
			}

			for _, httpMethod := range httpMethods {
				if httpMethod == "" {
					continue
				}
				codebase.Routes = append(codebase.Routes, types.Route{
					Method:  httpMethod,
					Path:    prefix + path,
					Handler: handler,
					File:    routesPath,
					Line:    lineNum,
				})
			}
			continue
		}

		// Root route: root to: 'controller#action'
		if strings.HasPrefix(trimmed, "root ") {
			if h := routeToRe.FindStringSubmatch(line); h != nil {
				codebase.Routes = append(codebase.Routes, types.Route{
					Method:  "GET",
					Path:    "/",
					Handler: h[1],
					File:    routesPath,
					Line:    lineNum,
				})
			}
		}
	}

	return scanner.Err()
}

// verbToHTTPMethod maps Rails DSL verbs to HTTP methods.
func verbToHTTPMethod(verb string) string {
	switch strings.ToLower(verb) {
	case "get":
		return "GET"
	case "post":
		return "POST"
	case "put":
		return "PUT"
	case "patch":
		return "PATCH"
	case "delete":
		return "DELETE"
	default:
		return strings.ToUpper(verb)
	}
}

// parseViaOption extracts HTTP methods from via: [:get, :post, ...].
func parseViaOption(via string) []string {
	parts := strings.Split(via, ",")
	methods := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.TrimPrefix(p, ":")
		p = strings.Trim(p, `'"`)
		p = strings.ToUpper(p)
		if p != "" {
			methods = append(methods, p)
		}
	}
	return methods
}

// extractOnlyOption parses the only: [:action1, :action2] clause from a resources line.
func extractOnlyOption(line string) map[string]bool {
	m := resourcesOnlyRe.FindStringSubmatch(line)
	if m == nil {
		return nil
	}
	only := make(map[string]bool)
	parts := strings.Split(m[1], ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.TrimPrefix(p, ":")
		if p != "" {
			only[p] = true
		}
	}
	return only
}

// expandResourceRoutes generates standard Rails REST routes for a resource.
// The 'only' map keys are bare action names (e.g., "index", "show"), matching the
// Rails DSL convention: resources :posts, only: [:index, :show].
func expandResourceRoutes(prefix, name string, only map[string]bool) []types.Route {
	all := []struct{ method, path, action string }{
		{"GET", "/" + name, name + "#index"},
		{"GET", "/" + name + "/new", name + "#new"},
		{"POST", "/" + name, name + "#create"},
		{"GET", "/" + name + "/:id", name + "#show"},
		{"GET", "/" + name + "/:id/edit", name + "#edit"},
		{"PATCH", "/" + name + "/:id", name + "#update"},
		{"PUT", "/" + name + "/:id", name + "#update"},
		{"DELETE", "/" + name + "/:id", name + "#destroy"},
	}

	routes := make([]types.Route, 0, len(all))
	for _, r := range all {
		if only != nil {
			// Extract bare action name from "resource#action" format
			actionParts := strings.SplitN(r.action, "#", 2)
			actionName := actionParts[len(actionParts)-1]
			if !only[actionName] {
				continue
			}
		}
		routes = append(routes, types.Route{
			Method:  r.method,
			Path:    prefix + r.path,
			Handler: r.action,
		})
	}
	return routes
}

// ---------------------------------------------------------------------------
// Regex patterns for model parsing
// ---------------------------------------------------------------------------

var (
	// Matches class ModelName < ApplicationRecord
	modelClassRe = regexp.MustCompile(`^\s*class\s+(\w+)\s*<\s*ApplicationRecord`)

	// ActiveRecord association macros
	belongsToRe           = regexp.MustCompile(`belongs_to\s+:(\w+)`)
	hasManyRe             = regexp.MustCompile(`has_many\s+:(\w+)`)
	hasOneRe              = regexp.MustCompile(`has_one\s+:(\w+)`)
	hasAndBelongsToManyRe = regexp.MustCompile(`has_and_belongs_to_many\s+:(\w+)`)

	// Validates macros
	validatesRe = regexp.MustCompile(`validates\s+:(\w+),\s*(.+)$`)
)

// ---------------------------------------------------------------------------
// Model parsing
// ---------------------------------------------------------------------------

// parseRailsModels walks app/models/*.rb and extracts model definitions.
func (p *Parser) parseRailsModels(rootDir string, codebase *types.Codebase) error {
	modelsDir := filepath.Join(rootDir, "app", "models")
	if _, err := os.Stat(modelsDir); os.IsNotExist(err) {
		return nil
	}

	return filepath.Walk(modelsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".rb") {
			return nil
		}
		relPath, _ := filepath.Rel(rootDir, path)

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		model := p.extractModel(string(content), relPath)
		if model != nil {
			codebase.Models = append(codebase.Models, *model)
		}
		return nil
	})
}

// extractModel extracts a single model from file content.
func (p *Parser) extractModel(content, filePath string) *types.Model {
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0

	var model *types.Model

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if m := modelClassRe.FindStringSubmatch(line); m != nil {
			model = &types.Model{
				Name:       m[1],
				File:       filePath,
				Line:       lineNum,
				Fields:     make([]types.Field, 0),
				Relations:  make([]types.Relation, 0),
				Validation: make([]types.ValidationRule, 0),
			}
			continue
		}

		if model == nil {
			continue
		}

		// belongs_to :other
		if m := belongsToRe.FindStringSubmatch(line); m != nil {
			model.Relations = append(model.Relations, types.Relation{
				Type:         "belongsTo",
				RelatedModel: m[1],
			})
		}

		// has_many :others
		if m := hasManyRe.FindStringSubmatch(line); m != nil {
			model.Relations = append(model.Relations, types.Relation{
				Type:         "hasMany",
				RelatedModel: m[1],
			})
		}

		// has_one :other
		if m := hasOneRe.FindStringSubmatch(line); m != nil {
			model.Relations = append(model.Relations, types.Relation{
				Type:         "hasOne",
				RelatedModel: m[1],
			})
		}

		// has_and_belongs_to_many :others
		if m := hasAndBelongsToManyRe.FindStringSubmatch(line); m != nil {
			model.Relations = append(model.Relations, types.Relation{
				Type:         "hasAndBelongsToMany",
				RelatedModel: m[1],
			})
		}

		// validates :field, presence: true, length: { minimum: 3 }
		if m := validatesRe.FindStringSubmatch(line); m != nil {
			fieldName := m[1]
			rest := m[2]
			// A validates line may contain multiple validation rules separated by commas
			rules := splitValidationRules(rest)
			for _, rule := range rules {
				model.Validation = append(model.Validation, types.ValidationRule{
					Field: fieldName,
					Rule:  strings.TrimSpace(rule),
				})
			}
		}
	}

	return model
}

// splitValidationRules splits the remainder of a validates line into individual rules.
// e.g. "presence: true, length: { minimum: 3 }" → ["presence: true", "length: { minimum: 3 }"]
func splitValidationRules(s string) []string {
	var rules []string
	depth := 0
	start := 0
	for i, ch := range s {
		switch ch {
		case '{':
			depth++
		case '}':
			depth--
		case ',':
			if depth == 0 {
				rules = append(rules, s[start:i])
				start = i + 1
			}
		}
	}
	// Append the last rule
	last := strings.TrimSpace(s[start:])
	if last != "" {
		rules = append(rules, last)
	}
	return rules
}

// ---------------------------------------------------------------------------
// Regex patterns for controller parsing
// ---------------------------------------------------------------------------

var (
	// Matches class UsersController < ApplicationController or ActionController::Base
	controllerClassRe = regexp.MustCompile(`^\s*class\s+(\w+Controller)\s*<\s*(ApplicationController|ActionController::Base|ActionController)`)
	// Matches a method definition
	methodDefRe = regexp.MustCompile(`^\s*def\s+(\w+)`)
)

// ---------------------------------------------------------------------------
// Controller parsing
// ---------------------------------------------------------------------------

// parseRailsControllers walks app/controllers/*.rb and extracts handler methods.
func (p *Parser) parseRailsControllers(rootDir string, codebase *types.Codebase) error {
	controllersDir := filepath.Join(rootDir, "app", "controllers")
	if _, err := os.Stat(controllersDir); os.IsNotExist(err) {
		return nil
	}

	return filepath.Walk(controllersDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".rb") {
			return nil
		}
		relPath, _ := filepath.Rel(rootDir, path)

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		handlers := p.extractHandlers(string(content), relPath)
		codebase.Handlers = append(codebase.Handlers, handlers...)
		return nil
	})
}

// extractHandlers extracts public action methods from a controller file.
func (p *Parser) extractHandlers(content, filePath string) []types.Handler {
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0

	var controllerName string
	isPrivate := false
	var handlers []types.Handler

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Detect class declaration
		if m := controllerClassRe.FindStringSubmatch(line); m != nil {
			controllerName = m[1]
			isPrivate = false
			continue
		}

		// Track visibility
		if trimmed == "private" || trimmed == "protected" {
			isPrivate = true
			continue
		}

		// end of class resets state
		if trimmed == "end" && controllerName != "" {
			// Only reset visibility on outer class end (simple heuristic)
			// We don't track nesting depth, so keep it simple
			continue
		}

		// Skip non-public methods
		if isPrivate {
			continue
		}

		// Detect method definitions
		if m := methodDefRe.FindStringSubmatch(line); m != nil {
			methodName := m[1]
			handlers = append(handlers, types.Handler{
				Name:       methodName,
				Controller: controllerName,
				File:       filePath,
				Line:       lineNum,
			})
		}
	}

	return handlers
}
