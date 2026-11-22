package api

import (
	"errors"
	"net/http"

	"github.com/TaushifReza/email-platform/backend/internal/identity"
	"github.com/TaushifReza/email-platform/backend/internal/platform/httpx"
	"github.com/gin-gonic/gin"
)

func (s *Server) ListUsers(c *gin.Context) {
	actor, ok := ActorFrom(c)
	if !ok {
		RespondProblem(c, http.StatusUnauthorized, CodeUnauthorized, "Unauthenticated", nil)
		return
	}
	users, err := s.users.List(c.Request.Context(), actor)
	if err != nil {
		RespondProblem(c, http.StatusForbidden, CodeForbidden, "Forbidden", nil)
		return
	}
	data := make([]User, 0, len(users))
	for _, u := range users {
		uid := u.ID
		email := u.Email
		name := u.DisplayName
		role := u.Role
		status := u.Status
		data = append(data, User{Id: &uid, Email: &email, DisplayName: &name, Role: &role, Status: &status})
	}
	c.JSON(http.StatusOK, UserList{Data: &data, Meta: &struct {
		HasMore    *bool   `json:"hasMore,omitempty"`
		NextCursor *string `json:"nextCursor,omitempty"`
	}{HasMore: boolPtr(false)}})
}

func (s *Server) InviteUser(c *gin.Context) {
	actor, ok := ActorFrom(c)
	if !ok {
		RespondProblem(c, http.StatusUnauthorized, CodeUnauthorized, "Unauthenticated", nil)
		return
	}
	var req InviteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" || req.Role == "" {
		RespondProblem(c, http.StatusBadRequest, CodeInvalidRequest, "Invalid request body", nil)
		return
	}
	info := identity.ClientInfo{IP: httpx.ClientIP(c.Request, s.cfg.TrustedProxyCIDRs), UserAgent: c.Request.UserAgent()}
	if err := s.users.Invite(c.Request.Context(), actor, req.Email, req.Role, info); err != nil {
		RespondProblem(c, http.StatusForbidden, CodeForbidden, "Forbidden", nil)
		return
	}
	RespondNoContent(c)
}

func (s *Server) UpdateUser(c *gin.Context, userID string) {
	actor, ok := ActorFrom(c)
	if !ok {
		RespondProblem(c, http.StatusUnauthorized, CodeUnauthorized, "Unauthenticated", nil)
		return
	}
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondProblem(c, http.StatusBadRequest, CodeInvalidRequest, "Invalid request body", nil)
		return
	}
	change := identity.UserChange{}
	if req.Role != nil {
		change.Role = *req.Role
	}
	if req.Deactivate != nil && *req.Deactivate {
		change.Deactivate = true
	}
	info := identity.ClientInfo{IP: httpx.ClientIP(c.Request, s.cfg.TrustedProxyCIDRs), UserAgent: c.Request.UserAgent()}
	if err := s.users.Update(c.Request.Context(), actor, userID, change, info); err != nil {
		if errors.Is(err, identity.ErrForbidden) {
			RespondProblem(c, http.StatusForbidden, CodeForbidden, "Forbidden", nil)
			return
		}
		RespondProblem(c, http.StatusConflict, CodeConflict, err.Error(), nil)
		return
	}
	RespondNoContent(c)
}

func boolPtr(b bool) *bool { return &b }
