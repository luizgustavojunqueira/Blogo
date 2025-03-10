CREATE TABLE posts (
    id BIGSERIAL PRIMARY KEY,
    title text not null,
    content TEXT not null,
    toc TEXT not null,
    parsed_content TEXT not null,
    slug text not null,
    created_at timestamp,
    modified_at timestamp 
);

