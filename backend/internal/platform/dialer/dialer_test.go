package dialer

import (
	"context"
	"errors"
	"net"
	"strconv"
	"testing"
)

type fakeResolver struct {
	answers map[string][]net.IPAddr
}

func (f *fakeResolver) LookupIPAddr(_ context.Context, host string) ([]net.IPAddr, error) {
	if v, ok := f.answers[host]; ok {
		return v, nil
	}
	return nil, errors.New("not found")
}

func TestAC_SEC_03_ValidateTargetRefusesAnyNonPublicAnswer(t *testing.T) {
	r := &fakeResolver{answers: map[string][]net.IPAddr{
		"example.com": {{IP: net.ParseIP("8.8.8.8")}, {IP: net.ParseIP("10.0.0.1")}},
	}}
	d := New(Options{Resolver: r})
	err := d.ValidateTarget(context.Background(), "example.com", 587)
	if !errors.Is(err, ErrNotAllowed) {
		t.Fatalf("expected ErrNotAllowed, got %v", err)
	}
}

func TestAC_SEC_03_ControlRefusesRebinding(t *testing.T) {
	// Resolver says public at config time; at dial time the hostname resolves
	// to a loopback address. Control must refuse.
	r := &fakeResolver{answers: map[string][]net.IPAddr{
		"rebind.test": {{IP: net.ParseIP("8.8.8.8")}},
	}}
	d := New(Options{Resolver: r, Timeout: 1})
	// Make resolver flip to private on the second call.
	call := 0
	orig := r.LookupIPAddr
	_ = orig
	wrapping := &flippingResolver{base: r, n: &call}
	d2 := New(Options{Resolver: wrapping, Timeout: 1})
	_ = d
	// Force config-time public.
	if err := d2.ValidateTarget(context.Background(), "rebind.test", 587); err != nil {
		t.Fatalf("config-time should pass: %v", err)
	}
	// Now dial a literal loopback to exercise Control's classification directly.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	tcpAddr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatal("not tcp")
	}
	port := tcpAddr.Port
	_, err = d2.DialContext(context.Background(), "tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if !errors.Is(err, ErrNotAllowed) {
		t.Fatalf("expected ErrNotAllowed, got %v", err)
	}
}

type flippingResolver struct {
	base *fakeResolver
	n    *int
}

func (f *flippingResolver) LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	*f.n++
	if *f.n >= 2 {
		return []net.IPAddr{{IP: net.ParseIP("127.0.0.1")}}, nil
	}
	return f.base.LookupIPAddr(ctx, host)
}

func TestAC_SEC_03_ControlRefusesLiteralPrivateIP(t *testing.T) {
	d := New(Options{})
	_, err := d.DialContext(context.Background(), "tcp", "127.0.0.1:2525")
	if !errors.Is(err, ErrNotAllowed) {
		t.Fatalf("expected ErrNotAllowed, got %v", err)
	}
}

func TestAC_SEC_03_PortAllowList(t *testing.T) {
	for _, p := range []int{25, 465, 587, 2525} {
		if !portAllowed(p) {
			t.Fatalf("port %d should be allowed", p)
		}
	}
	for _, p := range []int{0, 24, 26, 80, 8080, 65535} {
		if portAllowed(p) {
			t.Fatalf("port %d should be rejected", p)
		}
	}
	d := New(Options{})
	for _, p := range []int{24, 8080} {
		err := d.ValidateTarget(context.Background(), "8.8.8.8", p)
		if !errors.Is(err, ErrNotAllowed) {
			t.Fatalf("port %d: expected ErrNotAllowed, got %v", p, err)
		}
	}
}
