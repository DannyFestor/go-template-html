-- name: CreateUser :exec
INSERT INTO
  users (id, name, email, password, created_at)
VALUES
  (@id, @name, @email, @password, @created_at);

-- name: FindUser :one
SELECT
  id,
  name,
  email,
  password,
  created_at
FROM
  users
WHERE
  id = @id;
