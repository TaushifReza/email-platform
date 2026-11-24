package smtpclient

import "strings"

// Class is the SMTP reply classification used by the retry and pause logic.
type Class uint8

const (
	Accepted Class = iota
	Temporary
	TemporaryConnection
	PermanentRecipient
	PermanentMessage
	ProviderAuth
	ProviderPolicy
	ProviderQuota
	Uncertain // reply ambiguous after the body was already sent
)

// Stage is the point in the conversation where a failure happened.
type Stage uint8

const (
	StageConnect Stage = iota
	StageTLS
	StageAuth
	StageMailFrom
	StageRcpt
	StageData
	StageBody
	StageReply
)

// classify is the single reply-classification table from the sending protocol.
// It is stage-aware: a 550 at MAIL FROM is provider-wide, at RCPT it is a
// permanent recipient failure. Unknown replies never become a silent success.
//
//nolint:gocognit,gocyclo
func classify(stage Stage, code int, reply string) Class {
	text := strings.ToLower(reply)

	if code >= 200 && code < 300 && stage >= StageBody {
		return Accepted
	}

	// Provider-wide auth failures.
	if code == 530 || code == 534 || code == 535 {
		return ProviderAuth
	}

	// Sender block / reputation / policy (5.7.x).
	if strings.HasPrefix(text, "5.7.") || strings.Contains(text, "5.7.") {
		return ProviderPolicy
	}
	if strings.Contains(text, "blocked") || strings.Contains(text, "reputation") ||
		strings.Contains(text, "policy") {
		return ProviderPolicy
	}

	// Provider daily quota exceeded.
	if strings.Contains(text, "quota") && (strings.Contains(text, "exceed") ||
		strings.Contains(text, "exhaust")) {
		return ProviderQuota
	}

	switch stage {
	case StageConnect, StageTLS:
		return TemporaryConnection
	case StageAuth:
		if code >= 400 && code < 500 {
			return Temporary
		}
		return ProviderAuth
	case StageMailFrom:
		if code >= 400 && code < 500 {
			return Temporary
		}
		return ProviderPolicy
	case StageRcpt:
		if code == 550 || code == 551 || code == 553 {
			return PermanentRecipient
		}
		if code >= 400 && code < 500 {
			return Temporary
		}
		return Temporary
	case StageData, StageBody, StageReply:
		if code == 552 || code == 554 {
			return PermanentMessage
		}
		if code >= 400 && code < 500 {
			return Temporary
		}
		if code >= 500 {
			return PermanentMessage
		}
	}

	// Enhanced 4.x.x -> temporary.
	if len(text) >= 3 && text[0] == '4' && text[1] == '.' {
		return Temporary
	}
	if code >= 400 && code < 500 {
		return Temporary
	}
	if code >= 500 {
		return PermanentMessage
	}
	return Temporary
}
