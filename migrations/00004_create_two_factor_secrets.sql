-- +goose Up
CREATE TABLE two_factor_secrets (
  user_id uuid PRIMARY KEY,
  encrypted_secret bytea NOT NULL,
  last_used_step bigint NOT NULL DEFAULT 0
);
