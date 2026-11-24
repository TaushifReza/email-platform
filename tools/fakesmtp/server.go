package fakesmtp

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

// Server is a fake SMTP server with an HTTP API that lists received messages.
type Server struct {
	smtpListener net.Listener
	httpListener net.Listener
	httpServer   *http.Server
	script       *Script
	cert         *tls.Certificate
	certPEM      []byte

	mu       sync.RWMutex
	messages []*Message
	nextID   int

	closeOnce sync.Once
	done      chan struct{}
	wg        sync.WaitGroup
}

// Options configure the fake SMTP server.
type Options struct {
	Script      *Script
	SMTPAddr    string
	HTTPAddr    string
	TLSCertFile string
	TLSKeyFile  string
}

// New creates a fake SMTP server. Call Start to begin accepting connections.
func New(opts Options) (*Server, error) {
	script := opts.Script
	if script == nil {
		script = DefaultScript()
	}

	smtpAddr := opts.SMTPAddr
	if smtpAddr == "" {
		smtpAddr = "127.0.0.1:0"
	}
	httpAddr := opts.HTTPAddr
	if httpAddr == "" {
		httpAddr = "127.0.0.1:0"
	}

	var cert *tls.Certificate
	var certPEM []byte
	if script.TLS != TLSNone || opts.TLSCertFile != "" || opts.TLSKeyFile != "" {
		loaded, pemBytes, err := loadCertificate(opts.TLSCertFile, opts.TLSKeyFile)
		if err != nil {
			return nil, err
		}
		cert = loaded
		certPEM = pemBytes
	}

	smtpListener, err := net.Listen("tcp", smtpAddr)
	if err != nil {
		return nil, fmt.Errorf("listen SMTP: %w", err)
	}

	httpListener, err := net.Listen("tcp", httpAddr)
	if err != nil {
		_ = smtpListener.Close()
		return nil, fmt.Errorf("listen HTTP: %w", err)
	}

	server := &Server{
		smtpListener: smtpListener,
		httpListener: httpListener,
		script:       script,
		cert:         cert,
		certPEM:      certPEM,
		done:         make(chan struct{}),
	}
	server.httpServer = &http.Server{
		Handler:           server.handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	return server, nil
}

// Start begins serving SMTP and HTTP connections.
func (s *Server) Start() error {
	s.wg.Add(2)
	go s.serveSMTP()
	go s.serveHTTP()
	return nil
}

// Shutdown stops the server and waits for its accept loops to exit.
func (s *Server) Shutdown() error {
	s.closeOnce.Do(func() {
		close(s.done)
		_ = s.smtpListener.Close()
		_ = s.httpServer.Close()
		_ = s.httpListener.Close()
	})
	s.wg.Wait()
	return nil
}

// SMTPAddr returns the SMTP listen address.
func (s *Server) SMTPAddr() string {
	return s.smtpListener.Addr().String()
}

// CACertPEM returns the PEM-encoded certificate used by the fake server, so
// tests can trust it via a root pool instead of disabling verification.
func (s *Server) CACertPEM() []byte {
	return s.certPEM
}

// HTTPAddr returns the HTTP listen address.
func (s *Server) HTTPAddr() string {
	return s.httpListener.Addr().String()
}

// Messages returns a copy of all received messages.
func (s *Server) Messages() []*Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	messages := make([]*Message, len(s.messages))
	copy(messages, s.messages)
	return messages
}

// ClearMessages removes all received messages.
func (s *Server) ClearMessages() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = nil
}

func (s *Server) serveSMTP() {
	defer s.wg.Done()

	for {
		conn, err := s.smtpListener.Accept()
		if err != nil {
			return
		}
		go s.handleConn(conn)
	}
}

func (s *Server) serveHTTP() {
	defer s.wg.Done()
	_ = s.httpServer.Serve(s.httpListener) //nolint:errcheck // Closed during shutdown.
}

func (s *Server) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/messages", s.listMessages)
	mux.HandleFunc("GET /api/v1/messages/{id}", s.getMessage)
	mux.HandleFunc("DELETE /api/v1/messages", s.deleteMessages)
	mux.HandleFunc("GET /healthz", s.health)
	return mux
}

func (s *Server) listMessages(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.Messages())
}

func (s *Server) getMessage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	for _, message := range s.Messages() {
		if message.ID == id {
			writeJSON(w, http.StatusOK, message)
			return
		}
	}
	writeJSON(w, http.StatusNotFound, map[string]string{"error": "message not found"})
}

func (s *Server) deleteMessages(w http.ResponseWriter, _ *http.Request) {
	s.ClearMessages()
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	isTLS := false
	if s.script.TLS == TLSImplicit {
		if s.cert == nil {
			return
		}
		tlsConn := tls.Server(conn, &tls.Config{
			Certificates: []tls.Certificate{*s.cert},
			MinVersion:   tls.VersionTLS12,
		})
		if err := tlsConn.Handshake(); err != nil {
			return
		}
		conn = tlsConn
		isTLS = true
	}

	session := newSession(s, conn)
	session.tls = isTLS
	session.run()
}

func (s *Server) storeMessage(from string, to []string, raw []byte) *Message {
	message := parseMessage(from, to, raw)

	s.mu.Lock()
	s.nextID++
	message.ID = fmt.Sprintf("msg_%d", s.nextID)
	s.messages = append(s.messages, message)
	s.mu.Unlock()

	return message
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		http.Error(w, "encode response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err = w.Write(data); err != nil {
		return
	}
}

// DefaultScript accepts every message over plaintext SMTP.
func DefaultScript() *Script {
	return &Script{TLS: TLSNone}
}

func loadCertificate(certFile, keyFile string) (*tls.Certificate, []byte, error) {
	if certFile != "" || keyFile != "" {
		if certFile == "" || keyFile == "" {
			return nil, nil, errors.New("both TLSCertFile and TLSKeyFile are required")
		}
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, nil, fmt.Errorf("load TLS key pair: %w", err)
		}
		pemBytes, err := os.ReadFile(certFile) // #nosec G304
		if err != nil {
			return nil, nil, fmt.Errorf("read cert: %w", err)
		}
		return &cert, pemBytes, nil
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("generate key: %w", err)
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, fmt.Errorf("generate serial: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{"Fake SMTP Server"},
			CommonName:   "fakesmtp",
		},
		NotBefore:             time.Now().Add(-time.Minute),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"fakesmtp", "localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}

	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, nil, fmt.Errorf("create certificate: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, nil, fmt.Errorf("parse certificate: %w", err)
	}
	return &cert, certPEM, nil
}
