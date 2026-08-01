package auth

import (
	"context"
	"net/http"
	"strings"
)

// Role defines the access level for a user or API client.
type Role string

const (
	// RoleAdmin has full access to all resources.
	RoleAdmin Role = "admin"

	// RoleReviewer can review and approve/reject test plans, change proposals, and fix suggestions.
	RoleReviewer Role = "reviewer"

	// RoleViewer can view tests, results, runs, and dashboards but cannot modify.
	RoleViewer Role = "viewer"

	// RoleAPIClient is for machine-to-machine API access (CLI tools, CI/CD pipelines).
	// It can create runs, execute tests, and export results but cannot modify configuration.
	RoleAPIClient Role = "api_client"
)

// ValidRoles lists all known roles.
var ValidRoles = map[Role]bool{
	RoleAdmin:     true,
	RoleReviewer:  true,
	RoleViewer:    true,
	RoleAPIClient: true,
}

// RolePriority defines the access hierarchy. Higher numbers = more permissions.
var RolePriority = map[Role]int{
	RoleAdmin:     4,
	RoleReviewer:  3,
	RoleViewer:    2,
	RoleAPIClient: 1,
}

// CanManage returns true if the role can manage (create/update/delete) resources.
func (r Role) CanManage() bool {
	return r == RoleAdmin || r == RoleReviewer
}

// CanView returns true if the role can view resources.
func (r Role) CanView() bool {
	return true // all roles can view
}

// CanConfigure returns true if the role can change system settings.
func (r Role) CanConfigure() bool {
	return r == RoleAdmin
}

// CanRunTests returns true if the role can create and execute test runs.
func (r Role) CanRunTests() bool {
	return r == RoleAdmin || r == RoleReviewer || r == RoleAPIClient
}

// AtLeast returns true if the role is at least the given minimum role.
func (r Role) AtLeast(minRole Role) bool {
	return RolePriority[r] >= RolePriority[minRole]
}

// RequireRole is an HTTP middleware that enforces a minimum role level.
// It must be used AFTER authentication middleware (which sets claims in context).
func RequireRole(minRole Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r.Context())
			if claims == nil {
				writeRoleError(w, http.StatusUnauthorized, "authentication required")
				return
			}
			userRole := Role(claims.Role)
			if !userRole.AtLeast(minRole) {
				writeRoleError(w, http.StatusForbidden, "insufficient permissions: requires role "+string(minRole))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyRole is an HTTP middleware that allows any of the listed roles.
func RequireAnyRole(roles ...Role) func(http.Handler) http.Handler {
	allowed := make(map[Role]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r.Context())
			if claims == nil {
				writeRoleError(w, http.StatusUnauthorized, "authentication required")
				return
			}
			if !allowed[Role(claims.Role)] {
				list := make([]string, len(roles))
				for i, role := range roles {
					list[i] = string(role)
				}
				writeRoleError(w, http.StatusForbidden, "requires one of: "+strings.Join(list, ", "))
			}
			next.ServeHTTP(w, r)
		})
	}
}

// WithRole sets the role in a context. Used when constructing claims.
func WithRole(ctx context.Context, role Role) context.Context {
	claims := GetClaims(ctx)
	if claims != nil {
		claims.Role = string(role)
	}
	return ctx
}

func writeRoleError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"error":"` + msg + `"}`))
}
