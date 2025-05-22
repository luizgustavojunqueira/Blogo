-- name: CreateTagIfNotExists :exec
INSERT OR IGNORE INTO tags(name, created_at, modified_at)
VALUES (:name, :created_at, :modified_at);

-- name: GetTagByName :one
select *
from tags
where name =:name
limit 1
;

-- name: AddTagToPost :exec
insert into tags_posts (tag_id, post_id, created_at, modified_at)
values (:tag_id, :post_id, :created_at, :modified_at)
on conflict (tag_id, post_id) do nothing
;


-- name: GetTagsByPost :many
select t.*
from tags t
join tags_posts tp on t.id = tp.tag_id
join posts p on p.id = tp.post_id
where p.slug =:slug
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
where name like:search || '%' collate nocase
order by name
limit 10
;

-- name: ClearPostTagsBySlug :exec
delete from tags_posts
where post_id = (select id from posts where slug =:slug)
;

