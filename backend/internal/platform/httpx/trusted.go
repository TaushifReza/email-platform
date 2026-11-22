package httpx

import "net/http"

// IsTrustedPeer reports whether the immediate network peer is in trustedCIDRs.
// It deliberately ignores X-Forwarded-For: that header is client-controlled and
// must not be used to raise the trust level of the operator identity header.
func IsTrustedPeer(r *http.Request, trustedCIDRs string) bool {
	return ipInCIDRs(remoteIP(r.RemoteAddr), parseCIDRs(trustedCIDRs))
}
