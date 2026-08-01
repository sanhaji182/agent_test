package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-go-golems/gotest-agent/internal/audit"
	"github.com/go-go-golems/gotest-agent/internal/auth"
)

// recordAudit is a convenience wrapper that extracts actor info from JWT claims
// and records an audit entry. It is safe to call even when no session is present
// (e.g., for unauthenticated API key access).
func (s *Server) recordAudit(r *http.Request, action audit.Action, resource audit.Resource, resourceID string, detail string) {
	claims := auth.GetClaims(r.Context())
	actorID := "system"
	actorRole := "admin"
	if claims != nil {
		if claims.Email != "" {
			actorID = claims.Email
		}
		if claims.Role != "" {
			actorRole = claims.Role
		}
	}
	s.auditLog.Record(actorID, actorRole, action, resource, resourceID, detail)
}

// handleListAuditLog returns recent audit log entries for admin review.
// GET /api/v1/audit-log?limit=50
func (s *Server) handleListAuditLog(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil || !auth.Role(claims.Role).AtLeast(auth.RoleAdmin) {
		writeJSONError(w, http.StatusForbidden, "admin role required")
		return
	}

	limit := 50
	entries := s.auditLog.List(limit)
	if entries == nil {
		entries = []audit.Entry{}
	}
	writeJSON(w, http.StatusOK, entries)
}

// handleListAuditLogByActor returns audit entries for a specific user.
// GET /api/v1/audit-log/users/{actorID}?limit=50
func (s *Server) handleListAuditLogByActor(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil || !auth.Role(claims.Role).AtLeast(auth.RoleAdmin) {
		writeJSONError(w, http.StatusForbidden, "admin role required")
		return
	}

	actorID := chi.URLParam(r, "actorID")
	if actorID == "" {
		writeJSONError(w, http.StatusBadRequest, "actor id required")
		return
	}

	limit := 50
	entries := s.auditLog.ListByActor(actorID, limit)
	if entries == nil {
		entries = []audit.Entry{}
	}
	writeJSON(w, http.StatusOK, entries)
}

// handleListAuditLogByResource returns audit entries for a specific resource.
// GET /api/v1/audit-log/{resource}/{resourceID}?limit=50
func (s *Server) handleListAuditLogByResource(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil || !auth.Role(claims.Role).AtLeast(auth.RoleAdmin) {
		writeJSONError(w, http.StatusForbidden, "admin role required")
		return
	}

	resource := audit.Resource(chi.URLParam(r, "resource"))
	resourceID := chi.URLParam(r, "resourceID")
	if resource == "" || resourceID == "" {
		writeJSONError(w, http.StatusBadRequest, "resource and resource_id required")
		return
	}

	limit := 50
	entries := s.auditLog.ListByResource(resource, resourceID, limit)
	if entries == nil {
		entries = []audit.Entry{}
	}
	writeJSON(w, http.StatusOK, entries)
}
