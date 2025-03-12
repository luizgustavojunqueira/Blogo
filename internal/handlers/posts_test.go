package handlers

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/luizgustavojunqueira/Blogo/internal/repository"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

const charset = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ" + "0123456789"

func randomString(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	// Mapeia os bytes para os caracteres do charset
	for i, v := range b {
		b[i] = charset[int(v)%len(charset)]
	}
	return string(b), nil
}

func Test_validatePost(t *testing.T) {
	longContent, err := randomString(10001)
	if err != nil {
		t.Errorf("Error generating random string: %v", err)
	}

	type args struct {
		title   string
		content string
		slug    string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Empty title",
			args: args{
				title:   "",
				content: "Test content",
				slug:    "test-title",
			},
			wantErr: true,
		},
		{
			name: "Empty content",
			args: args{
				title:   "Test title",
				content: "",
				slug:    "test-title",
			},
			wantErr: true,
		},
		{
			name: "Empty slug",
			args: args{
				title:   "Test title",
				content: "Test content",
				slug:    "",
			},
			wantErr: true,
		},
		{
			name: "All empty",
			args: args{
				title:   "",
				content: "",
				slug:    "",
			},
			wantErr: true,
		},
		{
			name: "All filled",
			args: args{
				title:   "Test title",
				content: "Test content",
				slug:    "test-title",
			},
			wantErr: false,
		},
		{
			name: "Title too long",
			args: args{
				title:   "This title has more than 40 characters. And that's too long for this blog",
				slug:    "test-title",
				content: "Test content",
			},
			wantErr: true,
		},
		{
			name: "Title too short",
			args: args{
				title:   "Test",
				content: "Test content",
				slug:    "test-title",
			},
			wantErr: true,
		},
		{
			name: "Slug too long",
			args: args{
				title:   "Test title",
				content: "Test content",
				slug:    "this-slug-is-too-long-for-this-blog-and-should-not-be-accepted",
			},
			wantErr: true,
		},
		{
			name: "Slug too short",
			args: args{
				title:   "Test title",
				content: "Test content",
				slug:    "test",
			},
			wantErr: true,
		},
		{
			name: "Content too long",
			args: args{
				title:   "Test title",
				content: longContent,
				slug:    "test-title",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validatePost(tt.args.title, tt.args.content, tt.args.slug); (err != nil) != tt.wantErr {
				t.Errorf("validatePost() error = %v, wantErr %v,testName: %v", err, tt.wantErr, tt.name)
			}
		})
	}
}

func Test_getPostToc(t *testing.T) {
	md := goldmark.New(goldmark.WithExtensions(),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
			html.WithHardWraps(),
		))

	type args struct {
		md  goldmark.Markdown
		src string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Test case 1",
			args: args{
				md: md,
				src: `
# Teste
				`,
			},
			want: `
<ul>
<li>
<a href="#teste">Teste</a></li>
</ul>
			`,
			wantErr: false,
		},
		{
			name: "Test case 2",
			args: args{
				md: md,
				src: `
# Teste 
## Teste 2
				`,
			},
			want: `
<ul>
<li>
<a href="#teste">Teste</a><ul>
<li>
<a href="#teste-2">Teste 2</a></li>
</ul>
</li>
</ul>
			`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getPostToc(tt.args.md, []byte(tt.args.src))
			if (err != nil) != tt.wantErr {
				t.Errorf("getPostToc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotNormalized := strings.Join(strings.Fields(got), " ")
			wantNormalized := strings.Join(strings.Fields(tt.want), " ")
			if gotNormalized != wantNormalized {
				t.Errorf("getPostToc() =\n%v, want %v", got, tt.want)
			}
		})
	}
}

// FakeAuth implementa os métodos necessários para teste
type fakeAuth struct {
	cookieName string
	// se validToken for true, o token "valid-token" será considerado válido
	validToken bool
}

func (fa *fakeAuth) GetCookieName() string {
	return fa.cookieName
}

func (fa *fakeAuth) ValidateToken(token string) (bool, error) {
	if token == "valid-token" && fa.validToken {
		return true, nil
	}
	return false, fmt.Errorf("Invalid token")
}

