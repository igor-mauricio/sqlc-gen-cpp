CREATE TABLE IF NOT EXISTS post (
    id INTEGER PRIMARY KEY, -- great
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    parent_id INTEGER
);
