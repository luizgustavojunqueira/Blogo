package handlers

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
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

func checkAuth(r *http.Request) bool {
	cookie, err := r.Cookie("session")
	if err != nil {
		return false
	}

	return cookie.Value == "authenticated"
}

func (h *PostHandler) GetPosts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	posts, err := h.queries.GetPosts(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	authenticated := checkAuth(r)

	page := templates.MainPage(posts, authenticated)
	page.Render(ctx, w)
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	authenticated := checkAuth(r)

	if !authenticated {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

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

	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		log.Panic(err)
	}

	post := repository.CreatePostParams{
		Title:         title,
		Content:       content,
		ParsedContent: parsedContent.String(),
		Slug:          slug,
		CreatedAt:     pgtype.Timestamp{Time: time.Now().In(loc), Valid: true},
		ModifiedAt:    pgtype.Timestamp{Time: time.Now().In(loc), Valid: true},
	}

	createdPost, err := h.queries.CreatePost(ctx, post)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Location", "/")

	page := templates.PostCard(createdPost, authenticated)
	page.Render(ctx, w)
}

func (h *PostHandler) Editor(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	slug := r.PathValue("slug")

	authenticated := checkAuth(r)

	if !authenticated {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if slug != "" {

		post, err := h.queries.GetPostBySlug(ctx, slug)
		if err != nil {
			http.Error(w, fmt.Sprintf("Post not found: %s", err.Error()), http.StatusNotFound)
			return
		}

		page := templates.Editor(post.Content, post.Title, post.Slug, true)
		page.Render(ctx, w)
		return
	}

	editor := templates.Editor("", "", "", false)
	editor.Render(ctx, w)
}

func (h *PostHandler) ParseMarkdown(w http.ResponseWriter, r *http.Request) {
	authenticated := checkAuth(r)

	if !authenticated {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

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

	parsedMakdownRendered := templates.MarkdownPreview(buf.String(), title, slug)
	parsedMakdownRendered.Render(ctx, w)
}

func (h *PostHandler) ViewPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	slug := r.PathValue("slug")

	post, err := h.queries.GetPostBySlug(ctx, slug)
	if err != nil {
		http.Error(w, fmt.Sprintf("Post not found: %s", err.Error()), http.StatusNotFound)
		return
	}

	page := templates.PostPage(post)
	page.Render(ctx, w)
}

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	authenticated := checkAuth(r)

	if !authenticated {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	ctx := r.Context()

	slug := r.PathValue("slug")
	err := h.queries.DeletePostBySlug(ctx, slug)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Location", "/")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("Post deleted"))
}

func (h *PostHandler) EditPost(w http.ResponseWriter, r *http.Request) {
	authenticated := checkAuth(r)

	if !authenticated {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	ctx := r.Context()

	slug := r.PathValue("slug")

	newTitle := r.FormValue("title")
	newSlug := r.FormValue("slug")
	newContent := r.FormValue("content")

	md := goldmark.New(goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		))

	var parsedContent bytes.Buffer
	if err := md.Convert([]byte(newContent), &parsedContent); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		log.Panic(err)
	}

	post := repository.UpdatePostBySlugParams{
		Title:         newTitle,
		Slug:          slug,
		Slug_2:        newSlug,
		Content:       newContent,
		ParsedContent: parsedContent.String(),
		ModifiedAt:    pgtype.Timestamp{Time: time.Now().In(loc), Valid: true},
	}

	err = h.queries.UpdatePostBySlug(ctx, post)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Location", "/")
}
