package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDEV_DocsEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterDevRoutes(router, "../../api/openapi.yaml")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/docs", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "swagger-ui")

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "openapi: 3.0.3")
}

func TestDEV_DocsDisabledWithoutRegistration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/docs", nil))
	assert.Equal(t, http.StatusNotFound, rec.Code)
}
