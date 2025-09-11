SELECT
    username,
    first_name,
    last_name
FROM
    user_entity
WHERE
    id = $1
    AND realm_id = $2
LIMIT
    1;