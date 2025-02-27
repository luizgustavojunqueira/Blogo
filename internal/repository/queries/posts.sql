-- name: GetPosts :many
select *
from posts
;


-- name: CreatePost :one
insert into posts (title, content, parsed_content, slug, created_at, modified_at)
values (:title, :content, :parsed_content, :slug, :created_at, :modified_at)
returning *
;

