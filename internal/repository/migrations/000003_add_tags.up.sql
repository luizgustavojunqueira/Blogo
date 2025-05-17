CREATE table tags (
    id BIGSERIAL PRIMARY KEY,
    name text not null,
    created_at timestamp,
    modified_at timestamp
);

alter table tags
    add constraint tags_name_unique unique (name);

CREATE INDEX tags_name_idx ON tags (name);

create table tags_posts (
    tag_id bigint not null,
    post_id bigint not null,
    created_at timestamp,
    modified_at timestamp,
    primary key (tag_id, post_id),
    foreign key (tag_id) references tags(id) on delete cascade,
    foreign key (post_id) references posts(id) on delete cascade
);

