-- name: CategorySelectMany :many
SELECT
    c.id AS category_id,
    c.name AS category_name,
    is_safe,
    viewers,
    image,
    t.id AS tag_id,
    t.name AS tag_name
FROM
    tc_category c
LEFT OUTER JOIN
    tc_category_tag ct ON ct.id_category = c.id
LEFT OUTER JOIN
    tc_tag t ON t.id = ct.id_tag
LIMIT $1
OFFSET $2;


-- name: CategorySelect :one
SELECT
    *
FROM
    tc_category
WHERE
    id = $1
LIMIT 1;


-- name: CategoryInsert :one
INSERT INTO
    tc_category(name, link, viewers, image)
VALUES
    ($1, $2, $3, $4)
RETURNING *;
