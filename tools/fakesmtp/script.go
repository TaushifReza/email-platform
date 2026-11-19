package fakesmtp

// Script defines the scripted behaviors for the SMTP server
type Script struct {
	Behaviors []Behavior `json:"behaviors"`
}

// Behavior represents a single SMTP behavior in the script
type Behavior struct {
	Command   string `json:"command"`   // "ehlo", "mail", "rcpt", "data", "all"
	ReplyCode int    `json:"reply_code"`
	ReplyText string `json:"reply_text"`
	DropAfterData bool `json:"drop_after_data"`
	DelayMS   int    `json:"delay_ms"`
	TLS       string `json:"tls"` // "none", "starttls", "implicit"
}

// Message represents a received email message
type Message struct {
	ID         string     `json:"id"`
	From       string     `json:"from"`
	To         []string   `json:"to"`
	Size       int        `json:"size"`
	ReceivedAt int64      `json:"received_at"`
	Subject    string     `json:"subject"`
	Headers    []string   `json:"headers"`
	Body       string     `json:"body"`
}