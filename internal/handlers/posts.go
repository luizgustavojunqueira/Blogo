package handlers

import (
	"bytes"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/luizgustavojunqueira/Blog/internal/repository"
	"github.com/luizgustavojunqueira/Blog/internal/templates"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
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

	if title == "" || content == "" || slug == "" {
		http.Error(w, "Title, content and slug are required", http.StatusBadRequest)
		return
	}

	md := goldmark.New(goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		))

	var parsedContent bytes.Buffer
	if err := md.Convert([]byte(content), &parsedContent); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	post := repository.CreatePostParams{
		Title:         title,
		Content:       content,
		ParsedContent: parsedContent.String(),
		Slug:          slug,
		CreatedAt:     sql.NullTime{Time: time.Now(), Valid: true},
		ModifiedAt:    sql.NullTime{Time: time.Now(), Valid: true},
	}

	createdPost, err := h.queries.CreatePost(ctx, post)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Location", "/")

	page := templates.PostCard(createdPost)
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

	title := r.FormValue("title")
	content := r.FormValue("content")
	slug := r.FormValue("slug")
	md := goldmark.New(goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		))

	var buf bytes.Buffer
	if err := md.Convert([]byte(content), &buf); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	parsedMakdownRendered := templates.ParsedMarkdown(buf.String(), title, slug)
	parsedMakdownRendered.Render(ctx, w)
}
