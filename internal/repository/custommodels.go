package repository

import (
	"database/sql"
)

type PostWithTags struct {
	ID            int64
	Title         string
	Content       string
	Toc           string
	ParsedContent string
	Slug          string
	Readtime      sql.NullInt64
	CreatedAt     sql.NullTime
	ModifiedAt    sql.NullTime
	Description   sql.NullString
	Tags          []Tag
}