// FakeQueries implementa apenas o método GetPosts para teste
type fakeQueries struct{}

func (fq *fakeQueries) GetPosts(ctx context.Context) ([]repository.Post, error) {
	// Retorne um slice com um post dummy
	return []repository.Post{
		{
			Title: "Post de Teste",
			Slug:  "post-de-teste",
			CreatedAt: pgtype.Timestamp{
				Time:  time.Date(2024, 1, 2, 15, 4, 0, 0, time.UTC),
				Valid: true,
			},
			ModifiedAt: pgtype.Timestamp{
				Time:  time.Date(2024, 1, 2, 15, 4, 0, 0, time.UTC),
				Valid: true,
			},
			Toc:           "<ul><li><a href=\"#post-de-teste\">Post de Teste</a></li></ul>",
			ParsedContent: "<p>Conteúdo do post de teste</p>",
		},
	}, nil
}

func (fq *fakeQueries) UpdatePostBySlug(ctx context.Context, arg repository.UpdatePostBySlugParams) error {
	return nil
}

func (fq *fakeQueries) GetPostBySlug(ctx context.Context, slug string) (repository.Post, error) {
	return repository.Post{
		Title: "Post de Teste",
		Slug:  "post-de-teste",
		CreatedAt: pgtype.Timestamp{
			Time:  time.Date(2024, 1, 2, 15, 4, 0, 0, time.UTC),
			Valid: true,
		},
		ModifiedAt: pgtype.Timestamp{
			Time:  time.Date(2024, 1, 2, 15, 4, 0, 0, time.UTC),
			Valid: true,
		},
		Toc:           "<ul><li><a href=\"#post-de-teste\">Post de Teste</a></li></ul>",
		ParsedContent: "<p>Conteúdo do post de teste</p>",
	}, nil
}

func (fq *fakeQueries) CreatePost(ctx context.Context, arg repository.CreatePostParams) (repository.Post, error) {
	return repository.Post{}, nil
}

func (fq *fakeQueries) DeletePostBySlug(ctx context.Context, slug string) error {
	return nil
}

func TestPostHandler_GetPostsUnauthorized(t *testing.T) {
	fakeAuthInstance := &fakeAuth{
		cookieName: "session",
		validToken: false,
	}
	fakeQueriesInstance := &fakeQueries{}

	logger := log.New(io.Discard, "", 0)
	location := time.UTC

	postHandler := NewPostHandler(
		fakeQueriesInstance,
		location,
		logger,
		fakeAuthInstance,
		"Blog de Teste",
		"Página de Teste",
	)

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  fakeAuthInstance.GetCookieName(),
		Value: "valid-token",
	})

	rr := httptest.NewRecorder()

	postHandler.GetPosts(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	respBody := rr.Body.String()
	if !strings.Contains(respBody, "Post de Teste") {
		t.Errorf("Response body does not contain expected post title. Got: %s", respBody)
	}

	if strings.Contains(respBody, "New Post") {
		t.Errorf("Response body contains editor link. Got: %s", respBody)
	}

	if !strings.Contains(respBody, "Login") {
		t.Errorf("Response body does not contain Login link. Got: %s", respBody)
	}

	if strings.Contains(respBody, "delete") {
		t.Errorf("Response body contains delete link. Got: %s", respBody)
	}

	if strings.Contains(respBody, "editor") {
		t.Errorf("Response body contains edit link. Got: %s", respBody)
	}
}

func TestPostHandler_GetPostsAuthorized(t *testing.T) {
	fakeAuthInstance := &fakeAuth{
		cookieName: "session",
		validToken: true,
	}
	fakeQueriesInstance := &fakeQueries{}

	logger := log.New(io.Discard, "", 0)
	location := time.UTC

	postHandler := NewPostHandler(
		fakeQueriesInstance,
		location,
		logger,
		fakeAuthInstance,
		"Blog de Teste",
		"Página de Teste",
	)

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  fakeAuthInstance.GetCookieName(),
		Value: "valid-token",
	})

	rr := httptest.NewRecorder()

	postHandler.GetPosts(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	respBody := rr.Body.String()
	if !strings.Contains(respBody, "Post de Teste") {
		t.Errorf("Response body does not contain expected post title. Got: %s", respBody)
	}

	if !strings.Contains(respBody, "New Post") {
		t.Errorf("Response body contains editor link. Got: %s", respBody)
	}

	if !strings.Contains(respBody, "Logout") {
		t.Errorf("Response body does not contain Login link. Got: %s", respBody)
	}

	if !strings.Contains(respBody, "delete") {
		t.Errorf("Response body contains delete link. Got: %s", respBody)
	}

	if !strings.Contains(respBody, "editor") {
		t.Errorf("Response body contains edit link. Got: %s", respBody)
	}
}

