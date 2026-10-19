-- name: AddRecoveryCodes :exec
INSERT INTO
  recovery_codes (user_id, digest, encrypted_code)
SELECT
  sqlc.arg(user_id)::uuid,
  unnest(sqlc.arg(digests)::bytea[]),
  unnest(sqlc.arg(encrypted_codes)::bytea[]);

-- name: ListRecoveryCodes :many
SELECT
  encrypted_code
FROM
  recovery_codes
WHERE
  user_id = @user_id
ORDER BY
  digest;

-- name: ConsumeRecoveryCode :execrows
DELETE FROM recovery_codes
WHERE
  user_id = @user_id
  AND digest = @digest;

-- name: DeleteRecoveryCodes :exec
DELETE FROM recovery_codes
WHERE
  user_id = @user_id;
