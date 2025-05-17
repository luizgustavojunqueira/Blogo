package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
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
	repository PostRepository
	tagsRepo   TagRepository
	md         goldmark.Markdown
	location   *time.Location
	logger     *log.Logger
	auth       Auth
	blogName   string
	pagetitle  string
}

type PostRepository interface {
	GetPosts(ctx context.Context) ([]repository.Post, error)
	CreatePost(ctx context.Context, arg repository.CreatePostParams) (repository.Post, error)
	GetPostBySlug(ctx context.Context, slug string) (repository.Post, error)
	DeletePostBySlug(ctx context.Context, slug string) error
	UpdatePostBySlug(ctx context.Context, arg repository.UpdatePostBySlugParams) (repository.Post, error)
	GetPostsByTag(ctx context.Context, tag string) ([]repository.Post, error)
}

type Auth interface {
	ValidateToken(token string) (bool, error)
	GetCookieName() string
}

func NewPostHandler(repo PostRepository, tagsRepo TagRepository, location *time.Location, logger *log.Logger, auth Auth, blogName, pagetitle string) *PostHandler {
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
		repository: repo,
		tagsRepo:   tagsRepo,
		md:         md,
		logger:     logger,
		location:   location,
		auth:       auth,
		blogName:   blogName,
		pagetitle:  pagetitle,
	}
}

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

