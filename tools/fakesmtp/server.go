package fakesmtp

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"sync"
	"time"
)

// Server represents a fake SMTP server for testing
type Server struct {
	listener    net.Listener
	httpListener net.Listener
	script      *Script
	messages    []*Message
	mu          sync.RWMutex
	shutdown    chan struct{}
	cert        *tls.Certificate
}

// Options configure the fakesmtp server
type Options struct {
	// Script defines the behaviors for the SMTP server
	Script *Script
	
	// TLSCertFile specifies a path to a cert file (optional)
	TLSCertFile string
	
	// TLSKeyFile specifies a path to a key file (optional)
	TLSKeyFile string
	
	// HTTPAddr specifies the address for the HTTP API (default ":0")
	HTTPAddr string
}

// New creates a new fake SMTP server with the given options
func New(opts Options) (*Server, error) {
	s := &Server{
		script:   opts.Script,
		messages: []*Message{},
		shutdown: make(chan struct{}),
	}
	
	if s.script == nil {
		s.script = DefaultScript()
	}
	
	// Create HTTP listener for API
	httpAddr := opts.HTTPAddr
	if httpAddr == "" {
		httpAddr = ":0"
	}
	
	httpListener, err := net.Listen("tcp", httpAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP listener: %w", err)
	}
	s.httpListener = httpListener
	
	// Create SMTP listener
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, fmt.Errorf("failed to create SMTP listener: %w", err)
	}
	s.listener = listener
	
	// Generate self-signed certificate for TLS if needed
	if opts.TLSCertFile == "" && opts.TLSKeyFile == "" {
		cert, err := generateSelfSignedCert()
		if err != nil {
			return nil, fmt.Errorf("failed to generate certificate: %w", err)
		}
		s.cert = cert
	}
	
	return s, nil
}

// Start begins serving SMTP and HTTP requests in separate goroutines
func (s *Server) Start() error {
	go func() {
		for {
			select {
			case <-s.shutdown:
				return
			default:
				conn, err := s.listener.Accept()
				if err != nil {
					continue
				}
				go s.handleConn(conn)
			}
		}
	}()
	
	go s.setupHTTPHandler()
	
	return nil
}

// Shutdown stops the server
func (s *Server) Shutdown() error {
	close(s.shutdown)
	if s.listener != nil {
		s.listener.Close()
	}
	if s.httpListener != nil {
		s.httpListener.Close()
	}
	return nil
}

// SMTPAddr returns the address where the SMTP server is listening
func (s *Server) SMTPAddr() string {
	if s.listener == nil {
		return ""
	}
	return s.listener.Addr().String()
}

// HTTPAddr returns the address where the HTTP API is listening
func (s *Server) HTTPAddr() string {
	if s.httpListener == nil {
		return ""
	}
	return s.httpListener.Addr().String()
}

// Messages returns a copy of all messages received by the server
func (s *Server) Messages() []*Message {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	msgs := make([]*Message, len(s.messages))
	copy(msgs, s.messages)
	return msgs
}

// ClearMessages removes all messages from the server
func (s *Server) ClearMessages() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = []*Message{}
}

// handleConn handles a single SMTP connection
func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	
	// Check if TLS is required
	if s.cert != nil {
		// Upgrade to TLS
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{*s.cert},
			ServerName:   "fakesmtp",
		}
		
		tlsConn := tls.Server(conn, tlsConfig)
		conn = tlsConn
	}
	
	session := &session{
		conn:     conn,
		script:   s.script,
		server:   s,
		cert:     s.cert,
	}
	session.run()
}

// DefaultScript returns a script that accepts all messages
func DefaultScript() *Script {
	return &Script{
		Behaviors: []Behavior{
			{
				Command:   "all",
				ReplyCode: 250,
				ReplyText: "OK",
			},
		},
	}
}

// setupHTTPHandler sets up the HTTP API for message listing
func (s *Server) setupHTTPHandler() {
	mux := http.NewServeMux()
	
	mux.HandleFunc("GET /api/v1/messages", s.listMessagesHandler)
	mux.HandleFunc("GET /api/v1/messages/{id}", s.getMessageHandler)
	mux.HandleFunc("DELETE /api/v1/messages", s.clearMessagesHandler)
	mux.HandleFunc("GET /healthz", s.healthHandler)
	
	// This would be called by an HTTP server
	go func() {
		if s.httpListener != nil {
			http.Serve(s.httpListener, mux)
		}
	}()
}

func (s *Server) listMessagesHandler(w http.ResponseWriter, r *http.Request) {
	s.Messages() // Use to ensure we don't have unused variable warning
	
	w.Header().Set("Content-Type", "application/json")
	// For now just write an empty JSON array to satisfy the prototype
	fmt.Fprintf(w, "[]")
}

func (s *Server) getMessageHandler(w http.ResponseWriter, r *http.Request) {
	// Get message ID from URL path: /api/v1/messages/{id}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{}") // Placeholder for now
}

func (s *Server) clearMessagesHandler(w http.ResponseWriter, r *http.Request) {
	s.ClearMessages()
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status": "ok"}`)
}

// generateSelfSignedCert generates a self-signed certificate for TLS
func generateSelfSignedCert() (*tls.Certificate, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	
	SerialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, err
	}
	
	template := x509.Certificate{
		SerialNumber: SerialNumber,
		Subject: pkix.Name{
			Organization: []string{"Fake SMTP Server"},
			CommonName:   "fakesmtp",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour),
		
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, err
	}
	
	certBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	
	cert, err := tls.X509KeyPair(certBytes, keyBytes)
	if err != nil {
		return nil, err
	}
	
	return &cert, nil
}