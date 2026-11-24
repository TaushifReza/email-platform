package platformadmin

import (
	"context"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/audit"
	"github.com/TaushifReza/email-platform/backend/internal/tenancy"
)

// Operator is an identity asserted by the trusted platform identity provider.
// It carries no secret material in the application.
type Operator struct {
	ID    string
	Email string
}

// OperatorCredentials is used only during frontend operator login. The password
// hash is never returned by any API.
type OperatorCredentials struct {
	Operator
	PasswordHash     string
	FailedLoginCount int
	LockedUntil      *string
}

type OperatorSession struct {
	ID                string
	OperatorID        string
	OperatorEmail     string
	MFAVerified       bool
	RevokedAt         *string
	IdleExpiresAt     time.Time
	AbsoluteExpiresAt time.Time
}

// Page bounds an operator list query.
type Page struct {
	Limit  int
	Offset int
}

// CreateOrganizationInput is the operator-facing organization creation input.
type CreateOrganizationInput struct {
	Name              string
	DefaultSenderName string
	DefaultReplyTo    string
	TimeZone          string
}

// LimitsInput holds the per-organization quota and sending caps (ADM-01).
type LimitsInput struct {
	StorageQuotaBytes    int64
	CapsEnabled          bool
	DailyEmailCap        int
	CampaignRecipientCap int
}

// Repository owns all operator SQL. Implementations must run writes in a
// transaction together with the audit entry.
//
//nolint:interfacebloat
type Repository interface {
	LookupOperatorByEmail(ctx context.Context, email string) (Operator, error)
	LookupOperatorCredentials(ctx context.Context, email string) (OperatorCredentials, error)
	CreateOperatorSession(ctx context.Context, id, operatorID string, tokenHash []byte,
		mfaVerified bool, ip, userAgent string, idleExpiresAt, absoluteExpiresAt time.Time) error
	LookupOperatorSession(ctx context.Context, tokenHash []byte) (OperatorSession, error)
	RevokeOperatorSession(ctx context.Context, tokenHash []byte, now time.Time) error
	CreateOperator(ctx context.Context, id, email string) (bool, error)
	ListOrganizations(ctx context.Context, page Page) ([]tenancy.Organization, error)
	CreateOrganization(ctx context.Context, id string, in CreateOrganizationInput,
		entry audit.Entry) (tenancy.Organization, error)
	SetOrganizationStatus(ctx context.Context, orgID, expected, next string, entry audit.Entry) (bool, error)
	UpdateOrganizationLimits(ctx context.Context, orgID string, in LimitsInput,
		entry audit.Entry) (tenancy.Organization, error)
	ListOperatorAudit(ctx context.Context, page Page) ([]audit.Entry, error)
}
