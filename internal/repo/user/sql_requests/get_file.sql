SELECT 
    f.filename,
    ARRAY_AGG(b.block_id) AS blocks,
    ARRAY_AGG(b.language) AS languages
FROM 
    files f
LEFT JOIN 
    blocks b ON f.file_id = b.file_id
WHERE 
    f.file_id = $1
GROUP BY 
    f.file_id;