package handlers

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/luizgustavojunqueira/Blogo/internal/auth"
	"github.com/luizgustavojunqueira/Blogo/internal/repository"
	"github.com/luizgustavojunqueira/Blogo/internal/templates"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

type PostHandler struct {
	queries   *repository.Queries
	md        goldmark.Markdown
	location  *time.Location
	logger    *log.Logger
	auth      *auth.Auth
	blogName  string
	pagetitle string
}

func NewPostHandler(queries *repository.Queries, location *time.Location, logger *log.Logger, auth *auth.Auth, blogName, pagetitle string) *PostHandler {
	md := goldmark.New(goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		))

	return &PostHandler{
		queries:   queries,
		md:        md,
		logger:    logger,
		location:  location,
		auth:      auth,
		blogName:  blogName,
		pagetitle: pagetitle,
	}
}

func validatePost(title, content, slug string) error {
	if title == "" || content == "" || slug == "" {
		return fmt.Errorf("Title, content and slug are required")
	}

	if len(title) > 40 {
		return fmt.Errorf("Title must be less than 40 characters")
	} else if len(title) < 5 {
		return fmt.Errorf("Title must be more than 5 characters")
	}

	if len(slug) > 50 {
		return fmt.Errorf("Slug must be less than 50 characters")
	} else if len(slug) < 5 {
		return fmt.Errorf("Slug must be more than 5 characters")
	}

	if len(content) > 10000 {
		return fmt.Errorf("Content must be less than 10000 characters")
	}
	return nil
}

func (h *PostHandler) GetPosts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, err := r.Cookie(h.auth.CookieName)

	authenticated := false

	if err == nil {
		authenticated, err = h.auth.ValidateToken(cookie.Value)
		if err != nil {
			h.logger.Println(err)
		}
	}

	posts, err := h.queries.GetPosts(ctx)
	if err != nil {
		h.logger.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	page := templates.MainPage(h.blogName, h.pagetitle, posts, authenticated)
	page.Render(ctx, w)
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(h.auth.CookieName)

	authenticated := false

	if err == nil {
		authenticated, err = h.auth.ValidateToken(cookie.Value)
		if err != nil {
			h.logger.Println(err)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
	}

	if !authenticated {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	ctx := r.Context()

	err = r.ParseForm()
	if err != nil {
		h.logger.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	slug := r.FormValue("slug")

	if err := validatePost(title, content, slug); err != nil {
		h.logger.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var parsedContent bytes.Buffer
	if err := h.md.Convert([]byte(content), &parsedContent); err != nil {
		h.logger.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	post := repository.CreatePostParams{
		Title:         title,
		Content:       content,
		ParsedContent: parsedContent.String(),
		Slug:          slug,
		CreatedAt:     pgtype.Timestamp{Time: time.Now().In(h.location), Valid: true},
		ModifiedAt:    pgtype.Timestamp{Time: time.Now().In(h.location), Valid: true},
	}

	createdPost, err := h.queries.CreatePost(ctx, post)
	if err != nil {
		h.logger.Println(err)
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

	cookie, err := r.Cookie(h.auth.CookieName)

	authenticated := false

	if err == nil {
		authenticated, err = h.auth.ValidateToken(cookie.Value)
		if err != nil {
			h.logger.Println(err)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
	}

	if !authenticated {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if slug != "" {
		post, err := h.queries.GetPostBySlug(ctx, slug)
		if err != nil {
			h.logger.Println(err)
			http.Error(w, fmt.Sprintf("Post not found: %s", err.Error()), http.StatusNotFound)
			return
		}

		page := templates.Editor(h.blogName, h.pagetitle, post.Content, post.Title, post.Slug, true, authenticated)
		page.Render(ctx, w)
		return
	}

	editor := templates.Editor(h.blogName, h.pagetitle, "", "", "", false, authenticated)
	editor.Render(ctx, w)
}

func (h *PostHandler) ParseMarkdown(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(h.auth.CookieName)

	authenticated := false

	if err == nil {
		authenticated, err = h.auth.ValidateToken(cookie.Value)
		if err != nil {
			h.logger.Println(err)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
	}

	if !authenticated {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	ctx := r.Context()

	err = r.ParseForm()
	if err != nil {
		h.logger.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	slug := r.FormValue("slug")

	var buf bytes.Buffer
	if err := h.md.Convert([]byte(content), &buf); err != nil {
		h.logger.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	markdown := templates.Markdown(buf.String(), title, slug, time.Now(), time.Now())
	markdown.Render(ctx, w)
}

func (h *PostHandler) ViewPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	slug := r.PathValue("slug")

	post, err := h.queries.GetPostBySlug(ctx, slug)
	if err != nil {
		h.logger.Println(err)
	}

	cookie, err := r.Cookie(h.auth.CookieName)

	authenticated := false

	if err == nil {
		authenticated, err = h.auth.ValidateToken(cookie.Value)
		if err != nil {
			h.logger.Println(err)
		}
	}

	page := templates.PostPage(h.blogName, h.pagetitle, post, authenticated)
	page.Render(ctx, w)
}

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(h.auth.CookieName)

	authenticated := false

	if err == nil {
		authenticated, err = h.auth.ValidateToken(cookie.Value)
		if err != nil {
			h.logger.Println(err)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
	}

	if !authenticated {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	ctx := r.Context()

	slug := r.PathValue("slug")

	err = h.queries.DeletePostBySlug(ctx, slug)
	if err != nil {
		h.logger.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Location", "/")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("Post deleted"))
}

func (h *PostHandler) EditPost(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(h.auth.CookieName)

	authenticated := false

	if err == nil {
		authenticated, err = h.auth.ValidateToken(cookie.Value)
		if err != nil {
			h.logger.Println(err)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
	}

	if !authenticated {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	ctx := r.Context()

	slug := r.PathValue("slug")

	newTitle := r.FormValue("title")
	newSlug := r.FormValue("slug")
	newContent := r.FormValue("content")

	if err := validatePost(newTitle, newContent, newSlug); err != nil {
		h.logger.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var parsedContent bytes.Buffer
	if err := h.md.Convert([]byte(newContent), &parsedContent); err != nil {
		h.logger.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	post := repository.UpdatePostBySlugParams{
		Title:         newTitle,
		Slug:          newSlug,
		Slug_2:        slug,
		Content:       newContent,
		ParsedContent: parsedContent.String(),
		ModifiedAt:    pgtype.Timestamp{Time: time.Now().In(h.location), Valid: true},
	}

	err = h.queries.UpdatePostBySlug(ctx, post)
	if err != nil {
		h.logger.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Location", "/")
}
