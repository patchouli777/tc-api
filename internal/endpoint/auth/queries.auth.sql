-- name: AuthSelectUser :one
SELECT
    id,
    app_role
FROM
    tc_user
WHERE
    name = $1 AND password = $2;


-- name: AuthInsertUser :exec
INSERT INTO
    tc_user(name, password)
VALUES
    ($1, $2);