func (h *PostHandler) GetPosts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, err := r.Cookie(h.auth.GetCookieName())

	authenticated := false

	if err == nil {
		authenticated, err = h.auth.ValidateToken(cookie.Value)
		if err != nil {
			h.logger.Println(err)
		}
	}

	posts, err := h.repository.GetPosts(ctx)
	if err != nil {
		h.logger.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	postsWithTags := make([]repository.PostWithTags, len(posts))

	for i, post := range posts {
		tags, err := h.tagsRepo.GetTagsByPost(ctx, post.Slug)
		if err != nil {
			h.logger.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		postsWithTags[i] = repository.PostWithTags{
			ID:            post.ID,
			Title:         post.Title,
			Slug:          post.Slug,
			CreatedAt:     post.CreatedAt,
			ModifiedAt:    post.ModifiedAt,
			ParsedContent: post.ParsedContent,
			Description:   post.Description,
			Toc:           post.Toc,
			Content:       post.Content,
			Tags:          tags,
		}
	}

	mainPage := pages.MainPage(h.blogName, h.pagetitle, postsWithTags, authenticated, "")

	root := pages.Root(h.blogName, mainPage)

	root.Render(ctx, w)
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(h.auth.GetCookieName())

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
	tags := r.FormValue("tags")

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

	createdPost, err := h.repository.CreatePost(ctx, post)
	if err != nil {
		h.logger.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	createdTags := make([]repository.Tag, 0)

	if tags != "" {
		tagNames := strings.Split(tags, ",")

		h.logger.Printf("Creating tags: %v\n", tagNames)

		for _, tagName := range tagNames {
			tagName = strings.TrimSpace(tagName)
			if tagName == "" {
				continue
			}

			h.logger.Printf("Creating tag: %s\n", tagName)

			tag, err := h.tagsRepo.CreateTagIfNotExists(ctx, repository.CreateTagIfNotExistsParams{
				Name:       tagName,
				CreatedAt:  pgtype.Timestamp{Time: time.Now().In(h.location), Valid: true},
				ModifiedAt: pgtype.Timestamp{Time: time.Now().In(h.location), Valid: true},
			})
			if err != nil {
				h.logger.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			h.logger.Printf("Adding tag %s to post %s\n", tagName, createdPost.Slug)

			err = h.tagsRepo.AddTagToPost(ctx, repository.AddTagToPostParams{
				PostID: createdPost.ID,
				TagID:  tag[0].ID,
			})
			if err != nil {
				h.logger.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			createdTag := repository.Tag{
				ID:         tag[0].ID,
				Name:       tag[0].Name,
				CreatedAt:  tag[0].CreatedAt,
				ModifiedAt: tag[0].ModifiedAt,
			}

			createdTags = append(createdTags, createdTag)
		}
	}

	w.Header().Set("HX-Location", "/")
	w.WriteHeader(http.StatusOK)

	createdPostWithTags := repository.PostWithTags{
		ID:            createdPost.ID,
		Title:         createdPost.Title,
		Slug:          createdPost.Slug,
		CreatedAt:     createdPost.CreatedAt,
		ModifiedAt:    createdPost.ModifiedAt,
		ParsedContent: createdPost.ParsedContent,
		Description:   createdPost.Description,
		Content:       createdPost.Content,
		Toc:           createdPost.Toc,
		Tags:          createdTags,
	}

	card := components.PostCard(createdPostWithTags, authenticated)
	card.Render(ctx, w)
}

func (h *PostHandler) Editor(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	slug := r.PathValue("slug")

	cookie, err := r.Cookie(h.auth.GetCookieName())

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
		post, err := h.repository.GetPostBySlug(ctx, slug)
		if err != nil {
			h.logger.Println(err)
			http.Error(w, fmt.Sprintf("Post not found: %s", err.Error()), http.StatusNotFound)
			return
		}

		tags, err := h.tagsRepo.GetTagsByPost(ctx, post.Slug)
		if err != nil {
			h.logger.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		postWithTags := repository.PostWithTags{
			ID:            post.ID,
			Title:         post.Title,
			Slug:          post.Slug,
			CreatedAt:     post.CreatedAt,
			ModifiedAt:    post.ModifiedAt,
			ParsedContent: post.ParsedContent,
			Description:   post.Description,
			Content:       post.Content,
			Toc:           post.Toc,
			Tags:          tags,
		}

		tagsJson := make([]string, len(tags))
		for i, tag := range tags {
			tagsJson[i] = tag.Name
		}

		tagsJsonBytes, _ := json.Marshal(tagsJson) // retorna []byte JSON válido
		tagsJsonString := string(tagsJsonBytes)    // converte para string JSON: ["tag1","tag2"]

		h.logger.Printf("Tags: %s\n", tagsJsonString)

		editorPage := pages.EditorPage(h.blogName, h.pagetitle, postWithTags, true, authenticated, tagsJsonString)

		page := pages.Root(h.blogName, editorPage)
		page.Render(ctx, w)
		return
	}

	editorPage := pages.EditorPage(h.blogName, h.pagetitle, repository.PostWithTags{}, false, authenticated, "")

	page := pages.Root(h.blogName, editorPage)
	page.Render(ctx, w)
}

func (h *PostHandler) ParseMarkdown(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(h.auth.GetCookieName())

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
	tags := r.FormValue("tags")

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

	postTags := make([]repository.Tag, 0)

	if tags != "" {
		tagNames := strings.Split(tags, ",")
		for _, tagName := range tagNames {
			tagName = strings.TrimSpace(tagName)
			if tagName == "" {
				continue
			}
			tag := repository.Tag{
				Name:       tagName,
				CreatedAt:  pgtype.Timestamp{Time: time.Now().In(h.location), Valid: true},
				ModifiedAt: pgtype.Timestamp{Time: time.Now().In(h.location), Valid: true},
			}
			postTags = append(postTags, tag)
		}
	}

	post := repository.PostWithTags{
		Title:         title,
		Content:       content,
		ParsedContent: buf.String(),
		Toc:           toc,
		Slug:          slug,
		Tags:          postTags,
	}

	markdown := components.Markdown(post)
	markdown.Render(ctx, w)
}

func (h *PostHandler) ViewPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	slug := r.PathValue("slug")

	post, err := h.repository.GetPostBySlug(ctx, slug)
	if err != nil {
		h.logger.Println(err)
		return
	}

	cookie, err := r.Cookie(h.auth.GetCookieName())

	authenticated := false

	if err == nil {
		authenticated, err = h.auth.ValidateToken(cookie.Value)
		if err != nil {
			h.logger.Println(err)
		}
	}

	postTags, err := h.tagsRepo.GetTagsByPost(ctx, post.Slug)
	if err != nil {
		h.logger.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	postWithTags := repository.PostWithTags{
		ID:            post.ID,
		Title:         post.Title,
		Slug:          post.Slug,
		CreatedAt:     post.CreatedAt,
		ModifiedAt:    post.ModifiedAt,
		ParsedContent: post.ParsedContent,
		Description:   post.Description,
		Content:       post.Content,
		Toc:           post.Toc,
		Tags:          postTags,
	}

	postPage := pages.PostPage(h.blogName, h.pagetitle, postWithTags, authenticated)

	page := pages.Root(h.blogName, postPage)
	page.Render(ctx, w)
}

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(h.auth.GetCookieName())

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

	err = h.repository.DeletePostBySlug(ctx, slug)
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
	cookie, err := r.Cookie(h.auth.GetCookieName())

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
	newTags := r.FormValue("tags")

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

	updatedPost, err := h.repository.UpdatePostBySlug(ctx, post)
	if err != nil {
		h.logger.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.tagsRepo.ClearPostTagsBySlug(ctx, slug)
	if err != nil {
		h.logger.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if newTags != "" {
		tagNames := strings.Split(newTags, ",")
		for _, tagName := range tagNames {
			tagName = strings.TrimSpace(tagName)
			if tagName == "" {
				continue
			}
			tag, err := h.tagsRepo.CreateTagIfNotExists(ctx, repository.CreateTagIfNotExistsParams{
				Name:       tagName,
				CreatedAt:  pgtype.Timestamp{Time: time.Now().In(h.location), Valid: true},
				ModifiedAt: pgtype.Timestamp{Time: time.Now().In(h.location), Valid: true},
			})
			if err != nil {
				h.logger.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			err = h.tagsRepo.AddTagToPost(ctx, repository.AddTagToPostParams{
				PostID: updatedPost.ID,
				TagID:  tag[0].ID,
			})
		}
	}

	w.Header().Set("HX-Location", "/")
}

func (h *PostHandler) GetPostsByTag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cookie, err := r.Cookie(h.auth.GetCookieName())
	authenticated := false
	if err == nil {
		authenticated, err = h.auth.ValidateToken(cookie.Value)
		if err != nil {
			h.logger.Println(err)
		}
	}

	tag := r.PathValue("tag")
	if tag == "" {
		http.Error(w, "Tag parameter is required", http.StatusBadRequest)
		return
	}

	tags, err := h.tagsRepo.SearchTags(ctx, pgtype.Text{String: tag, Valid: true})
	if err != nil {
		h.logger.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	postsWithTags := make([]repository.PostWithTags, 0)

	for _, tag := range tags {
		posts, err := h.repository.GetPostsByTag(ctx, tag)
		if err != nil {
			h.logger.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, post := range posts {
			tags, err := h.tagsRepo.GetTagsByPost(ctx, post.Slug)
			if err != nil {
				h.logger.Println(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			postsWithTags = append(postsWithTags, repository.PostWithTags{
				ID:            post.ID,
				Title:         post.Title,
				Slug:          post.Slug,
				CreatedAt:     post.CreatedAt,
				ModifiedAt:    post.ModifiedAt,
				ParsedContent: post.ParsedContent,
				Description:   post.Description,
				Toc:           post.Toc,
				Content:       post.Content,
				Tags:          tags,
			},
			)
		}
	}
	mainPage := pages.MainPage(h.blogName, h.pagetitle, postsWithTags, authenticated, tag)
	root := pages.Root(h.blogName, mainPage)
	root.Render(ctx, w)
}
