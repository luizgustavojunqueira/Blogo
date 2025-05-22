-- name: GetPosts :many
select *
from posts
order by created_at desc
;

-- name: CreatePost :one
insert into posts (title, toc, content, parsed_content, description, slug, created_at, modified_at, readtime)
values (:title, :toc, :content, :parsed_content, :description, :slug, :created_at, :modified_at, :readtime)
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

-- name: UpdatePostBySlug :one
update posts
set title = :title, toc = :toc, slug = :new_slug, content = :content, parsed_content = :parsed_content, modified_at = :modified_at, description = :description, readtime = :readtime
where slug = :slug
returning *
;

-- name: GetPostsByTag :many
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
    cast(sqlc.narg('tag_name') as text) is null
    or p.id in (
        select tp2.post_id
        from tags_posts tp2
        join tags t2 on tp2.tag_id = t2.id
        where t2.name = sqlc.narg('tag_name')
    )
order by p.created_at desc, p.id, t.id
;

