package ai

import (
	"testing"

	"github.com/go-go-golems/gotest-agent/internal/parser/types"
)

func TestConfidenceScorer_ScoreTestCase(t *testing.T) {
	scorer := &ConfidenceScorer{}

	tests := []struct {
		name          string
		test          *TestCase
		codebase      *types.Codebase
		expectedRange [2]int // [min, max]
	}{
		{
			name: "simple unit test",
			test: &TestCase{
				Name: "test_get_user",
				Type: "unit",
			},
			codebase: &types.Codebase{
				Routes: []types.Route{
					{Path: "/users", Handler: "get_user"},
				},
				Models: []types.Model{
					{Name: "User", Fields: []types.Field{{Name: "id"}}},
				},
				Handlers: []types.Handler{
					{Name: "get_user", Parameters: []types.Parameter{{Name: "id"}}},
				},
			},
			expectedRange: [2]int{90, 100},
		},
		{
			name: "complex integration test with many middleware",
			test: &TestCase{
				Name: "test_admin_dashboard",
				Type: "integration",
			},
			codebase: &types.Codebase{
				Routes: []types.Route{
					{
						Path:    "/admin/dashboard",
						Handler: "admin_dashboard",
						Middleware: []string{
							"auth",
							"admin_only",
							"rate_limit",
							"log_access",
						},
					},
				},
				Models: []types.Model{
					{
						Name:   "User",
						Fields: make([]types.Field, 15), // Many fields
						Relations: []types.Relation{
							{Type: "belongsTo", RelatedModel: "Organization"},
							{Type: "hasMany", RelatedModel: "Post"},
							{Type: "hasMany", RelatedModel: "Comment"},
							{Type: "manyToMany", RelatedModel: "Role"},
						},
					},
				},
				Handlers: []types.Handler{
					{
						Name: "admin_dashboard",
						Parameters: []types.Parameter{
							{Name: "request"},
							{Name: "response"},
							{Name: "next"},
							{Name: "user"},
							{Name: "session"},
							{Name: "config"},
						},
						ReturnType: "map[string]interface{}",
					},
				},
			},
			expectedRange: [2]int{40, 70},
		},
		{
			name: "e2e security test",
			test: &TestCase{
				Name: "test_auth_flow_security",
				Type: "e2e",
			},
			codebase: &types.Codebase{
				Routes: []types.Route{
					{
						Path:    "/auth/login",
						Handler: "auth_login",
						Middleware: []string{
							"rate_limit",
						},
					},
				},
				Models:   []types.Model{},
				Handlers: []types.Handler{},
			},
			expectedRange: [2]int{50, 75},
		},
		{
			name: "test with complex route parameters",
			test: &TestCase{
				Name: "test_get_post_comments",
				Type: "unit",
			},
			codebase: &types.Codebase{
				Routes: []types.Route{
					{
						Path:       "/posts/:postId/comments/:commentId",
						Handler:    "get_post_comment",
						Middleware: []string{},
					},
				},
				Models:   []types.Model{},
				Handlers: []types.Handler{},
			},
			expectedRange: [2]int{75, 90},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.ScoreTestCase(tt.test, tt.codebase)

			if score < tt.expectedRange[0] || score > tt.expectedRange[1] {
				t.Errorf("expected score in range [%d, %d], got %d",
					tt.expectedRange[0], tt.expectedRange[1], score)
			}

			// Verify score is within valid bounds
			if score < 0 || score > 100 {
				t.Errorf("score %d is outside valid range [0, 100]", score)
			}
		})
	}
}

func TestConfidenceScorer_CountRouteParams(t *testing.T) {
	scorer := &ConfidenceScorer{}

	tests := []struct {
		path     string
		expected int
	}{
		{"/users", 0},
		{"/users/:id", 1},
		{"/posts/:postId/comments", 1},
		{"/posts/:postId/comments/:commentId", 2},
		{"/users/:userId/posts/:postId/comments/:commentId", 3},
		{"/", 0},
		{"/*", 0},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			count := scorer.countRouteParams(tt.path)
			if count != tt.expected {
				t.Errorf("expected %d params, got %d", tt.expected, count)
			}
		})
	}
}

