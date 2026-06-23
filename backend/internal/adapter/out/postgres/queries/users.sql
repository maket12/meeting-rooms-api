-- name: CreateUser :one
INSERT INTO users (
    id,
    email,
    password_hash,
    role,
    created_at
) VALUES (
    @id, @email, @password_hash, @role, @created_at
)
RETURNING *;

-- name: GetUserByID :one
SELECT
    id,
    email,
    password_hash,
    role,
    created_at
FROM users
WHERE id = @id;

-- name: GetUserByEmail :one
SELECT
    id,
    email,
    password_hash,
    role,
    created_at
FROM users
WHERE email = @email;

-- name: EnsureDummyUsers :exec
INSERT INTO users (id, email, password_hash, role, created_at)
VALUES
    (@admin_id, 'admin@avito.com', 'password', 'admin', (NOW() AT TIME ZONE 'UTC')),
    (@user_id, 'user@avito.com', 'password', 'user', (NOW() AT TIME ZONE 'UTC'))
ON CONFLICT (id) DO UPDATE SET
    email = EXCLUDED.email,
    role = EXCLUDED.role;
