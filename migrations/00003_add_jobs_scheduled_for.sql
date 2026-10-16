-- +goose Up
ALTER TABLE jobs
ADD COLUMN scheduled_for timestamptz,
ADD CONSTRAINT jobs_kind_scheduled_for_key UNIQUE (kind, scheduled_for);
