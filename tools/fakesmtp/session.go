package fakesmtp

import (
	"bytes"
	"crypto/tls"
	"io"
	"net"
	"net/mail"
	"net/textproto"
	"strings"
	"time"
)

type session struct {
	server *Server
	conn   net.Conn
	tp     *textproto.Conn
	from   string
	to     []string
	tls    bool
}

func newSession(server *Server, conn net.Conn) *session {
	return &session{
		server: server,
		conn:   conn,
		tp:     textproto.NewConn(conn),
	}
}

func (s *session) run() {
	s.reply("connect", 220, "fake SMTP server ready")

	for {
		line, err := s.tp.ReadLine()
		if err != nil {
			return
		}

		command, argument := splitCommand(line)
		switch command {
		case "EHLO":
			s.handleEHLO(argument)
		case "HELO":
			s.handleHELO(argument)
		case "MAIL":
			s.handleMail(argument)
		case "RCPT":
			s.handleRcpt(argument)
		case "DATA":
			s.handleData()
		case "RSET":
			s.reset()
			s.reply("rset", 250, "OK")
		case "NOOP":
			s.reply("noop", 250, "OK")
		case "STARTTLS":
			s.handleStartTLS()
		case "AUTH":
			s.reply("auth", 235, "2.7.0 Authentication successful")
		case "QUIT":
			s.reply("quit", 221, "Bye")
			return
		default:
			s.reply("", 500, "Command not recognized")
		}
	}
}

func (s *session) handleEHLO(_ string) {
	behavior, hasBehavior := s.server.script.behavior("ehlo")
	if hasBehavior && behavior.ReplyCode != 0 && behavior.ReplyCode != 250 {
		s.reply("ehlo", behavior.ReplyCode, behavior.ReplyText)
		return
	}

	s.sleep("ehlo")
	text := "fakesmtp"
	if hasBehavior && behavior.ReplyText != "" {
		text = behavior.ReplyText
	}

	s.writeLine("250-%s", text)
	if s.server.script.TLS == TLSStartTLS && !s.tls {
		s.writeLine("250-STARTTLS")
	}
	s.writeLine("250 AUTH PLAIN LOGIN")
}

func (s *session) handleHELO(_ string) {
	s.reset()
	s.reply("helo", 250, "fakesmtp")
}

func (s *session) handleMail(argument string) {
	if !strings.HasPrefix(strings.ToUpper(argument), "FROM:") {
		s.reply("mail", 501, "Syntax: MAIL FROM:<address>")
		return
	}

	behavior, hasBehavior := s.server.script.behavior("mail")
	if hasBehavior && behavior.ReplyCode != 0 && behavior.ReplyCode != 250 {
		s.reply("mail", behavior.ReplyCode, behavior.ReplyText)
		return
	}

	s.from = parseAddress(argument[len("FROM:"):])
	s.reply("mail", 250, "OK")
}

func (s *session) handleRcpt(argument string) {
	if !strings.HasPrefix(strings.ToUpper(argument), "TO:") {
		s.reply("rcpt", 501, "Syntax: RCPT TO:<address>")
		return
	}

	behavior, hasBehavior := s.server.script.behavior("rcpt")
	if hasBehavior && behavior.ReplyCode != 0 && behavior.ReplyCode != 250 {
		s.reply("rcpt", behavior.ReplyCode, behavior.ReplyText)
		return
	}

	s.to = append(s.to, parseAddress(argument[len("TO:"):]))
	s.reply("rcpt", 250, "OK")
}

func (s *session) handleData() {
	behavior, hasBehavior := s.server.script.behavior("data")
	if hasBehavior && behavior.ReplyCode != 0 && behavior.ReplyCode != 250 {
		s.reply("data", behavior.ReplyCode, behavior.ReplyText)
		return
	}

	s.sleep("data")
	s.writeLine("354 End data with <CR><LF>.<CR><LF>")

	raw, err := io.ReadAll(s.tp.DotReader())
	if err != nil {
		return
	}

	if s.server.script.dropAfterData() {
		s.writeLine("250 OK")
		return
	}

	s.server.storeMessage(s.from, append([]string(nil), s.to...), raw)
	s.reset()
	s.writeLine("250 OK")
}

func (s *session) handleStartTLS() {
	if s.server.script.TLS != TLSStartTLS || s.tls || s.server.cert == nil {
		s.reply("starttls", 502, "Command not implemented")
		return
	}

	s.writeLine("220 Ready to start TLS")

	tlsConn := tls.Server(s.conn, &tls.Config{
		Certificates: []tls.Certificate{*s.server.cert},
		MinVersion:   tls.VersionTLS12,
	})
	if err := tlsConn.Handshake(); err != nil {
		return
	}

	s.conn = tlsConn
	s.tp = textproto.NewConn(tlsConn)
	s.tls = true
	s.reset()
}

func (s *session) reset() {
	s.from = ""
	s.to = nil
}

func (s *session) reply(command string, defaultCode int, defaultText string) {
	code := defaultCode
	text := defaultText

	if behavior, ok := s.server.script.behavior(command); ok {
		if behavior.ReplyCode != 0 {
			code = behavior.ReplyCode
		}
		if behavior.ReplyText != "" {
			text = behavior.ReplyText
		}
	}

	s.sleep(command)
	s.writeLine("%d %s", code, text)
}

func (s *session) writeLine(format string, args ...any) {
	_ = s.tp.PrintfLine(format, args...) //nolint:errcheck // The connection closes when the client disconnects.
}

func (s *session) sleep(command string) {
	if delay := s.server.script.delay(command); delay > 0 {
		time.Sleep(delay)
	}
}

func splitCommand(line string) (string, string) {
	line = strings.TrimSpace(line)
	index := strings.IndexAny(line, " \t")
	if index == -1 {
		return strings.ToUpper(line), ""
	}
	return strings.ToUpper(line[:index]), strings.TrimSpace(line[index+1:])
}

func parseAddress(value string) string {
	value = strings.TrimSpace(value)
	if index := strings.IndexAny(value, " \t"); index != -1 {
		value = value[:index]
	}
	return strings.Trim(value, "<>")
}

func parseMessage(from string, to []string, raw []byte) *Message {
	message := &Message{
		From:       from,
		To:         to,
		Headers:    map[string][]string{},
		Body:       string(raw),
		Raw:        string(raw),
		Size:       len(raw),
		ReceivedAt: time.Now().Unix(),
	}

	parsed, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return message
	}

	for key, values := range parsed.Header {
		message.Headers[key] = append([]string(nil), values...)
	}
	message.Subject = parsed.Header.Get("Subject")

	if body, err := io.ReadAll(parsed.Body); err == nil {
		message.Body = string(body)
	}

	return message
}
