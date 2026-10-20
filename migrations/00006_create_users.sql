-- +goose Up
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  name text NOT NULL,
  email citext NOT NULL UNIQUE,
  password text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE two_factor_secrets
ADD FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE;

ALTER TABLE recovery_codes
ADD FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE;
