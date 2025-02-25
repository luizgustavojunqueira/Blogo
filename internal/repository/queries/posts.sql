-- name: GetPosts :many
select *
from posts
;


-- name: CreatePost :one
insert into posts (title, body, created_at, modified_at)
values (:title, :body, :created_at, :modified_at)
returning *
;

