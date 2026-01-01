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
