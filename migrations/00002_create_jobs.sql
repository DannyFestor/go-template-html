-- +goose Up
CREATE TABLE jobs (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  kind text NOT NULL,
  args jsonb NOT NULL,
  state text NOT NULL DEFAULT 'available' CHECK (state IN ('available', 'running', 'failed')),
  attempt integer NOT NULL DEFAULT 0,
  run_at timestamptz NOT NULL DEFAULT now(),
  locked_until timestamptz,
  last_error text,
  failed_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX jobs_available_run_at_idx ON jobs (run_at)
WHERE
  state = 'available';

CREATE INDEX jobs_running_locked_until_idx ON jobs (locked_until)
WHERE
  state = 'running';
