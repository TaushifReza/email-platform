package fakesmtp

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"
)

// session handles a single SMTP connection
type session struct {
	conn     net.Conn
	script   *Script
	server   *Server
	cert     *tls.Certificate
	from     string
	to       []string
	data     strings.Builder
}

// run starts the SMTP session
func (s *session) run() {
	defer s.conn.Close()
	
	// Send greeting
	s.sendResponse(220, "fake Smtp server ready")
	
	// Process commands
	scanner := bufio.NewScanner(s.conn)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		
		cmd := strings.ToUpper(line)
		if strings.HasPrefix(cmd, "EHLO ") || strings.HasPrefix(cmd, "HELO ") {
			s.handleEhlo(cmd)
		} else if strings.HasPrefix(cmd, "MAIL FROM:") {
			s.handleMailFrom(cmd)
		} else if strings.HasPrefix(cmd, "RCPT TO:") {
			s.handleRcptTo(cmd)
		} else if cmd == "DATA" {
			s.handleData()
		} else if cmd == "QUIT" {
			s.handleQuit()
			return
		} else if cmd == "RSET" {
			s.handleRset()
		} else {
			s.sendResponse(500, "Command not recognized")
		}
		
		if s.conn.RemoteAddr() == nil {
			break // connection closed
		}
	}
}

func (s *session) handleEhlo(cmd string) {
	var response string
	
	// Check if TLS is required or supported
	if s.server.cert != nil {
		response = "250-OK\r\n250-STARTTLS\r\n250 OK"
	} else {
		response = "250 OK"
	}
	
	s.sendResponseString(response)
}

func (s *session) handleMailFrom(cmd string) {
	// Parse the FROM parameter
	from := strings.TrimPrefix(cmd, "MAIL FROM:")
	from = strings.TrimSpace(from)
	
	// Apply behavior rules
	replyCode, replyText := s.getBehaviorReply("mail", from, "")
	
	if replyCode != 250 {
		s.sendResponse(replyCode, replyText)
		return
	}
	
	s.from = strings.TrimPrefix(strings.TrimSuffix(from, ">"), "<")
	s.sendResponse(250, "OK")
}

func (s *session) handleRcptTo(cmd string) {
	to := strings.TrimPrefix(cmd, "RCPT TO:")
	to = strings.TrimSpace(to)
	
	// Apply behavior rules
	replyCode, replyText := s.getBehaviorReply("rcpt", "", to)
	
	if replyCode != 250 {
		s.sendResponse(replyCode, replyText)
		return
	}
	
	s.to = append(s.to, strings.TrimPrefix(strings.TrimSuffix(to, ">"), "<"))
	s.sendResponse(250, "OK")
}

func (s *session) handleData() {
	// Apply behavior rules before accepting DATA
	replyCode, replyText := s.getBehaviorReply("data", "", "")
	
	if replyCode != 250 {
		s.sendResponse(replyCode, replyText)
		return
	}
	
	// Send 354 - Start mail input
	s.sendResponse(354, "End data with <CR><LF>.<CR><LF>")
	
	// Read the message content
	var message strings.Builder
	for {
		line := ""
		if err := s.conn.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
			break
		}
		
		scanner := bufio.NewScanner(s.conn)
		if scanner.Scan() {
			line = scanner.Text()
		} else {
			break
		}
		
		if line == "." {
			break
		}
		
		message.WriteString(line)
		message.WriteString("\r\n")
		s.data.WriteString(line)
		s.data.WriteString("\r\n")
	}
	
	// Send 250 - Message accepted
	s.sendResponse(250, "OK")
	
	// Store the message if not dropped
	if !s.willDrop() {
		s.storeMessage()
	}
}

func (s *session) handleQuit() {
	s.sendResponse(221, "Goodbye")
}

func (s *session) handleRset() {
	// Reset session state
	s.from = ""
	s.to = []string{}
	s.data.Reset()
	s.sendResponse(250, "OK")
}

// getBehaviorReply gets the response for a given command
func (s *session) getBehaviorReply(command, from, to string) (int, string) {
	// Find matching behavior in script
	for _, behavior := range s.script.Behaviors {
		if behavior.Command == command || behavior.Command == "all" {
			return behavior.ReplyCode, behavior.ReplyText
		}
	}
	
	// Default fallback - accept all commands
	return 250, "OK"
}

// willDrop checks if the current message should be dropped after DATA
func (s *session) willDrop() bool {
	for _, behavior := range s.script.Behaviors {
		if behavior.Command == "data" || behavior.Command == "all" {
			if behavior.DropAfterData {
				return true
			}
		}
	}
	return false
}

// storeMessage stores the received message
func (s *session) storeMessage() {
	// Simple UUID generation for now
	messageID := fmt.Sprintf("msg_%d", time.Now().UnixNano())
	
	msg := &Message{
		ID:         messageID,
		From:       s.from,
		To:         s.to,
		Size:       len(s.data.String()),
		ReceivedAt: time.Now().Unix(),
		Body:       s.data.String(),
	}
	
	// Store the message
	s.server.mu.Lock()
	s.server.messages = append(s.server.messages, msg)
	s.server.mu.Unlock()
}

// sendResponse sends an SMTP response to client
func (s *session) sendResponse(code int, text string) {
	response := fmt.Sprintf("%d %s\r\n", code, text)
	s.sendResponseString(response)
}

// sendResponseString sends a raw response string to client
func (s *session) sendResponseString(response string) {
	// Apply delay if needed
	for _, behavior := range s.script.Behaviors {
		if behavior.Command == "all" || behavior.Command == "data" {
			if behavior.DelayMS > 0 {
				time.Sleep(time.Duration(behavior.DelayMS) * time.Millisecond)
			}
			break
		}
	}
	
	s.conn.Write([]byte(response))
}

// setTLS sets up TLS for the session if needed
func (s *session) setTLS(tlsConfig *tls.Config) error {
	tlsConn := tls.Server(s.conn, tlsConfig)
	s.conn = tlsConn
	return nil
}