-- name: GetPosts :many
select *
from posts
order by created_at desc
;

-- name: CreatePost :one
insert into posts (title, toc, content, parsed_content, description, slug, created_at, modified_at, readtime)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
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

-- name: UpdatePostBySlug :one
update posts
set title = $1, toc = $2, slug = $3, content = $4, parsed_content = $5, modified_at = $6, description = $7, readtime = $8
where slug = $9
returning *
;

-- name: GetPostsByTag :many
select p.*
from posts p
join tags_posts tp on p.id = tp.post_id
join tags t on t.id = tp.tag_id
where t.name = $1
order by p.created_at desc
;

-- name: ListPostsWithTags :many
select
    p.id,
    p.title,
    p.content,
    p.toc,
    p.parsed_content,
    p.slug,
    p.description,
    p.readtime,
    p.created_at,
    p.modified_at,
    t.id as tag_id,
    t.name as tag_name,
    t.created_at as tag_created_at,
    t.modified_at as tag_modified_at
from posts p
left join tags_posts tp on p.id = tp.post_id
left join tags t on tp.tag_id = t.id
where
    sqlc.narg('tag_name')::text is null
    or p.id in (
        select tp2.post_id
        from tags_posts tp2
        join tags t2 on tp2.tag_id = t2.id
        where t2.name = sqlc.narg('tag_name')
    )
order by p.created_at desc, p.id, t.id
;

