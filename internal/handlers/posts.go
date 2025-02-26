package handlers

import (
	"Blog/internal/repository"
	"Blog/internal/templates"
	"bytes"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/yuin/goldmark"
)

type PostHandler struct {
	queries *repository.Queries
}

func NewPostHandler(queries *repository.Queries) *PostHandler {
	return &PostHandler{
		queries: queries,
	}
}

func (h *PostHandler) GetPosts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	posts, err := h.queries.GetPosts(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	page := templates.MainPage(posts)
	page.Render(ctx, w)
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	log.Println("Creating post...")
	ctx := r.Context()

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	slug := r.FormValue("slug")

	post := repository.CreatePostParams{
		Title:      title,
		Content:    content,
		Slug:       slug,
		CreatedAt:  sql.NullTime{Time: time.Now(), Valid: true},
		ModifiedAt: sql.NullTime{Time: time.Now(), Valid: true},
	}

	createdPost, err := h.queries.CreatePost(ctx, post)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	page := templates.Post(createdPost)
	page.Render(ctx, w)
}

func (h *PostHandler) Editor(w http.ResponseWriter, r *http.Request) {
	log.Println("Editor...")
	ctx := r.Context()

	editor := templates.Editor()
	editor.Render(ctx, w)
}

func (h *PostHandler) ParseMarkdown(w http.ResponseWriter, r *http.Request) {
	log.Println("Parsing markdown...")
	ctx := r.Context()

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	content := r.FormValue("content")

	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(content), &buf); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	parsedMakdownRendered := templates.ParsedMarkdown(buf.String())
	parsedMakdownRendered.Render(ctx, w)
}
