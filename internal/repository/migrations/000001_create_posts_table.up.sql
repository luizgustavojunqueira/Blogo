CREATE TABLE posts (
    id BIGSERIAL PRIMARY KEY,
    title text not null,
    body text,
    created_at timestamp,
    modified_at timestamp 
);

