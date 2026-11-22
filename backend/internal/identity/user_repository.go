package identity

import (
	"context"
	"database/sql"
	"time"

	"github.com/TaushifReza/email-platform/backend/internal/audit"
	"github.com/TaushifReza/email-platform/backend/internal/platform/db"
	"github.com/TaushifReza/email-platform/backend/internal/platform/db/sqlc"
)

type userRepository struct{ db *db.DB }

func NewUserRepository(database *db.DB) UserRepository {
	return &userRepository{db: database}
}

func (r *userRepository) List(ctx context.Context, orgID string, limit, offset int) ([]User, error) {
	var out []User
	err := r.db.WithOrg(ctx, orgID, func(tx *sql.Tx) error {
		rows, err := sqlc.New(tx).ListUsers(ctx, sqlc.ListUsersParams{
			OrganizationID: orgID,
			RowLimit:       int32(limit),  // #nosec G115
			RowOffset:      int32(offset), // #nosec G115
		})
		if err != nil {
			return err
		}
		for _, row := range rows {
			out = append(out, User{ID: row.ID, Email: row.Email, DisplayName: row.DisplayName, Role: row.Role, Status: row.Status, CreatedAt: row.CreatedAt}) //nolint:lll
		}
		return nil
	})
	return out, err
}

func (r *userRepository) Get(ctx context.Context, orgID, userID string) (User, error) {
	var out User
	err := r.db.WithOrg(ctx, orgID, func(tx *sql.Tx) error {
		row, err := sqlc.New(tx).GetUser(ctx, sqlc.GetUserParams{OrganizationID: orgID, ID: userID})
		if err != nil {
			return err
		}
		out = User{ID: row.ID, Email: row.Email, DisplayName: row.DisplayName, Role: row.Role, Status: row.Status, CreatedAt: row.CreatedAt} //nolint:lll
		return nil
	})
	return out, err
}

func (r *userRepository) CreateInvited(ctx context.Context, orgID, userID, email, displayName, role string, now time.Time, entry audit.Entry) error { //nolint:lll
	return r.db.WithOrg(ctx, orgID, func(tx *sql.Tx) error {
		if err := sqlc.New(tx).CreateInvitedUser(ctx, sqlc.CreateInvitedUserParams{
			ID:             userID,
			OrganizationID: orgID,
			Email:          email,
			DisplayName:    displayName,
			Role:           role,
		}); err != nil {
			return err
		}
		if entry.At.IsZero() {
			entry.At = now
		}
		return audit.Append(ctx, tx, entry)
	})
}

func (r *userRepository) ChangeRole(ctx context.Context, orgID, userID, role string, _ time.Time, entry audit.Entry) error { //nolint:lll
	return r.db.WithOrg(ctx, orgID, func(tx *sql.Tx) error {
		rows, err := sqlc.New(tx).ChangeUserRole(ctx, sqlc.ChangeUserRoleParams{
			OrganizationID: orgID,
			ID:             userID,
			Role:           role,
		})
		if err != nil {
			return err
		}
		if rows == 0 {
			return ErrUserNotFound
		}
		return audit.Append(ctx, tx, entry)
	})
}

func (r *userRepository) Deactivate(ctx context.Context, orgID, userID string, _ time.Time, entry audit.Entry) error {
	return r.db.WithOrg(ctx, orgID, func(tx *sql.Tx) error {
		rows, err := sqlc.New(tx).DeactivateUser(ctx, sqlc.DeactivateUserParams{
			OrganizationID: orgID,
			ID:             userID,
		})
		if err != nil {
			return err
		}
		if rows == 0 {
			return ErrUserNotFound
		}
		return audit.Append(ctx, tx, entry)
	})
}

func (r *userRepository) LockActiveOwners(ctx context.Context, orgID string) (int, error) {
	var n int
	err := r.db.WithOrg(ctx, orgID, func(tx *sql.Tx) error {
		ids, err := sqlc.New(tx).LockActiveOwners(ctx, orgID)
		if err != nil {
			return err
		}
		n = len(ids)
		return nil
	})
	return n, err
}
