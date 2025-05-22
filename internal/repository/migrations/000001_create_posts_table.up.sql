CREATE TABLE posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title text not null,
    content TEXT not null,
    toc TEXT not null,
    parsed_content TEXT not null,
    slug text not null,
    created_at DATETIME,
    modified_at DATETIME 
);

