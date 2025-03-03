-- name: GetPosts :many
select *
from posts
order by created_at desc
;

-- name: CreatePost :one
insert into posts (title, content, parsed_content, slug, created_at, modified_at)
values ($1, $2, $3, $4, $5, $6)
returning *
;

-- name: GetPostBySlug :one
select *
from posts
where slug = $1
;


-- name: DeletePostBySlug :exec
delete from posts
where slug = $1
;


-- name: UpdatePostBySlug :exec
update posts
set title = $1, slug = $2, content = $3, parsed_content = $4, modified_at = $5
where slug = $6
;

