-- name: getImages :many
SELECT * FROM images;

-- name: createImage :exec
INSERT INTO images (title, content) VALUES (?, ?);
