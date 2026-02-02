SELECT
    id
FROM
    user_entity
WHERE
    username = $1
    AND realm_id = $2
LIMIT
    1;