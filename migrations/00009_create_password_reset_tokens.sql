-- +goose Up
CREATE TABLE password_reset_tokens (
  user_id uuid PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
  token_digest bytea NOT NULL UNIQUE,
  expires_at timestamptz NOT NULL
);
