package providers

import (
	"context"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/audit"
	"github.com/TaushifReza/email-platform/backend/internal/platform/clock"
	"github.com/TaushifReza/email-platform/backend/internal/platform/crypto"
)

type Provider struct {
	ID             string
	OrganizationID string
	Host           string
	Port           int
	TLSMode        string
	Username       string
	FromAddress    string
	IsDefault      bool
	Status         string
	DailyLimit     int
	MinuteLimit    int
	HourLimit      int
	ConfigVersion  int
	VerifiedAt     *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Input struct {
	Host        string
	Port        int
	TLSMode     string
	Username    string
	Password    string
	FromAddress string
	IsDefault   bool
	DailyLimit  int
	MinuteLimit int
	HourLimit   int
}

type Update struct {
	Host        *string
	Port        *int
	TLSMode     *string
	FromAddress *string
	IsDefault   *bool
	DailyLimit  *int
	MinuteLimit *int
	HourLimit   *int
}

type Credentials struct {
	Username string
	Password string
}

type Repository interface {
	List(ctx context.Context, orgID string) ([]Provider, error)
	Get(ctx context.Context, orgID, id string) (Provider, error)
	Create(ctx context.Context, orgID string, p Provider, enc []byte, keyID string, now time.Time, entry audit.Entry) error
	Update(ctx context.Context, orgID, id string, u Update, now time.Time, entry audit.Entry) (Provider, error)
	Rotate(ctx context.Context, orgID, id, username string, enc []byte, keyID string, now time.Time, entry audit.Entry) (Provider, error) //nolint:lll
	Deactivate(ctx context.Context, orgID, id string, now time.Time, entry audit.Entry) (Provider, error)
	SetDefault(ctx context.Context, orgID, id string, now time.Time, entry audit.Entry) (Provider, error)
	EncryptedCredentials(ctx context.Context, orgID, id string) (username string, ciphertext []byte, keyID string, err error) //nolint:lll
}

type Encryptor interface {
	Encrypt(purpose string, aad crypto.AAD, plaintext []byte) ([]byte, string, error)
	Decrypt(purpose string, aad crypto.AAD, ciphertext []byte, keyID string) ([]byte, error)
}

type TargetValidator interface {
	ValidateTarget(ctx context.Context, host string, port int) error
}

type Service struct {
	repo    Repository
	keyring Encryptor
	targets TargetValidator
	clock   clock.Clock
}

func NewService(repo Repository, keyring Encryptor, targets TargetValidator, clk clock.Clock) *Service {
	if clk == nil {
		clk = clock.System{}
	}
	return &Service{repo: repo, keyring: keyring, targets: targets, clock: clk}
}
