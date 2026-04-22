SELECT
    id,
    username
FROM
    user_entity
WHERE
    username LIKE $1
    AND realm_id = $2
ORDER BY
    username
LIMIT
    $3;
