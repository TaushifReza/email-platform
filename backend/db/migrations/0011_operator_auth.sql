-- +goose Up
-- Allow operator-issued invitations (no tenant user acts as inviter), e.g. the
-- first Owner of a newly created organization. Composite FK remains valid
-- because MATCH SIMPLE accepts NULL in invited_by_user_id.
ALTER TABLE invitations ALTER COLUMN invited_by_user_id DROP NOT NULL;

-- +goose Down
-- expand-only: leave column nullable