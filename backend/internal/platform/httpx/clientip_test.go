package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientIPIgnoresSpoofedForwardedFor_SEC12(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.5:1234"
	r.Header.Set("X-Forwarded-For", "1.2.3.4")
	got := ClientIP(r, "")
	if got != "10.0.0.5" {
		t.Fatalf("expected peer IP, got %s", got)
	}
}

func TestClientIPTrustedProxyOnly_AUD02(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.5:1234"
	r.Header.Set("X-Forwarded-For", "1.2.3.4, 10.0.0.5")
	got := ClientIP(r, "10.0.0.0/8")
	if got != "1.2.3.4" {
		t.Fatalf("expected client IP, got %s", got)
	}
}
