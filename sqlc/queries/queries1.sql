-- name: getReplyIds :many
SELECT id FROM post WHERE parent_id = ?;

-- name: getAllPosts :many
SELECT
    id, title, content, parent_id
FROM post;

-- name: updatePost :exec
UPDATE post
SET title = ?, content = ? WHERE id = ?;

-- name: createPost :exec
INSERT INTO post (title, content, parent_id) VALUES (?, ?, ?);
