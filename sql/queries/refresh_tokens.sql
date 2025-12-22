-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at)
VALUES (
    $1
    ,NOW()
    ,NOW()
    ,$2
    ,NOW() + interval '60 days'
)
RETURNING *;

-- name: GetUserFromRefreshToken :one
SELECT user_id
FROM refresh_tokens
WHERE revoked_at IS NULL
    AND expires_at > NOW()
    AND token = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens 
SET (updated_at, revoked_at) = (NOW(), NOW())
WHERE token = $1;