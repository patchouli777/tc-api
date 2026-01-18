-- name: UserSelect :one
SELECT
    id,
    name,
    pfp,
    is_banned,
    is_partner,
    first_livestream,
    last_livestream
FROM
    tc_user
WHERE
    id = $1;


-- name: UserSelectByUsername :one
SELECT
    id,
    name,
    pfp,
    is_banned,
    is_partner,
    first_livestream,
    last_livestream
FROM
    tc_user
WHERE
    name = $1;


-- name: UserInsert :one
INSERT INTO tc_user (
    name,
    password,
    pfp)
VALUES (
    $1,
    $2,
    $3)
RETURNING id;


-- name: UserUpdate :exec
UPDATE
    tc_user
SET
    name       = CASE WHEN @name_do_update::boolean THEN @name ELSE name END,
    password   = CASE WHEN @password_do_update::boolean THEN @password ELSE password END,
    is_banned  = CASE WHEN @is_banned_do_update::boolean THEN @is_banned ELSE is_banned END,
    is_partner = CASE WHEN @is_partner_do_update::boolean THEN @is_partner ELSE is_partner END,
    pfp        = CASE WHEN @pfp_do_update::boolean THEN @pfp ELSE pfp END,
    updated_at = CURRENT_DATE
WHERE
    id = @id
RETURNING *;


-- name: UserDelete :exec
DELETE FROM
    tc_user
WHERE
    id = $1;
