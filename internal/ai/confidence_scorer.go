package ai

import (
	"github.com/go-go-golems/gotest-agent/internal/parser/types"
	"strings"
)

// ConfidenceScorer scores test cases based on codebase complexity
type ConfidenceScorer struct{}

// ScoreTestCase calculates confidence score (0-100) for a test case
func (s *ConfidenceScorer) ScoreTestCase(test *TestCase, codebase *types.Codebase) int {
	score := 100

	// Reduce score for complex routes
	route := s.findRoute(test.Name, codebase.Routes)
	if route != nil {
		// Complex middleware chains
		if len(route.Middleware) > 3 {
			score -= 10
		}

		// Many route parameters
		paramCount := s.countRouteParams(route.Path)
		if paramCount > 1 {
			score -= 15
		}

		// Complex path patterns (regex, wildcards)
		if strings.Contains(route.Path, "*") || strings.Contains(route.Path, "?") {
			score -= 10
		}
	}

	// Reduce score for complex models
	model := s.findModel(test.Name, codebase.Models)
	if model != nil {
		// Many fields
		if len(model.Fields) > 10 {
			score -= 10
		}

		// Complex relationships
		if len(model.Relations) > 3 {
			score -= 15
		}

		// Many-to-many relationships
		for _, rel := range model.Relations {
			if rel.Type == "manyToMany" {
				score -= 10
				break
			}
		}
	}

	// Reduce score for complex handlers
	handler := s.findHandler(test.Name, codebase.Handlers)
	if handler != nil {
		// Many parameters
		if len(handler.Parameters) > 5 {
			score -= 10
		}

		// Complex return types
		if strings.Contains(handler.ReturnType, "map") || strings.Contains(handler.ReturnType, "interface") {
			score -= 5
		}
	}

	// Reduce score for integration/e2e tests (harder to verify)
	if test.Type == "integration" {
		score -= 10
	} else if test.Type == "e2e" {
		score -= 20
	}

	// Reduce score for security tests (complex to verify)
	if strings.Contains(strings.ToLower(test.Name), "security") ||
		strings.Contains(strings.ToLower(test.Name), "auth") {
		score -= 5
	}

	// Ensure score is within bounds
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// findRoute finds a route that matches the test name
func (s *ConfidenceScorer) findRoute(testName string, routes []types.Route) *types.Route {
	testNameLower := strings.ToLower(testName)
	for i := range routes {
		route := &routes[i]
		// Match if test name contains route path or handler name
		if strings.Contains(testNameLower, strings.ToLower(route.Path)) ||
			strings.Contains(testNameLower, strings.ToLower(route.Handler)) {
			return route
		}
	}
	return nil
}

// findModel finds a model that matches the test name
func (s *ConfidenceScorer) findModel(testName string, models []types.Model) *types.Model {
	testNameLower := strings.ToLower(testName)
	for i := range models {
		model := &models[i]
		if strings.Contains(testNameLower, strings.ToLower(model.Name)) {
			return model
		}
	}
	return nil
}

// findHandler finds a handler that matches the test name
func (s *ConfidenceScorer) findHandler(testName string, handlers []types.Handler) *types.Handler {
	testNameLower := strings.ToLower(testName)
	for i := range handlers {
		handler := &handlers[i]
		if strings.Contains(testNameLower, strings.ToLower(handler.Name)) {
			return handler
		}
	}
	return nil
}

// countRouteParams counts the number of parameters in a route path
// Examples: /users/:id -> 1, /users/:userId/posts/:postId -> 2
func (s *ConfidenceScorer) countRouteParams(path string) int {
	count := 0
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ":") || strings.HasPrefix(part, "{") {
			count++
		}
	}
	return count
}
