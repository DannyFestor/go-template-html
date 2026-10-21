-- name: CreateUser :exec
INSERT INTO
  users (
    id,
    name,
    email,
    password,
    email_verified_at,
    created_at
  )
VALUES
  (
    @id,
    @name,
    @email,
    @password,
    @email_verified_at,
    @created_at
  );

-- name: FindUser :one
SELECT
  id,
  name,
  email,
  password,
  created_at,
  email_verified_at
FROM
  users
WHERE
  id = @id;

-- name: FindUserByEmail :one
SELECT
  id,
  name,
  email,
  password,
  created_at,
  email_verified_at
FROM
  users
WHERE
  email = @email;

-- name: UpdateUserPassword :exec
UPDATE users
SET
  password = @password
WHERE
  id = @id;

-- name: MarkUserEmailVerified :exec
UPDATE users
SET
  email_verified_at = sqlc.arg(verified_at)::timestamptz
WHERE
  id = @id
  AND email_verified_at IS NULL;
