package notifications

import (
	"context"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/platform/clock"
	"github.com/TaushifReza/email-platform/backend/internal/platform/db"
	"github.com/TaushifReza/email-platform/backend/internal/platform/outbox"
	"go.uber.org/zap"
)

// Processor drains the system email outbox.
type Processor struct {
	db          *db.DB
	store       *outbox.Store
	sender      *SMTPSender
	clock       clock.Clock
	logger      *zap.Logger
	maxAttempts int
}

// NewProcessor creates a processor.
func NewProcessor(database *db.DB, store *outbox.Store, sender *SMTPSender,
	clk clock.Clock, logger *zap.Logger,
) *Processor {
	if clk == nil {
		clk = clock.System{}
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Processor{db: database, store: store, sender: sender, clock: clk, logger: logger, maxAttempts: 5}
}

// Run loops until ctx is done.
func (p *Processor) Run(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = p.ProcessOnce(ctx) //nolint:errcheck
		}
	}
}

// ProcessOnce processes one batch.
func (p *Processor) ProcessOnce(ctx context.Context) error {
	lease := p.clock.Now().Add(2 * time.Minute)
	claimed, err := p.store.Claim(ctx, 10, lease)
	if err != nil {
		p.logger.Error("claim system emails failed", zap.Error(err))
		return err
	}
	for _, m := range claimed {
		sendErr := p.sender.Send(ctx, m.Recipient, m.Subject, m.TextBody)
		if sendErr == nil {
			if err := p.store.MarkSent(ctx, m.ID); err != nil {
				p.logger.Error("mark sent failed", zap.Error(err), zap.String("id", m.ID))
			}
			continue
		}
		shift := int(m.AttemptCount)
		if shift > 10 {
			shift = 10
		}
		next := p.clock.Now().Add(time.Duration(1<<shift) * time.Minute)
		if next.After(p.clock.Now().Add(15 * time.Minute)) {
			next = p.clock.Now().Add(15 * time.Minute)
		}
		if err := p.store.MarkFailed(ctx, m.ID, int(m.AttemptCount), p.maxAttempts, next, sendErr.Error()); err != nil {
			p.logger.Error("mark failed", zap.Error(err), zap.String("id", m.ID))
		}
	}
	return nil
}
