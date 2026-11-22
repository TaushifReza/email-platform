package tenancy

import (
	"context"
	"database/sql"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/audit"
	"github.com/TaushifReza/email-platform/backend/internal/platform/db"
	"github.com/TaushifReza/email-platform/backend/internal/platform/db/sqlc"
)

type repository struct{ db *db.DB }

func NewRepository(database *db.DB) OrganizationRepository {
	return &repository{db: database}
}

func (r *repository) Get(ctx context.Context, orgID string) (Organization, error) {
	var out Organization
	err := r.db.WithOrg(ctx, orgID, func(tx *sql.Tx) error {
		row, err := sqlc.New(tx).GetOrganization(ctx, orgID)
		if err != nil {
			return err
		}
		out = toOrg(row)
		return nil
	})
	return out, err
}

func (r *repository) UpdateProfile(ctx context.Context, orgID string, p ProfileUpdate, now time.Time, entry audit.Entry) (Organization, error) { //nolint:lll
	var out Organization
	err := r.db.WithOrg(ctx, orgID, func(tx *sql.Tx) error {
		row, err := sqlc.New(tx).UpdateOrganizationProfile(ctx, sqlc.UpdateOrganizationProfileParams{
			ID:                orgID,
			Name:              p.Name,
			DefaultSenderName: p.DefaultSenderName,
			DefaultReplyTo:    p.DefaultReplyTo,
			TimeZone:          p.TimeZone,
		})
		if err != nil {
			return err
		}
		out = toOrg(row)
		if entry.At.IsZero() {
			entry.At = now
		}
		return audit.Append(ctx, tx, entry)
	})
	return out, err
}

func (r *repository) UpdatePolicy(ctx context.Context, orgID string, p PolicyUpdate, now time.Time, entry audit.Entry) (Organization, error) { //nolint:lll
	var out Organization
	err := r.db.WithOrg(ctx, orgID, func(tx *sql.Tx) error {
		row, err := sqlc.New(tx).UpdateOrganizationPolicy(ctx, sqlc.UpdateOrganizationPolicyParams{
			ID:              orgID,
			RequireApproval: p.RequireApproval,
			RetentionDays:   int32(p.RetentionDays), // #nosec G115
		})
		if err != nil {
			return err
		}
		out = toOrg(row)
		if entry.At.IsZero() {
			entry.At = now
		}
		return audit.Append(ctx, tx, entry)
	})
	return out, err
}

func toOrg(row sqlc.Organization) Organization {
	o := Organization{
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
