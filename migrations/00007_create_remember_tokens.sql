-- +goose Up
CREATE TABLE remember_tokens (
  selector text PRIMARY KEY,
  validator_digest bytea NOT NULL,
  user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  expires_at timestamptz NOT NULL
);

CREATE INDEX remember_tokens_user_id_idx ON remember_tokens (user_id);

CREATE INDEX remember_tokens_expires_at_idx ON remember_tokens (expires_at);
