package httpx

import (
	"net"
	"net/http"
	"strings"
)

// ClientIP resolves the client IP for r. Trusted proxy headers are only
// honored when the immediate peer is in trustedCIDRs.
func ClientIP(r *http.Request, trustedCIDRs string) string {
	peer := remoteIP(r.RemoteAddr)
	trusted := parseCIDRs(trustedCIDRs)
	if !ipInCIDRs(peer, trusted) {
		return peer
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		for i := len(parts) - 1; i >= 0; i-- {
			candidate := strings.TrimSpace(parts[i])
			if candidate == "" {
				continue
			}
			ip := parseIP(candidate)
			if ip == nil {
				continue
			}
			if !ipInCIDRs(ip.String(), trusted) {
				return ip.String()
			}
		}
	}
	if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); xri != "" {
		if ip := parseIP(xri); ip != nil {
			return ip.String()
		}
	}
	return peer
}

func remoteIP(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}

func parseIP(s string) net.IP {
	return net.ParseIP(s)
}

func parseCIDRs(s string) []net.IPNet {
	var out []net.IPNet
	for _, c := range strings.Split(s, ",") {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		n, ok := parseCIDR(c)
		if ok {
			out = append(out, n)
		}
	}
	return out
}

func parseCIDR(c string) (net.IPNet, bool) {
	if strings.Contains(c, "/") {
		_, n, err := net.ParseCIDR(c)
		if err != nil {
			return net.IPNet{}, false
		}
		return *n, true
	}
	ip := net.ParseIP(c)
	if ip == nil {
		return net.IPNet{}, false
	}
	if ip.To4() != nil {
		return net.IPNet{IP: ip, Mask: net.CIDRMask(32, 32)}, true
	}
	return net.IPNet{IP: ip, Mask: net.CIDRMask(128, 128)}, true
}

func ipInCIDRs(ip string, nets []net.IPNet) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	for _, n := range nets {
		if n.Contains(parsed) {
			return true
		}
	}
	return false
}
