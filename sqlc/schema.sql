create table if not exists post (
    id integer primary key,
    title text not null,
    content text not null,
    parent_id integer
) strict;
