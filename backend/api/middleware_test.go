package api

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSameOriginRequiresMatchingHostIncludingPort(t *testing.T) {
	assert.True(t, sameOrigin("http://localhost:8081", "localhost:8081"))
	assert.False(t, sameOrigin("http://localhost:8081", "localhost"))
}

func TestNginxPreservesHostHeaderForCSRF(t *testing.T) {
	data, err := os.ReadFile("../../infra/nginx/nginx.conf")
	require.NoError(t, err)
	config := string(data)
	assert.Contains(t, config, "proxy_set_header Host $http_host;")
	assert.NotContains(t, config, "proxy_set_header Host $host;")
	assert.Contains(t, config, `proxy_set_header Cf-Access-Authenticated-User-Email "";`)
	assert.Contains(t, config, "server_name admin.localhost admin;")
}