func TestConfidenceScorer_FindRoute(t *testing.T) {
	scorer := &ConfidenceScorer{}

	routes := []types.Route{
		{Path: "/users", Handler: "get_users"},
		{Path: "/users/:id", Handler: "get_user"},
		{Path: "/posts", Handler: "get_posts"},
	}

	tests := []struct {
		testName string
		expected *types.Route
	}{
		{"test_get_users", &routes[0]},
		{"test_get_user_by_id", &routes[1]},
		{"test_get_posts_list", &routes[2]},
		{"test_nonexistent", nil},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result := scorer.findRoute(tt.testName, routes)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %+v", result)
				}
			} else {
				if result == nil {
					t.Errorf("expected route, got nil")
				} else if result.Path != tt.expected.Path {
					t.Errorf("expected path %s, got %s", tt.expected.Path, result.Path)
				}
			}
		})
	}
}

func TestConfidenceScorer_FindModel(t *testing.T) {
	scorer := &ConfidenceScorer{}

	models := []types.Model{
		{Name: "User", Fields: []types.Field{{Name: "id"}}},
		{Name: "Post", Fields: []types.Field{{Name: "id"}}},
	}

	tests := []struct {
		testName string
		expected *types.Model
	}{
		{"test_user_creation", &models[0]},
		{"test_create_post", &models[1]},
		{"test_nonexistent", nil},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result := scorer.findModel(tt.testName, models)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %+v", result)
				}
			} else {
				if result == nil {
					t.Errorf("expected model, got nil")
				} else if result.Name != tt.expected.Name {
					t.Errorf("expected name %s, got %s", tt.expected.Name, result.Name)
				}
			}
		})
	}
}

func TestConfidenceScorer_FindHandler(t *testing.T) {
	scorer := &ConfidenceScorer{}

	handlers := []types.Handler{
		{Name: "get_users", Parameters: []types.Parameter{{Name: "request"}}},
		{Name: "create_post", Parameters: []types.Parameter{{Name: "request"}}},
	}

	tests := []struct {
		testName string
		expected *types.Handler
	}{
		{"test_get_users_api", &handlers[0]},
		{"test_create_post_endpoint", &handlers[1]},
		{"test_nonexistent", nil},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result := scorer.findHandler(tt.testName, handlers)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %+v", result)
				}
			} else {
				if result == nil {
					t.Errorf("expected handler, got nil")
				} else if result.Name != tt.expected.Name {
					t.Errorf("expected name %s, got %s", tt.expected.Name, result.Name)
				}
			}
		})
	}
}

func TestConfidenceScorer_ScoreBounds(t *testing.T) {
	scorer := &ConfidenceScorer{}

	// Test that score never goes below 0 or above 100
	test := &TestCase{
		Name: "test_extreme_complexity",
		Type: "e2e",
	}

	codebase := &types.Codebase{
		Routes: []types.Route{
			{
				Path:    "/extremely/complex/:path/:with/:many/:params",
				Handler: "extremely_complex_handler",
				Middleware: []string{
					"auth",
					"admin",
					"rate_limit",
					"log",
					"cache",
					"compress",
				},
			},
		},
		Models: []types.Model{
			{
				Name:   "ComplexModel",
				Fields: make([]types.Field, 20),
				Relations: []types.Relation{
					{Type: "manyToMany", RelatedModel: "A"},
					{Type: "manyToMany", RelatedModel: "B"},
					{Type: "manyToMany", RelatedModel: "C"},
					{Type: "manyToMany", RelatedModel: "D"},
					{Type: "manyToMany", RelatedModel: "E"},
				},
			},
		},
		Handlers: []types.Handler{
			{
				Name: "extremely_complex_handler",
				Parameters: make([]types.Parameter, 10),
				ReturnType: "map[string]interface{}",
			},
		},
	}

	score := scorer.ScoreTestCase(test, codebase)

	if score < 0 {
		t.Errorf("score %d is below minimum 0", score)
	}
	if score > 100 {
		t.Errorf("score %d is above maximum 100", score)
	}

	t.Logf("Extreme complexity test scored: %d", score)
}
