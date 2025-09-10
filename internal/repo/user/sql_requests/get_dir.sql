SELECT
    file_id,
    filename,
    dir
FROM
    files
    JOIN file_access USING (file_id)
WHERE
    (
        public = true
    )
    AND parent_id = $1