-- name: GetPosts :many
select *
from posts
order by created_at desc
;

-- name: CreatePost :one
insert into posts (title, content, parsed_content, slug, created_at, modified_at)
values (:title, :content, :parsed_content, :slug, :created_at, :modified_at)
returning *
;

-- name: GetPostBySlug :one
select *
from posts
where slug =:slug
;


-- name: DeletePostBySlug :exec
delete from posts
where slug =:slug
;

