-- name: FollowSelectUserId :one
SELECT
    id
FROM
    tc_user u
WHERE
    u.name = $1;


-- name: FollowSelect :one
SELECT
    u1.id AS follower_id,
    u1.name AS follower_name,
    u2.id AS following_id,
    u2.name AS following_name
FROM
    tc_user u1
INNER JOIN
    tc_user_follow f ON u1.id = f.id_user
INNER JOIN
    tc_user u2 ON u2.id = f.id_follow
WHERE
    u1.name = $1 AND u2.name = $2;


-- name: FollowInsert :exec
INSERT INTO tc_user_follow(id_user, id_follow)
SELECT
  (SELECT id FROM tc_user WHERE name = $1),
  (SELECT id FROM tc_user WHERE name = $2);


-- name: FollowDelete :exec
DELETE FROM
    tc_user_follow f
USING
    tc_user u1, tc_user u2
WHERE
    f.id_user = u1.id
AND f.id_follow = u2.id
AND u1.name = $1
AND u2.name = $2;


-- name: FollowSelectMany :many
SELECT
    u1.name,
    u1.avatar
FROM
    tc_user_follow f
INNER JOIN
    tc_user u1
ON
    u1.id = f.id_follow
INNER JOIN
    tc_user u2
ON
    u2.id = f.id_user
WHERE
    u2.name = $1;


-- name: FollowSelectManyExtended :many
SELECT
    u.name AS username,
    f.name AS following,
    f.is_live,
    f.avatar AS avatar,
    ls.viewers,
    ls.title,
    c.name AS category
FROM
    tc_user_follow uf
JOIN
    tc_user u ON uf.id_user = u.id
LEFT JOIN
    tc_user f ON uf.id_follow = f.id
LEFT JOIN
    tc_livestream ls ON uf.id_follow = ls.id_user
LEFT JOIN
    tc_category c ON ls.id_category = c.id
WHERE
    u.name = $1
ORDER BY
    f.is_live DESC,
    ls.viewers DESC
LIMIT 50;
