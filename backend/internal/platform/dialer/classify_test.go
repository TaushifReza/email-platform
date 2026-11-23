package dialer

import (
	"net"
	"testing"
)

func TestPRV04_ClassifyIPv4(t *testing.T) {
	bad := []string{
		"0.0.0.0", "10.0.0.1", "127.0.0.1", "169.254.169.254",
		"172.16.0.1", "192.168.1.1", "100.64.0.1", "192.0.0.1",
		"198.18.0.1", "224.0.0.1", "240.0.0.1", "255.255.255.255",
		"192.0.2.1", "198.51.100.1", "203.0.113.1",
	}
	for _, s := range bad {
		if IsPublicIP(net.ParseIP(s)) {
			t.Fatalf("expected %s non-public", s)
		}
	}
}

func TestPRV04_ClassifyIPv6(t *testing.T) {
	bad := []string{
		"::1", "fe80::1", "fc00::1", "ff02::1", "::",
		"2001:db8::1", "64:ff9b::a00:1", "::ffff:10.0.0.1", "::ffff:127.0.0.1",
	}
	for _, s := range bad {
		if IsPublicIP(net.ParseIP(s)) {
			t.Fatalf("expected %s non-public", s)
		}
	}
}

func TestPRV04_AcceptsPublicAddresses(t *testing.T) {
	good := []string{"8.8.8.8", "1.1.1.1", "2606:4700:4700::1111"}
	for _, s := range good {
		if !IsPublicIP(net.ParseIP(s)) {
			t.Fatalf("expected %s public", s)
		}
	}
}

func TestPRV04_GenericErrorMessage(t *testing.T) {
	if ErrNotAllowed.Error() != "host not allowed" {
		t.Fatalf("wrong message: %s", ErrNotAllowed.Error())
	}
}
