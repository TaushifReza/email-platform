package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	CodeInvalidRequest     ProblemCode = "invalid_request"
	CodeInvalidCredentials ProblemCode = "invalid_credentials" // #nosec G101
	CodeRateLimited        ProblemCode = "rate_limited"
	CodeUnavailable        ProblemCode = "unavailable"
	CodeInternalError      ProblemCode = "internal_error"
	CodeMFARequired        ProblemCode = "mfa_required"
	CodeInvalidCode        ProblemCode = "invalid_code"
	CodeLocked             ProblemCode = "locked"
	CodeCSRFRejected       ProblemCode = "csrf_rejected"
	CodeForbidden          ProblemCode = "forbidden"
	CodeNotFound           ProblemCode = "not_found"
	CodeConflict           ProblemCode = "conflict"
	CodeValidationFailed   ProblemCode = "validation_failed"
	CodeUnauthorized       ProblemCode = "unauthorized"
)

var problemTitles = map[ProblemCode]string{ //nolint:gochecknoglobals
	CodeInvalidRequest:     "Invalid request",
	CodeInvalidCredentials: "Invalid credentials",
	CodeRateLimited:        "Too many requests",
	CodeUnavailable:        "Service unavailable",
	CodeInternalError:      "Internal server error",
	CodeMFARequired:        "MFA required",
	CodeInvalidCode:        "Invalid code",
	CodeLocked:             "Account locked",
	CodeCSRFRejected:       "CSRF rejected",
	CodeForbidden:          "Forbidden",
	CodeNotFound:           "Not found",
	CodeConflict:           "Conflict",
	CodeValidationFailed:   "Validation failed",
	CodeUnauthorized:       "Unauthorized",
}

// RespondProblem writes a problem details response.
func RespondProblem(c *gin.Context, status int, code ProblemCode, detail string, fieldErrors []FieldError) {
	detailPtr := &detail
	instance := c.Request.URL.Path
	var errorsPtr *[]FieldError
	if len(fieldErrors) > 0 {
		errorsPtr = &fieldErrors
	}
	c.Header("Content-Type", "application/problem+json")
	c.JSON(status, Problem{
		Type:      "about:blank",
		Title:     problemTitles[code],
		Status:    status,
		Detail:    detailPtr,
		Instance:  &instance,
		Code:      code,
		RequestId: RequestIDFrom(c),
		Errors:    errorsPtr,
	})
}

// RespondValidation writes a 422 problem with field errors.
func RespondValidation(c *gin.Context, fieldErrors []FieldError) {
	RespondProblem(c, http.StatusUnprocessableEntity, CodeValidationFailed, "One or more fields are invalid", fieldErrors)
}

// RespondNoContent writes 204.
func RespondNoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}
