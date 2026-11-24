package providers

import (
	"context"
	"database/sql"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/audit"
	"github.com/TaushifReza/email-platform/backend/internal/platform/db"
	"github.com/TaushifReza/email-platform/backend/internal/platform/db/sqlc"
)

type repository struct{ db *db.DB }

func NewRepository(database *db.DB) Repository {
	return &repository{db: database}
}

func toProviderList(row sqlc.ListProvidersRow) Provider {
	p := Provider{
		ID: row.ID, OrganizationID: row.OrganizationID, Host: row.Host,
		Port: int(row.Port), TLSMode: row.TlsMode, Username: row.Username,
		FromAddress: row.FromAddress, IsDefault: row.IsDefault,
		Status: row.Status, DailyLimit: int(row.DailyLimit),
		MinuteLimit: int(row.MinuteLimit), HourLimit: int(row.HourLimit),
		ConfigVersion: int(row.ConfigVersion), CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
	if row.VerifiedAt.Valid {
		t := row.VerifiedAt.Time
		p.VerifiedAt = &t
	}
	return p
}

func toProviderGet(row sqlc.GetProviderRow) Provider {
	p := Provider{
		ID: row.ID, OrganizationID: row.OrganizationID, Host: row.Host,
		Port: int(row.Port), TLSMode: row.TlsMode, Username: row.Username,
		FromAddress: row.FromAddress, IsDefault: row.IsDefault,
		Status: row.Status, DailyLimit: int(row.DailyLimit),
		MinuteLimit: int(row.MinuteLimit), HourLimit: int(row.HourLimit),
		ConfigVersion: int(row.ConfigVersion), CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
	if row.VerifiedAt.Valid {
		t := row.VerifiedAt.Time
		p.VerifiedAt = &t
	}
	return p
}

func toProviderUpdate(row sqlc.UpdateProviderConfigRow) Provider {
	p := Provider{
		ID: row.ID, OrganizationID: row.OrganizationID, Host: row.Host,
		Port: int(row.Port), TLSMode: row.TlsMode, Username: row.Username,
		FromAddress: row.FromAddress, IsDefault: row.IsDefault,
		Status: row.Status, DailyLimit: int(row.DailyLimit),
		MinuteLimit: int(row.MinuteLimit), HourLimit: int(row.HourLimit),
		ConfigVersion: int(row.ConfigVersion), CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
	if row.VerifiedAt.Valid {
		t := row.VerifiedAt.Time
		p.VerifiedAt = &t
	}
	return p
}

func toProviderRotate(row sqlc.RotateProviderSecretRow) Provider {
	p := Provider{
		ID: row.ID, OrganizationID: row.OrganizationID, Host: row.Host,
		Port: int(row.Port), TLSMode: row.TlsMode, Username: row.Username,
		FromAddress: row.FromAddress, IsDefault: row.IsDefault,
		Status: row.Status, DailyLimit: int(row.DailyLimit),
		MinuteLimit: int(row.MinuteLimit), HourLimit: int(row.HourLimit),
		ConfigVersion: int(row.ConfigVersion), CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
	if row.VerifiedAt.Valid {
		t := row.VerifiedAt.Time
		p.VerifiedAt = &t
	}
	return p
}

func toProviderDeactivate(row sqlc.DeactivateProviderRow) Provider {
	p := Provider{
		ID: row.ID, OrganizationID: row.OrganizationID, Host: row.Host,
		Port: int(row.Port), TLSMode: row.TlsMode, Username: row.Username,
		FromAddress: row.FromAddress, IsDefault: row.IsDefault,
		Status: row.Status, DailyLimit: int(row.DailyLimit),
		MinuteLimit: int(row.MinuteLimit), HourLimit: int(row.HourLimit),
		ConfigVersion: int(row.ConfigVersion), CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
	if row.VerifiedAt.Valid {
		t := row.VerifiedAt.Time
		p.VerifiedAt = &t
	}
	return p
}

func toProviderSetDefault(row sqlc.SetProviderDefaultRow) Provider {
	p := Provider{
		ID: row.ID, OrganizationID: row.OrganizationID, Host: row.Host,
		Port: int(row.Port), TLSMode: row.TlsMode, Username: row.Username,
		FromAddress: row.FromAddress, IsDefault: row.IsDefault,
		Status: row.Status, DailyLimit: int(row.DailyLimit),
		MinuteLimit: int(row.MinuteLimit), HourLimit: int(row.HourLimit),
		ConfigVersion: int(row.ConfigVersion), CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
	if row.VerifiedAt.Valid {
		t := row.VerifiedAt.Time
		p.VerifiedAt = &t
	}
	return p
}

func (r *repository) List(ctx context.Context, orgID string) ([]Provider, error) {
	var out []Provider
	err := r.db.WithOrg(ctx, orgID, func(tx *sql.Tx) error {
		rows, err := sqlc.New(tx).ListProviders(ctx, orgID)
		if err != nil {
			return err
		}
		for _, row := range rows {
			out = append(out, toProviderList(row))
		}
		return nil
	})
	return out, err
}

func (r *repository) Get(ctx context.Context, orgID, id string) (Provider, error) {
	var out Provider
	err := r.db.WithOrg(ctx, orgID, func(tx *sql.Tx) error {
		row, err := sqlc.New(tx).GetProvider(ctx, sqlc.GetProviderParams{OrganizationID: orgID, ID: id})
		if err != nil {
			return err
		}
		out = toProviderGet(row)
		return nil
	})
	return out, err
}

func (r *repository) Create(ctx context.Context, orgID string, p Provider, enc []byte, keyID string, _ time.Time, entry audit.Entry) error { //nolint:lll
	return r.db.WithOrg(ctx, orgID, func(tx *sql.Tx) error {
		q := sqlc.New(tx)
		isDefault := p.IsDefault
		// first provider becomes default
		n, err := q.CountProviders(ctx, orgID)
		if err != nil {
			return err
		}
		if n == 0 {
			isDefault = true
		}
		if isDefault {
			if err := q.ClearProviderDefault(ctx, orgID); err != nil {
				return err
			}
		}
		id := p.ID
		if _, err := q.CreateProvider(ctx, sqlc.CreateProviderParams{
			ID: id, OrganizationID: orgID, Host: p.Host, Port: int32(p.Port), // #nosec G115
			TlsMode: p.TLSMode, Username: p.Username, PasswordEncrypted: enc,
			PasswordKeyID: keyID, FromAddress: p.FromAddress, IsDefault: isDefault,
			Status: p.Status, DailyLimit: int32(p.DailyLimit), MinuteLimit: int32(p.MinuteLimit), // #nosec G115
			HourLimit: int32(p.HourLimit), // #nosec G115
		}); err != nil {
			return err
		}
		return audit.Append(ctx, tx, entry)
	})
}

func (r *repository) Update(ctx context.Context, orgID, id string, u Update, _ time.Time, entry audit.Entry) (Provider, error) { //nolint:lll
	var out Provider
	err := r.db.WithOrg(ctx, orgID, func(tx *sql.Tx) error {
		q := sqlc.New(tx)
		cur, err := q.GetProvider(ctx, sqlc.GetProviderParams{OrganizationID: orgID, ID: id})
		if err != nil {
			return err
		}
		host := cur.Host
		port := cur.Port
		tls := cur.TlsMode
		from := cur.FromAddress
		daily := cur.DailyLimit
		minute := cur.MinuteLimit
		hour := cur.HourLimit
		if u.Host != nil {
			host = *u.Host
		}
		if u.Port != nil {
			port = int32(*u.Port) // #nosec G115
		}
		if u.TLSMode != nil {
			tls = *u.TLSMode
		}
		if u.FromAddress != nil {
			from = *u.FromAddress
		}
		if u.DailyLimit != nil {
			daily = int32(*u.DailyLimit) // #nosec G115
		}
		if u.MinuteLimit != nil {
			minute = int32(*u.MinuteLimit) // #nosec G115
		}
		if u.HourLimit != nil {
			hour = int32(*u.HourLimit) // #nosec G115
		}
		row, err := q.UpdateProviderConfig(ctx, sqlc.UpdateProviderConfigParams{
			OrganizationID: orgID, ID: id, Host: host, Port: port, TlsMode: tls,
			FromAddress: from, DailyLimit: daily, MinuteLimit: minute, HourLimit: hour,
		})
		if err != nil {
			return err
		}
		out = toProviderUpdate(row)
		if u.IsDefault != nil && *u.IsDefault {
			if err := q.ClearProviderDefault(ctx, orgID); err != nil {
				return err
			}
			row2, err := q.SetProviderDefault(ctx, sqlc.SetProviderDefaultParams{OrganizationID: orgID, ID: id})
			if err != nil {
				return err
			}
			out = toProviderSetDefault(row2)
		}
		return audit.Append(ctx, tx, entry)
	})
	return out, err
}

func (r *repository) Rotate(ctx context.Context, orgID, id, username string, enc []byte, keyID string, _ time.Time, entry audit.Entry) (Provider, error) { //nolint:lll
	var out Provider
	err := r.db.WithOrg(ctx, orgID, func(tx *sql.Tx) error {
		q := sqlc.New(tx)
		cur, err := q.GetProvider(ctx, sqlc.GetProviderParams{OrganizationID: orgID, ID: id})
		if err != nil {
			return err
		}
		u := cur.Username
		if username != "" {
			u = username
		}
		row, err := q.RotateProviderSecret(ctx, sqlc.RotateProviderSecretParams{
			OrganizationID: orgID, ID: id, Username: u, PasswordEncrypted: enc,
			PasswordKeyID: keyID,
		})
		if err != nil {
			return err
		}
		out = toProviderRotate(row)
		return audit.Append(ctx, tx, entry)
	})
	return out, err
}

func (r *repository) Deactivate(ctx context.Context, orgID, id string, _ time.Time, entry audit.Entry) (Provider, error) { //nolint:lll
	var out Provider
	err := r.db.WithOrg(ctx, orgID, func(tx *sql.Tx) error {
		row, err := sqlc.New(tx).DeactivateProvider(ctx, sqlc.DeactivateProviderParams{OrganizationID: orgID, ID: id})
		if err != nil {
			return err
		}
		out = toProviderDeactivate(row)
		return audit.Append(ctx, tx, entry)
	})
	return out, err
}

func (r *repository) SetDefault(ctx context.Context, orgID, id string, _ time.Time, entry audit.Entry) (Provider, error) { //nolint:lll
	var out Provider
	err := r.db.WithOrg(ctx, orgID, func(tx *sql.Tx) error {
		q := sqlc.New(tx)
		if err := q.ClearProviderDefault(ctx, orgID); err != nil {
			return err
		}
		row, err := q.SetProviderDefault(ctx, sqlc.SetProviderDefaultParams{OrganizationID: orgID, ID: id})
		if err != nil {
			return err
		}
		out = toProviderSetDefault(row)
		return audit.Append(ctx, tx, entry)
	})
	return out, err
}

func (r *repository) EncryptedCredentials(ctx context.Context, orgID, id string) (string, []byte, string, error) {
	var username string
	var ct []byte
	var keyID string
	err := r.db.WithOrg(ctx, orgID, func(tx *sql.Tx) error {
		row, err := sqlc.New(tx).GetProviderCredentials(ctx, sqlc.GetProviderCredentialsParams{OrganizationID: orgID, ID: id})
		if err != nil {
			return err
		}
		username = row.Username
		ct = row.PasswordEncrypted
		keyID = row.PasswordKeyID
		return nil
	})
	return username, ct, keyID, err
}
