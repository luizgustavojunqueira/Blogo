package repository

import "github.com/jackc/pgx/v5/pgtype"

type PostWithTags struct {
	ID            int64
	Title         string
	Content       string
	Toc           string
	ParsedContent string
	Slug          string
	Readtime      pgtype.Int4
	CreatedAt     pgtype.Timestamp
	ModifiedAt    pgtype.Timestamp
	Description   pgtype.Text
	Tags          []Tag
}
