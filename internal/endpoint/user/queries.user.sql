-- name: UserSelectById :one
SELECT
    id,
    name,
    avatar,
    is_banned,
    is_partner,
    first_livestream,
    last_livestream
FROM
    tc_user
WHERE
    id = $1;


-- name: UserInsert :one
INSERT INTO tc_user (
    name,
    password,
    avatar)
VALUES (
    $1,
    $2,
    $3)
RETURNING
    id;


-- name: UserUpdateById :exec
UPDATE
    tc_user
SET
    name       = $1,
    password   = $2,
    is_banned  = $3,
    is_partner = $4,
    avatar     = $5,
    updated_at = CURRENT_DATE
WHERE
    id = $6;


-- name: UserDeleteById :exec
DELETE FROM
    tc_user
WHERE
    id = $1;
