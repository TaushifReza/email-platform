package platformadmin

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/platform/clock"
)

// OperatorSessionResult describes a newly created operator session.
type OperatorSessionResult struct {
	Token             string
	OperatorID        string
	IdleExpiresAt     time.Time
	AbsoluteExpiresAt time.Time
}

// IssueOperatorSession creates a session for a successfully authenticated operator.
func IssueOperatorSession(ctx context.Context, repo Repository, clk clock.Clock,
	op Operator, ip, userAgent string, idle, abs time.Duration,
) (OperatorSessionResult, error) {
	now := clk.Now()
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return OperatorSessionResult{}, err
	}
	raw := base64.RawURLEncoding.EncodeToString(b)
	h := sha256.Sum256([]byte(raw))
	id, err := newUUID()
	if err != nil {
		return OperatorSessionResult{}, err
	}
	idleExp := now.Add(idle)
	absExp := now.Add(abs)
	if err := repo.CreateOperatorSession(ctx, id, op.ID, h[:], true, ip, userAgent, idleExp, absExp); err != nil {
		return OperatorSessionResult{}, err
	}
	return OperatorSessionResult{Token: raw, OperatorID: op.ID, IdleExpiresAt: idleExp, AbsoluteExpiresAt: absExp}, nil
}
