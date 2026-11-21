package fakesmtp

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"testing"
	"time"
)

func startServer(t *testing.T, script *Script) *Server {
	t.Helper()

	server, err := New(Options{Script: script})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := server.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() {
		if err := server.Shutdown(); err != nil {
			t.Errorf("Shutdown: %v", err)
		}
	})
	return server
}

func sendPlain(t *testing.T, server *Server, subject string) error {
	t.Helper()

	body := fmt.Sprintf("Subject: %s\r\n\r\nhello body", subject)
	return smtp.SendMail(server.SMTPAddr(), nil, "from@example.com", []string{"to@example.com"}, []byte(body))
}

func TestAcceptsAndListsMessage(t *testing.T) {
	server := startServer(t, nil)

	if err := sendPlain(t, server, "first"); err != nil {
		t.Fatalf("SendMail: %v", err)
	}

	messages := server.Messages()
	if len(messages) != 1 {
		t.Fatalf("got %d messages, want 1", len(messages))
	}
	if messages[0].Subject != "first" {
		t.Fatalf("subject = %q, want first", messages[0].Subject)
	}
	if !strings.Contains(messages[0].Body, "hello body") {
		t.Fatalf("body = %q", messages[0].Body)
	}
	if messages[0].From != "from@example.com" {
		t.Fatalf("from = %q", messages[0].From)
	}
	if len(messages[0].To) != 1 || messages[0].To[0] != "to@example.com" {
		t.Fatalf("to = %#v", messages[0].To)
	}
}

func TestScriptedFailures(t *testing.T) {
	tests := []struct {
		name   string
		script *Script
		want   string
	}{
		{
			name:   "4xx mail",
			script: &Script{Behaviors: []Behavior{{Command: "mail", ReplyCode: 450, ReplyText: "try later"}}},
			want:   "450",
		},
		{
			name:   "5xx rcpt",
			script: &Script{Behaviors: []Behavior{{Command: "rcpt", ReplyCode: 550, ReplyText: "no such user"}}},
			want:   "550",
		},
		{
			name:   "5xx data",
			script: &Script{Behaviors: []Behavior{{Command: "data", ReplyCode: 554, ReplyText: "rejected"}}},
			want:   "554",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := startServer(t, test.script)

			err := sendPlain(t, server, "rejected")
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want code %s", err, test.want)
			}
			if messages := server.Messages(); len(messages) != 0 {
				t.Fatalf("stored %d messages after failure", len(messages))
			}
		})
	}
}

func TestDropsMessageAfterData(t *testing.T) {
	server := startServer(t, &Script{
		Behaviors: []Behavior{{Command: "data", DropAfterData: true}},
	})

	if err := sendPlain(t, server, "dropped"); err != nil {
		t.Logf("send returned after drop: %v", err)
	}

	if messages := server.Messages(); len(messages) != 0 {
		t.Fatalf("stored %d messages after drop", len(messages))
	}
}

func TestDelaysResponse(t *testing.T) {
	server := startServer(t, &Script{
		Behaviors: []Behavior{{Command: "data", DelayMS: 150}},
	})

	start := time.Now()
	if err := sendPlain(t, server, "delayed"); err != nil {
		t.Fatalf("SendMail: %v", err)
	}
	if elapsed := time.Since(start); elapsed < 100*time.Millisecond {
		t.Fatalf("elapsed = %v, want at least 100ms", elapsed)
	}
}

func TestSTARTTLS(t *testing.T) {
	server := startServer(t, &Script{TLS: TLSStartTLS})

	conn, err := net.Dial("tcp", server.SMTPAddr())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, "localhost")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if ok, _ := client.Extension("STARTTLS"); !ok {
		t.Fatal("server did not advertise STARTTLS")
	}
	tlsConfig := &tls.Config{InsecureSkipVerify: true} //nolint:gosec // Test-only self-signed certificate.
	if err := client.StartTLS(tlsConfig); err != nil {
		t.Fatalf("StartTLS: %v", err)
	}
	sendViaClient(t, client, "starttls")

	if messages := server.Messages(); len(messages) != 1 || messages[0].Subject != "starttls" {
		t.Fatalf("messages = %#v", messages)
	}
}

func TestImplicitTLS(t *testing.T) {
	server := startServer(t, &Script{TLS: TLSImplicit})

	tlsConfig := &tls.Config{InsecureSkipVerify: true} //nolint:gosec // Test-only self-signed certificate.
	conn, err := tls.Dial("tcp", server.SMTPAddr(), tlsConfig)
	if err != nil {
		t.Fatalf("tls dial: %v", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, "localhost")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	sendViaClient(t, client, "implicit")

	if messages := server.Messages(); len(messages) != 1 || messages[0].Subject != "implicit" {
		t.Fatalf("messages = %#v", messages)
	}
}

func TestHTTPAPI(t *testing.T) {
	server := startServer(t, nil)

	if err := sendPlain(t, server, "http"); err != nil {
		t.Fatalf("SendMail: %v", err)
	}

	base := "http://" + server.HTTPAddr()
	var messages []Message
	getJSON(t, base+"/api/v1/messages", &messages)
	if len(messages) != 1 || messages[0].Subject != "http" {
		t.Fatalf("messages = %#v", messages)
	}

	var message Message
	getJSON(t, base+"/api/v1/messages/"+messages[0].ID, &message)
	if message.ID != messages[0].ID {
		t.Fatalf("message ID = %q, want %q", message.ID, messages[0].ID)
	}

	var health map[string]string
	getJSON(t, base+"/healthz", &health)
	if health["status"] != "ok" {
		t.Fatalf("health = %#v", health)
	}

	request, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, base+"/api/v1/messages", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("DELETE: %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("DELETE status = %d", response.StatusCode)
	}
	if messages := server.Messages(); len(messages) != 0 {
		t.Fatalf("messages remain after delete: %#v", messages)
	}
}

func sendViaClient(t *testing.T, client *smtp.Client, subject string) {
	t.Helper()

	if err := client.Mail("from@example.com"); err != nil {
		t.Fatalf("Mail: %v", err)
	}
	if err := client.Rcpt("to@example.com"); err != nil {
		t.Fatalf("Rcpt: %v", err)
	}
	writer, err := client.Data()
	if err != nil {
		t.Fatalf("Data: %v", err)
	}
	if _, err := io.WriteString(writer, "Subject: "+subject+"\r\n\r\nbody"); err != nil {
		t.Fatalf("write DATA: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close DATA: %v", err)
	}
	if err := client.Quit(); err != nil {
		t.Fatalf("Quit: %v", err)
	}
}

func getJSON(t *testing.T, url string, target any) {
	t.Helper()

	request, err := http.NewRequestWithContext(
		context.Background(), http.MethodGet, url, nil,
	)
	if err != nil {
		t.Fatalf("NewRequest %s: %v", url, err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(response.Body)
		if readErr != nil {
			t.Fatalf("read GET %s response: %v", url, readErr)
		}
		t.Fatalf("GET %s status %d: %s", url, response.StatusCode, body)
	}
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatalf("decode %s: %v", url, err)
	}
}
