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
	"github.com/luizgustavojunqueira/Blogo/internal/templates/components"
	"github.com/luizgustavojunqueira/Blogo/internal/templates/pages"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/toc"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
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
	md := goldmark.New(goldmark.WithExtensions(extension.GFM, extension.Table, extension.Typographer, highlighting.NewHighlighting(
		highlighting.WithStyle("dracula"),
		highlighting.WithFormatOptions(
			chromahtml.WithLineNumbers(true),
		),
	)),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithAttribute(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
			html.WithHardWraps(),
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

	mainPage := pages.MainPage(h.blogName, h.pagetitle, posts, authenticated)

	root := pages.Root(h.blogName, mainPage)

	root.Render(ctx, w)
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
	description := r.FormValue("description")

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

	toc, err := getPostToc(h.md, []byte(content))
	if err != nil {
		h.logger.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	post := repository.CreatePostParams{
		Title:         title,
		Toc:           toc,
		Content:       content,
		ParsedContent: parsedContent.String(),
		Description:   pgtype.Text{String: description, Valid: true},
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

	card := components.PostCard(createdPost, authenticated)
	card.Render(ctx, w)
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

		editorPage := pages.EditorPage(h.blogName, h.pagetitle, post, true, authenticated)

		page := pages.Root(h.blogName, editorPage)
		page.Render(ctx, w)
		return
	}

	editorPage := pages.EditorPage(h.blogName, h.pagetitle, repository.Post{}, false, authenticated)

	page := pages.Root(h.blogName, editorPage)
	page.Render(ctx, w)
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

	toc, err := getPostToc(h.md, []byte(content))
	if err != nil {
		h.logger.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	post := repository.Post{
		Title:         title,
		Content:       content,
		ParsedContent: buf.String(),
		Toc:           toc,
		Slug:          slug,
	}

	markdown := components.Markdown(post)
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

	postPage := pages.PostPage(h.blogName, h.pagetitle, post, authenticated)

	page := pages.Root(h.blogName, postPage)
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
	newDescription := r.FormValue("description")

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

	toc, err := getPostToc(h.md, []byte(newContent))
	if err != nil {
		h.logger.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	post := repository.UpdatePostBySlugParams{
		Title:         newTitle,
		Toc:           toc,
		Slug:          newSlug,
		Slug_2:        slug,
		Description:   pgtype.Text{String: newDescription, Valid: true},
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
