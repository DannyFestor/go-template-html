-- name: ReplacePasswordResetToken :exec
INSERT INTO
  password_reset_tokens (user_id, token_digest, expires_at)
VALUES
  (@user_id, @token_digest, @expires_at)
ON CONFLICT (user_id) DO UPDATE
SET
  token_digest = excluded.token_digest,
  expires_at = excluded.expires_at;

-- name: ConsumePasswordResetToken :one
DELETE FROM password_reset_tokens
WHERE
  token_digest = @token_digest
  AND expires_at > sqlc.arg(now)::timestamptz
RETURNING
  user_id;
