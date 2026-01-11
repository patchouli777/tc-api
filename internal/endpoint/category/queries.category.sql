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
    tc_category(name, link, image)
VALUES
    ($1, $2, $3)
RETURNING *;


-- name: CategoryUpdate :exec
UPDATE tc_category
SET
    name = CASE WHEN @name_do_update::boolean THEN @name ELSE name END,
    link = CASE WHEN @link_do_update::boolean THEN @link ELSE link END,
    is_safe = CASE WHEN @is_safe_do_update::boolean THEN @is_safe ELSE is_safe END,
    image = CASE WHEN @image_do_update::boolean THEN @image ELSE image END
WHERE
    id = @id
RETURNING *;


-- name: CategoryDeleteTags :exec
DELETE FROM
    tc_category_tag
WHERE
    id_category = $1;


-- name: CategoryAddTags :many
WITH inserted AS (
INSERT INTO
    tc_category_tag (id_category, id_tag)
    SELECT
        $1::int,
        UNNEST($2::int[])
    ON CONFLICT
        (id_category, id_tag)
    DO NOTHING
    RETURNING *
)
SELECT
    inserted.id_tag AS tag_id,
    name AS tag_name
FROM
    inserted
INNER JOIN
    tc_tag
ON
    inserted.id_tag = tc_tag.id
WHERE
    id_tag = inserted.id_tag;


-- name: CategoryDelete :exec
DELETE FROM
    tc_category
WHERE
    id = $1;
