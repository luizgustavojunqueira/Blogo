package handlers

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/luizgustavojunqueira/Blogo/internal/repository"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/toc"
)

func validatePost(title, content, slug string) error {
	if title == "" || content == "" || slug == "" {
		return fmt.Errorf("title, content and slug are required")
	}

	if len(title) > 40 {
		return fmt.Errorf("title must be less than 40 characters")
	} else if len(title) < 5 {
		return fmt.Errorf("title must be more than 5 characters")
	}

	if len(slug) > 50 {
		return fmt.Errorf("slug must be less than 50 characters")
	} else if len(slug) < 5 {
		return fmt.Errorf("slug must be more than 5 characters")
	}

	if len(content) > 10000 {
		return fmt.Errorf("content must be less than 10000 characters")
	}
	return nil
}

func getPostToc(md goldmark.Markdown, src []byte) (string, error) {
	doc := md.Parser().Parse(text.NewReader(src))
	tree, err := toc.Inspect(doc, src)
	if err != nil {
		return "", err
	}

	list := toc.RenderList(tree)
	var tocBuf bytes.Buffer
	if list != nil {
		if err := md.Renderer().Render(&tocBuf, src, list); err != nil {
			return "", err
		}
	}
	return tocBuf.String(), nil
}

func (h *PostHandler) isAuthenticated(r *http.Request) bool {
	cookie, err := r.Cookie(h.auth.GetCookieName())

	authenticated := false

	if cookie == nil {
		h.logger.Println("Cookie is nil")
		return authenticated
	}

	if err != nil {
		h.logger.Println("Error getting cookie:", err)
		return authenticated
	}

	authenticated, err = h.auth.ValidateToken(cookie.Value)
	if err != nil {
		h.logger.Println(err)
	}

	if authenticated {
		h.logger.Println("Authenticated")
	} else {
		h.logger.Println("Not authenticated")
	}

	return authenticated
}

func generatePostsWithTags(rows []repository.GetPostsByTagRow) []repository.PostWithTags {
	result := make([]repository.PostWithTags, 0, len(rows))
	var currentPost *repository.PostWithTags

	for _, row := range rows {
		// Se é o primeiro post ou mudou o ID do anterior para o atual, cria um novo item

		if currentPost == nil || currentPost.ID != row.ID {
			currentPost = &repository.PostWithTags{
				ID:            row.ID,
				Title:         row.Title,
				Content:       row.Content,
				Toc:           row.Toc,
				ParsedContent: row.ParsedContent,
				Slug:          row.Slug,
				Description:   row.Description,
				Readtime:      row.Readtime,
				CreatedAt:     row.CreatedAt,
				ModifiedAt:    row.ModifiedAt,
				Tags:          []repository.Tag{},
			}
			result = append(result, *currentPost)
		}

		if row.TagID.Valid && row.TagName.Valid && row.TagCreatedAt.Valid && row.TagModifiedAt.Valid {
			lastPostIndex := len(result) - 1
			result[lastPostIndex].Tags = append(result[lastPostIndex].Tags, repository.Tag{
				ID:         row.TagID.Int64,
				Name:       row.TagName.String,
				CreatedAt:  row.TagCreatedAt,
				ModifiedAt: row.TagModifiedAt,
			})
		}
	}

	return result
}
