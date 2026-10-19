-- name: SaveTwoFactorSecret :exec
INSERT INTO
  two_factor_secrets (user_id, encrypted_secret)
VALUES
  (@user_id, @encrypted_secret)
ON CONFLICT (user_id) DO UPDATE
SET
  encrypted_secret = excluded.encrypted_secret,
  last_used_step = excluded.last_used_step;

-- name: FindTwoFactorSecret :one
SELECT
  encrypted_secret
FROM
  two_factor_secrets
WHERE
  user_id = @user_id;

-- name: UseTwoFactorStep :execrows
UPDATE two_factor_secrets
SET
  last_used_step = @step
WHERE
  user_id = @user_id
  AND last_used_step < @step;

-- name: DeleteTwoFactorSecret :exec
DELETE FROM two_factor_secrets
WHERE
  user_id = @user_id;
