-- name: GetPosts :many
select *
from posts
order by created_at desc
;

-- name: CreatePost :one
insert into posts (title, toc, content, parsed_content, slug, created_at, modified_at)
values ($1, $2, $3, $4, $5, $6, $7)
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
set title = $1, toc = $2, slug = $3, content = $4, parsed_content = $5, modified_at = $6
where slug = $7
;

