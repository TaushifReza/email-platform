package platformadmin

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/audit"
	"github.com/TaushifReza/email-platform/backend/internal/platform/db"
	"github.com/TaushifReza/email-platform/backend/internal/platform/db/sqlc"
	"github.com/TaushifReza/email-platform/backend/internal/tenancy"
)

type repository struct {
	db *db.DB
}

const nullJSON = "null"

func NewRepository(database *db.DB) Repository {
	return &repository{db: database}
}

type operatorPayload struct {
	ID     string `json:"id"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

func (r *repository) LookupOperatorByEmail(ctx context.Context, email string) (Operator, error) {
	var out Operator
	err := r.db.WithPreAuth(ctx, func(tx *sql.Tx) error {
		raw, err := sqlc.New(tx).LookupOperatorByEmail(ctx, email)
		if err != nil {
			return err
		}
		if len(raw) == 0 || string(raw) == nullJSON {
			return nil
		}
		var payload operatorPayload
		if err := json.Unmarshal(raw, &payload); err != nil {
			return err
		}
		if payload.Status == "active" {
			out = Operator{ID: payload.ID, Email: payload.Email}
		}
		return nil
	})
	return out, err
}

func (r *repository) CreateOperatorSession(ctx context.Context, id, operatorID string,
	tokenHash []byte, mfaVerified bool, ip, userAgent string,
	idleExpiresAt, absoluteExpiresAt time.Time,
) error {
	return r.db.WithSystem(ctx, func(tx *sql.Tx) error {
		return sqlc.New(tx).CreateOperatorSession(ctx, sqlc.CreateOperatorSessionParams{
			ID:                id,
			OperatorID:        operatorID,
			TokenHash:         tokenHash,
			MfaVerified:       mfaVerified,
			Ip:                ip,
			UserAgent:         userAgent,
			IdleExpiresAt:     idleExpiresAt,
			AbsoluteExpiresAt: absoluteExpiresAt,
		})
	})
}

func (r *repository) LookupOperatorSession(ctx context.Context, tokenHash []byte) (OperatorSession, error) {
	var out OperatorSession
	err := r.db.WithPreAuth(ctx, func(tx *sql.Tx) error {
		raw, err := sqlc.New(tx).LookupOperatorSession(ctx, tokenHash)
		if err != nil {
			return err
		}
		if len(raw) == 0 || string(raw) == nullJSON {
			return ErrOperatorNotFound
		}
		var payload struct {
			SessionID         string    `json:"session_id"`
			OperatorID        string    `json:"operator_id"`
			OperatorEmail     string    `json:"operator_email"`
			MFAVerified       bool      `json:"mfa_verified"`
			RevokedAt         *string   `json:"revoked_at"`
			IdleExpiresAt     time.Time `json:"idle_expires_at"`
			AbsoluteExpiresAt time.Time `json:"absolute_expires_at"`
		}
		if err := json.Unmarshal(raw, &payload); err != nil {
			return err
		}
		out = OperatorSession{
			ID:                payload.SessionID,
			OperatorID:        payload.OperatorID,
			OperatorEmail:     payload.OperatorEmail,
			MFAVerified:       payload.MFAVerified,
			RevokedAt:         payload.RevokedAt,
			IdleExpiresAt:     payload.IdleExpiresAt,
			AbsoluteExpiresAt: payload.AbsoluteExpiresAt,
		}
		return nil
	})
	return out, err
}

func (r *repository) RevokeOperatorSession(ctx context.Context, tokenHash []byte, now time.Time) error {
	return r.db.WithSystem(ctx, func(tx *sql.Tx) error {
		return sqlc.New(tx).RevokeOperatorSession(ctx, sqlc.RevokeOperatorSessionParams{
			Now:       now,
			TokenHash: tokenHash,
		})
	})
}

func (r *repository) LookupOperatorCredentials(ctx context.Context, email string) (OperatorCredentials, error) {
	var out OperatorCredentials
	err := r.db.WithPreAuth(ctx, func(tx *sql.Tx) error {
		raw, err := sqlc.New(tx).LookupOperatorCredentials(ctx, email)
		if err != nil {
			return err
		}
		if len(raw) == 0 || string(raw) == nullJSON {
			return ErrOperatorNotFound
		}
		var payload struct {
			ID               string  `json:"id"`
			Email            string  `json:"email"`
			Status           string  `json:"status"`
			PasswordHash     *string `json:"password_hash"`
			FailedLoginCount int     `json:"failed_login_count"`
			LockedUntil      *string `json:"locked_until"`
		}
		if err := json.Unmarshal(raw, &payload); err != nil {
			return err
		}
		if payload.Status != "active" {
			return ErrOperatorNotFound
		}
		out.Operator = Operator{ID: payload.ID, Email: payload.Email}
		if payload.PasswordHash != nil {
			out.PasswordHash = *payload.PasswordHash
		}
		out.FailedLoginCount = payload.FailedLoginCount
		if payload.LockedUntil != nil {
			out.LockedUntil = payload.LockedUntil
		}
		return nil
	})
	return out, err
}

func (r *repository) CreateOperator(ctx context.Context, id, email string) (bool, error) {
	created := false
	err := r.db.WithSystem(ctx, func(tx *sql.Tx) error {
		rows, err := sqlc.New(tx).CreateOperator(ctx, sqlc.CreateOperatorParams{ID: id, Email: email})
		if err != nil {
			return err
		}
		created = rows > 0
		return nil
	})
	return created, err
}

func (r *repository) ListOrganizations(ctx context.Context, page Page) ([]tenancy.Organization, error) {
	var out []tenancy.Organization
	err := r.db.WithSystem(ctx, func(tx *sql.Tx) error {
		rows, err := sqlc.New(tx).ListOrganizations(ctx, sqlc.ListOrganizationsParams{
			RowLimit:  int32(page.Limit),  // #nosec G115 -- bounded by normalizePage
			RowOffset: int32(page.Offset), // #nosec G115 -- bounded by normalizePage
		})
		if err != nil {
			return err
		}
		out = make([]tenancy.Organization, 0, len(rows))
		for _, row := range rows {
			out = append(out, toOrganization(row))
		}
		return nil
	})
	return out, err
}

func (r *repository) CreateOrganization(ctx context.Context, id string, in CreateOrganizationInput,
	entry audit.Entry,
) (tenancy.Organization, error) {
	var out tenancy.Organization
	err := r.db.WithSystem(ctx, func(tx *sql.Tx) error {
		queries := sqlc.New(tx)
		err := queries.CreateOrganization(ctx, sqlc.CreateOrganizationParams{
			ID:                id,
			Name:              in.Name,
			DefaultSenderName: in.DefaultSenderName,
			DefaultReplyTo:    nullString(in.DefaultReplyTo),
			TimeZone:          in.TimeZone,
			RequireApproval:   false,
			RetentionDays:     defaultRetentionDays,
			StorageQuotaBytes: defaultStorageQuota,
			CapsEnabled:       false,
		})
		if err != nil {
			return err
		}
		row, err := queries.GetOrganization(ctx, id)
		if err != nil {
			return err
		}
		out = toOrganization(row)
		return audit.Append(ctx, tx, entry)
	})
	return out, err
}

func (r *repository) SetOrganizationStatus(ctx context.Context, orgID, expected, next string,
	entry audit.Entry,
) (bool, error) {
	changed := false
	err := r.db.WithSystem(ctx, func(tx *sql.Tx) error {
		queries := sqlc.New(tx)
		rows, err := queries.SetOrganizationStatus(ctx, sqlc.SetOrganizationStatusParams{
			ID:       orgID,
			Status:   next,
			Expected: expected,
		})
		if err != nil {
			return err
		}
		changed = rows > 0
		if !changed {
			return nil
		}
		return audit.Append(ctx, tx, entry)
	})
	return changed, err
}

func (r *repository) UpdateOrganizationLimits(ctx context.Context, orgID string, in LimitsInput,
	entry audit.Entry,
) (tenancy.Organization, error) {
	var out tenancy.Organization
	err := r.db.WithSystem(ctx, func(tx *sql.Tx) error {
		queries := sqlc.New(tx)
		row, err := queries.SetOrganizationLimits(ctx, sqlc.SetOrganizationLimitsParams{
			ID:                   orgID,
			StorageQuotaBytes:    in.StorageQuotaBytes,
			CapsEnabled:          in.CapsEnabled,
			DailyEmailCap:        nullInt32(in.DailyEmailCap),
			CampaignRecipientCap: nullInt32(in.CampaignRecipientCap),
		})
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		out = toOrganization(row)
		return audit.Append(ctx, tx, entry)
	})
	return out, err
}

func (r *repository) ListOperatorAudit(ctx context.Context, page Page) ([]audit.Entry, error) {
	var out []audit.Entry
	err := r.db.WithSystem(ctx, func(tx *sql.Tx) error {
		rows, err := sqlc.New(tx).ListOperatorAuditEntries(ctx, sqlc.ListOperatorAuditEntriesParams{
			RowLimit:  int32(page.Limit),  // #nosec G115 -- bounded by normalizePage
			RowOffset: int32(page.Offset), // #nosec G115 -- bounded by normalizePage
		})
		if err != nil {
			return err
		}
		out = make([]audit.Entry, 0, len(rows))
		for _, row := range rows {
			out = append(out, toAuditEntry(row))
		}
		return nil
	})
	return out, err
}

func toOrganization(row sqlc.Organization) tenancy.Organization {
	o := tenancy.Organization{
		ID: row.ID, Name: row.Name,
		DefaultSenderName: row.DefaultSenderName, TimeZone: row.TimeZone,
		Status: row.Status, RequireApproval: row.RequireApproval,
		RetentionDays: int(row.RetentionDays), StorageQuotaBytes: row.StorageQuotaBytes,
		CapsEnabled: row.CapsEnabled, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
	if row.LogoObjectKey.Valid {
		o.LogoObjectKey = row.LogoObjectKey.String
	}
	if row.DefaultReplyTo.Valid {
		o.DefaultReplyTo = row.DefaultReplyTo.String
	}
	if row.DailyEmailCap.Valid {
		o.DailyEmailCap = int(row.DailyEmailCap.Int32)
		o.HasDailyEmailCap = true
	}
	if row.CampaignRecipientCap.Valid {
		o.CampaignRecipientCap = int(row.CampaignRecipientCap.Int32)
		o.HasCampaignRecipientCap = true
	}
	return o
}

func toAuditEntry(row sqlc.AuditLog) audit.Entry {
	e := audit.Entry{
		ID: row.ID, Action: row.Action, Outcome: row.Outcome, At: row.CreatedAt,
	}
	if row.OrganizationID.Valid {
		e.OrganizationID = row.OrganizationID.String
	}
	e.ActorType = row.ActorType
	if row.ActorID.Valid {
		e.ActorID = row.ActorID.String
	}
	if row.ActorRole.Valid {
		e.ActorRole = row.ActorRole.String
	}
	if row.TargetType.Valid {
		e.TargetType = row.TargetType.String
	}
	if row.TargetID.Valid {
		e.TargetID = row.TargetID.String
	}
	if row.SourceIp.Valid {
		e.SourceIP = row.SourceIp.String
	}
	if row.Client.Valid {
		e.Client = row.Client.String
	}
	_ = json.Unmarshal(row.Details, &e.Details) //nolint:errcheck // details are app-written JSON
	return e
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullInt32(n int) sql.NullInt32 {
	return sql.NullInt32{Int32: int32(n), Valid: n > 0} // #nosec G115 -- caps are small positive values
}