func TestPostHandler_ViewPostUnauthorized(t *testing.T) {
	fakeAuthInstance := &fakeAuth{
		cookieName: "session",
		validToken: false,
	}
	fakeQueriesInstance := &fakeQueries{}

	logger := log.New(io.Discard, "", 0)
	location := time.UTC

	postHandler := NewPostHandler(
		fakeQueriesInstance,
		location,
		logger,
		fakeAuthInstance,
		"Blog de Teste",
		"Página de Teste",
	)

	req := httptest.NewRequest("GET", "/post/post-de-teste", nil)
	req.AddCookie(&http.Cookie{
		Name:  fakeAuthInstance.GetCookieName(),
		Value: "valid-token",
	})

	rr := httptest.NewRecorder()

	postHandler.ViewPost(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	respBody := rr.Body.String()

	if strings.Contains(respBody, "editor") {
		t.Errorf("Response body contains edit link. Got: %s", respBody)
	}

	if strings.Contains(respBody, "delete") {
		t.Errorf("Response body contains delete link. Got: %s", respBody)
	}

	if strings.Contains(respBody, "New Post") {
		t.Errorf("Response body contains editor link. Got: %s", respBody)
	}

	if !strings.Contains(respBody, "Login") {
		t.Errorf("Response body does not contain Login link. Got: %s", respBody)
	}

	if !strings.Contains(respBody, "Post de Teste") {
		t.Errorf("Response body does not contain expected post title. Got: %s", respBody)
	}

	if !strings.Contains(respBody, "Conteúdo do post de teste") {
		t.Errorf("Response body does not contain expected post content. Got: %s", respBody)
	}

	toc := "<ul><li><a href=\"#post-de-teste\">Post de Teste</a></li></ul>"

	if !strings.Contains(respBody, toc) {
		t.Errorf("Response body does not contain expected toc. Got: %s", respBody)
	}
}

func TestPostHandler_ViewPostAuthorized(t *testing.T) {
	fakeAuthInstance := &fakeAuth{
		cookieName: "session",
		validToken: true,
	}
	fakeQueriesInstance := &fakeQueries{}

	logger := log.New(io.Discard, "", 0)
	location := time.UTC

	postHandler := NewPostHandler(
		fakeQueriesInstance,
		location,
		logger,
		fakeAuthInstance,
		"Blog de Teste",
		"Página de Teste",
	)

	req := httptest.NewRequest("GET", "/post/post-de-teste", nil)
	req.AddCookie(&http.Cookie{
		Name:  fakeAuthInstance.GetCookieName(),
		Value: "valid-token",
	})

	rr := httptest.NewRecorder()

	postHandler.ViewPost(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	respBody := rr.Body.String()

	if !strings.Contains(respBody, "Edit") {
		t.Errorf("Response body does not contain edit link. Got: %s", respBody)
	}

	if !strings.Contains(respBody, "Logout") {
		t.Errorf("Response body does not contain Login link. Got: %s", respBody)
	}

	if !strings.Contains(respBody, "Post de Teste") {
		t.Errorf("Response body does not contain expected post title. Got: %s", respBody)
	}

	if !strings.Contains(respBody, "Conteúdo do post de teste") {
		t.Errorf("Response body does not contain expected post content. Got: %s", respBody)
	}

	toc := "<ul><li><a href=\"#post-de-teste\">Post de Teste</a></li></ul>"

	if !strings.Contains(respBody, toc) {
		t.Errorf("Response body does not contain expected toc. Got: %s", respBody)
	}
}
