package identity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvitationLinkUsesAcceptFragment(t *testing.T) {
	got := invitationURL("https://example.test", "abc-123")
	assert.Equal(t, "https://example.test/invitations/accept#token=abc-123", got)
}
