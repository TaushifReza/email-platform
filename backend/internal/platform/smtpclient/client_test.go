package smtpclient

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/TaushifReza/email-platform/tools/fakesmtp"
)

func startFake(t *testing.T, script *fakesmtp.Script) *fakesmtp.Server {
	t.Helper()
	srv, err := fakesmtp.New(fakesmtp.Options{Script: script})
	if err != nil {
		t.Fatal(err)
	}
	if err := srv.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = srv.Shutdown() }) //nolint:errcheck
	return srv
}

func rootPool(t *testing.T, srv *fakesmtp.Server) *x509.CertPool {
	t.Helper()
	pool := x509.NewCertPool()
	block, _ := pem.Decode(srv.CACertPEM())
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	pool.AddCert(cert)
	return pool
}

func TestAC_SEC_04_PlaintextRefused(t *testing.T) {
	if _, err := New(Options{Host: "h", Port: 25, Mode: "none", Dialer: &net.Dialer{}}); !errors.Is(err, ErrPlaintext) {
		t.Fatalf("expected ErrPlaintext, got %v", err)
	}
	if _, err := New(Options{Host: "h", Port: 25, Mode: "", Dialer: &net.Dialer{}}); !errors.Is(err, ErrPlaintext) {
		t.Fatalf("expected ErrPlaintext, got %v", err)
	}
}

func TestAC_SEC_03_DialerRequired(t *testing.T) {
	if _, err := New(Options{Host: "h", Port: 25, Mode: ModeSTARTTLS}); err == nil {
		t.Fatal("expected error for nil dialer")
	}
}

func TestAC_SEC_04_StartTLSRequired(t *testing.T) {
	srv := startFake(t, &fakesmtp.Script{TLS: fakesmtp.TLSNone})
	host, port := splitAddr(t, srv.SMTPAddr())
	c, err := New(Options{Host: host, Port: port, Mode: ModeSTARTTLS, Dialer: &net.Dialer{}, Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	r := c.Send(context.Background(), Envelope{From: "a@x.com", To: []string{"b@x.com"}, Data: []byte("hi")})
	if r.Class != TemporaryConnection {
		t.Fatalf("expected TemporaryConnection, got %v", r.Class)
	}
}

func TestAC_SEC_04_ImplicitTLS(t *testing.T) {
	srv := startFake(t, &fakesmtp.Script{TLS: fakesmtp.TLSImplicit})
	pool := rootPool(t, srv)
	host, port := splitAddr(t, srv.SMTPAddr())
	c, err := New(Options{Host: host, Port: port, Mode: ModeImplicit, Dialer: &net.Dialer{}, RootCAs: pool, Timeout: 5 * time.Second}) //nolint:lll
	if err != nil {
		t.Fatal(err)
	}
	r := c.Send(context.Background(), Envelope{From: "a@x.com", To: []string{"b@x.com"}, Data: []byte("hi\r\n")})
	if r.Class != Accepted {
		t.Fatalf("expected Accepted, got %v (%v)", r.Class, r.Err)
	}
}

func TestAC_SEC_04_CertificateVerified(t *testing.T) {
	srv := startFake(t, &fakesmtp.Script{TLS: fakesmtp.TLSImplicit})
	host, port := splitAddr(t, srv.SMTPAddr())
	c, err := New(Options{Host: host, Port: port, Mode: ModeImplicit, Dialer: &net.Dialer{}, Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	r := c.Send(context.Background(), Envelope{From: "a@x.com", To: []string{"b@x.com"}, Data: []byte("hi\r\n")})
	if r.Class == Accepted {
		t.Fatal("expected failure without trusted root")
	}
}

func TestAC_SEC_02_CredentialsNeverInResult(t *testing.T) {
	srv := startFake(t, &fakesmtp.Script{
		TLS:       fakesmtp.TLSStartTLS,
		Behaviors: []fakesmtp.Behavior{{Command: "auth", ReplyCode: 535, ReplyText: "bad"}},
	})
	pool := rootPool(t, srv)
	host, port := splitAddr(t, srv.SMTPAddr())
	c, err := New(Options{Host: host, Port: port, Mode: ModeSTARTTLS, Username: "u", Password: "SuperSecret123!", Dialer: &net.Dialer{}, RootCAs: pool, Timeout: 5 * time.Second}) //nolint:lll
	if err != nil {
		t.Fatal(err)
	}
	r := c.Send(context.Background(), Envelope{From: "a@x.com", To: []string{"b@x.com"}, Data: []byte("hi\r\n")})
	if r.Class != ProviderAuth {
		t.Fatalf("expected ProviderAuth, got %v", r.Class)
	}
	if contains(r.Reply+r.Err.Error(), "SuperSecret123!") {
		t.Fatal("password leaked into result")
	}
}

func TestSND_AcceptedOverStartTLS(t *testing.T) {
	srv := startFake(t, &fakesmtp.Script{TLS: fakesmtp.TLSStartTLS})
	pool := rootPool(t, srv)
	host, port := splitAddr(t, srv.SMTPAddr())
	c, err := New(Options{Host: host, Port: port, Mode: ModeSTARTTLS, Dialer: &net.Dialer{}, RootCAs: pool, Timeout: 5 * time.Second}) //nolint:lll
	if err != nil {
		t.Fatal(err)
	}
	r := c.Send(context.Background(), Envelope{From: "a@x.com", To: []string{"b@x.com"}, Data: []byte("hi\r\n")})
	if r.Class != Accepted {
		t.Fatalf("expected Accepted, got %v (%v)", r.Class, r.Err)
	}
}

func splitAddr(t *testing.T, addr string) (string, int) {
	t.Helper()
	h, p, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(p)
	if err != nil {
		t.Fatal(err)
	}
	return h, port
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}
