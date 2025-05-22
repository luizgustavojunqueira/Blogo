CREATE table tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name text not null unique,
    created_at DATETIME,
    modified_at DATETIME
);

CREATE INDEX tags_name_idx ON tags (name);

create table tags_posts (
    tag_id bigint not null,
    post_id bigint not null,
    created_at DATETIME,
    modified_at DATETIME,
    primary key (tag_id, post_id),
    foreign key (tag_id) references tags(id) on delete cascade,
    foreign key (post_id) references posts(id) on delete cascade
);

