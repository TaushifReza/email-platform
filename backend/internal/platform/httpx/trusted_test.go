package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsTrustedPeerUsesRemoteAddrOnly(t *testing.T) {
	tests := []struct {
		name    string
		peer    string
		cidrs   string
		xff     string
		trusted bool
	}{
		{name: "trusted peer", peer: "172.18.0.5:1234", cidrs: "172.16.0.0/12", trusted: true},
		{
			name: "untrusted peer spoofing xff", peer: "203.0.113.9:1234",
			cidrs: "172.16.0.0/12", xff: "172.18.0.5", trusted: false,
		},
		{name: "no trusted cidrs", peer: "172.18.0.5:1234", cidrs: "", trusted: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/operator/session", nil)
			r.RemoteAddr = tt.peer
			if tt.xff != "" {
				r.Header.Set("X-Forwarded-For", tt.xff)
			}
			if got := IsTrustedPeer(r, tt.cidrs); got != tt.trusted {
				t.Fatalf("IsTrustedPeer() = %v, want %v", got, tt.trusted)
			}
		})
	}
}
