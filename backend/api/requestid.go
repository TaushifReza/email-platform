package api

import (
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	requestIDHeader = "X-Request-ID"
	requestIDKey    = "request_id"
)

var requestIDRe = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)

// RequestIDMiddleware ensures every request has a valid request id.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := strings.TrimSpace(c.GetHeader(requestIDHeader))
		if !requestIDRe.MatchString(id) {
			id = uuid.NewString()
		}
		c.Set(requestIDKey, id)
		c.Header(requestIDHeader, id)
		c.Next()
	}
}

// RequestIDFrom returns the request id stored in context.
func RequestIDFrom(c *gin.Context) string {
	if v, ok := c.Get(requestIDKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
