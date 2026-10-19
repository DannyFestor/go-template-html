-- +goose Up
CREATE TABLE recovery_codes (
  user_id uuid NOT NULL,
  digest bytea NOT NULL,
  encrypted_code bytea NOT NULL,
  PRIMARY KEY (user_id, digest)
);
