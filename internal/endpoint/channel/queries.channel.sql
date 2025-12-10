-- name: ChannelSelect :one
SELECT
    id,
    name,
    is_banned,
    is_partner,
    first_livestream,
    last_livestream,
    offline_background AS background,
    avatar,
    description,
    links,
    tags
FROM
    tc_user
WHERE
    name = $1;
