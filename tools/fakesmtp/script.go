package fakesmtp

import "time"

// TLSMode controls how the fake server exposes TLS.
type TLSMode string

const (
	TLSNone     TLSMode = "none"
	TLSStartTLS TLSMode = "starttls"
	TLSImplicit TLSMode = "implicit"
)

// Script defines the scripted behaviors for the SMTP server.
type Script struct {
	TLS       TLSMode    `json:"tls"`
	Behaviors []Behavior `json:"behaviors"`
}

// Behavior represents a single scripted SMTP response. Command is one of
// "connect", "ehlo", "mail", "rcpt", "data", or "all".
type Behavior struct {
	Command       string `json:"command"`
	ReplyCode     int    `json:"reply_code"`
	ReplyText     string `json:"reply_text"`
	DropAfterData bool   `json:"drop_after_data"`
	DelayMS       int    `json:"delay_ms"`
}

// Message represents a received email message.
type Message struct {
	ID         string              `json:"id"`
	From       string              `json:"from"`
	To         []string            `json:"to"`
	Subject    string              `json:"subject"`
	Headers    map[string][]string `json:"headers"`
	Body       string              `json:"body"`
	Raw        string              `json:"raw"`
	Size       int                 `json:"size"`
	ReceivedAt int64               `json:"received_at"`
}

func (s *Script) behavior(command string) (Behavior, bool) {
	for _, candidate := range s.Behaviors {
		if candidate.Command == command {
			return candidate, true
		}
	}
	for _, candidate := range s.Behaviors {
		if candidate.Command == "all" {
			return candidate, true
		}
	}
	return Behavior{}, false
}

func (s *Script) delay(command string) time.Duration {
	behavior, ok := s.behavior(command)
	if !ok || behavior.DelayMS <= 0 {
		return 0
	}
	return time.Duration(behavior.DelayMS) * time.Millisecond
}

func (s *Script) dropAfterData() bool {
	behavior, ok := s.behavior("data")
	return ok && behavior.DropAfterData
}
