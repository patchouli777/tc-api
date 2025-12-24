-- name: LivestreamSelectCategoryLinkById :one
SELECT
    link
FROM
    tc_category
WHERE
    id = $1;


-- name: LivestreamSelectMany :many
SELECT
    c.id AS cat_id,
    c.name AS cat_name,
    c.is_safe AS cat_is_safe,
    c.viewers AS cat_viewers,
    c.image AS cat_image,
    u.id AS u_id,
    u.name AS u_name,
    u.is_partner AS u_is_partner,
    u.avatar AS u_avatar
FROM
    tc_livestream l
INNER JOIN
    tc_user u ON u.id = l.id_user
INNER JOIN
    tc_category c ON c.id = l.id_category
LIMIT $1
OFFSET $2;


-- name: LivestreamSelectManyFromCategory :many
SELECT
    c.id AS cat_id,
    c.name AS cat_name,
    c.is_safe AS cat_is_safe,
    c.viewers AS cat_viewers,
    c.image AS cat_image,
    u.id AS u_id,
    u.name AS u_name,
    u.is_partner AS u_is_partner,
    u.avatar AS u_avatar
FROM
    tc_livestream l
INNER JOIN
    tc_user u ON u.id = l.id_user
INNER JOIN
    tc_category c ON c.id = l.id_category
WHERE
    c.name = $1
LIMIT $2
OFFSET $3;


-- name: LivestreamSelectById :one
SELECT c.name AS cat_name,
    c.is_safe AS cat_is_safe,
    c.link AS cat_link,
    c.viewers AS cat_viewers,
    c.image AS cat_image,
    l.title AS ls_title,
    u.id AS u_id,
    u.name AS u_name,
    u.is_partner AS u_is_partner,
    u.avatar AS u_avatar,
    u.last_livestream AS u_last_livestream,
    u.description AS u_description,
    u.links AS u_links
FROM
    tc_livestream l
INNER JOIN
    tc_user u ON u.id = l.id_user
INNER JOIN
    tc_category c ON c.id = l.id_category
WHERE
    c.name = $1;


-- name: LivestreamSelectUser :one
SELECT
    u.avatar AS u_avatar,
    u.is_banned AS u_is_banned,
    u.is_partner AS u_is_partner
FROM
    tc_user u
WHERE
    u.name = $1;


-- name: LivestreamSelectUsersDetails :many
SELECT
    u.id,
    u.name,
    u.avatar
FROM
    tc_user u
WHERE
    u.name IN (sqlc.slice('usernames'));


-- name: LivestreamSelectUserDetails :one
SELECT
    u.avatar
FROM
    tc_user u
WHERE
    u.name = $1;


-- name: LivestreamUpdate :one
WITH updated AS (
    UPDATE tc_livestream
    SET
        title = $1,
        viewers = $2,
        id_category = (SELECT id FROM tc_category c WHERE c.link = $3)
    WHERE
        id_user = (SELECT id FROM tc_user u WHERE u.name = $4)
        RETURNING *
    )
SELECT
    u.avatar as user_avatar,
    u.name AS user_name,
    c.link AS category_link,
    c.name AS category_name
FROM
    updated
JOIN
    tc_user u ON updated.id_user = u.id
JOIN
    tc_category c ON updated.id_category = c.id;


-- name: LivestreamInsert :one
WITH inserted AS (
    INSERT INTO tc_livestream (
        id_user,
        id_category,
        is_multistream,
        title
    ) VALUES (
        (SELECT id FROM tc_user u WHERE u.name = $1),
        (SELECT id FROM tc_category c WHERE c.link = $2),
        FALSE,
        $3
    )
    RETURNING *
)
SELECT
    inserted.id AS livestream_id,
    u.avatar AS user_avatar,
    u.name AS user_name,
    c.link AS category_link,
    c.name AS category_name,
    started_at
FROM
    inserted
JOIN
    tc_user u ON inserted.id_user = u.id
JOIN
    tc_category c ON inserted.id_category = c.id;


-- name: LivestreamDelete :exec
DELETE FROM
    tc_livestream l
WHERE
    l.id_user = (
        SELECT
            id
        FROM
            tc_user
        WHERE
            name = $1
    );


-- name: LivestreamUpdateViewers :exec
UPDATE tc_livestream
    SET
        viewers = $1
    FROM
        tc_user
    WHERE
        tc_livestream.id_user = tc_user.id
    AND
        tc_user.name = $2;
