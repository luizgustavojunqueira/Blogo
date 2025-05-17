-- name: CreateTagIfNotExists :many
with
    ins as (
        insert into tags(name, created_at, modified_at)
        values ($1, $2, $3) on conflict(name) do nothing
        returning *
    )
select *
from ins
union all
select *
from tags
where name = $1
limit 1
;


-- name: AddTagToPost :exec
insert into tags_posts (tag_id, post_id, created_at, modified_at)
values ($1, $2, $3, $4)
on conflict (tag_id, post_id) do nothing
;

-- name: RemoveTagFromPost :exec
delete from tags_posts
where tag_id = $1 and post_id = $2
;

-- name: GetTagsByPost :many
select t.*
from tags t
join tags_posts tp on t.id = tp.tag_id
join posts p on p.id = tp.post_id
where p.slug = $1
order by t.name
;

-- name: GetTags :many
select *
from tags
order by name
;

-- name: SearchTags :many
select name
from tags
where name ilike $1 || '%'
order by name
limit 10
;

-- name: ClearPostTagsBySlug :exec
delete from tags_posts
where post_id = (select id from posts where slug = $1)
;

