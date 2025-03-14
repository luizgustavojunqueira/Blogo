package handlers

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"slices"
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

type authMock struct {
	cookieName string
	validToken bool
}

func (fa *authMock) GetCookieName() string {
	return fa.cookieName
}

func (fa *authMock) ValidateToken(token string) (bool, error) {
	if fa.validToken {
		return true, nil
	}
	return false, fmt.Errorf("Invalid token")
}

type databaseMock struct {
	posts []repository.Post
}

type queriesMock struct {
	dbMock *databaseMock
}

func (fq *queriesMock) GetPosts(ctx context.Context) ([]repository.Post, error) {
	return fq.dbMock.posts, nil
}

func (fq *queriesMock) UpdatePostBySlug(ctx context.Context, arg repository.UpdatePostBySlugParams) error {
	for i, post := range fq.dbMock.posts {
		if post.Slug == arg.Slug {
			fq.dbMock.posts[i].Title = arg.Title
			fq.dbMock.posts[i].Content = arg.Content
			fq.dbMock.posts[i].ModifiedAt = arg.ModifiedAt
			fq.dbMock.posts[i].ParsedContent = arg.ParsedContent
			fq.dbMock.posts[i].Toc = arg.Toc
			return nil
		}
	}
	return fmt.Errorf("Post not found")
}

func (fq *queriesMock) GetPostBySlug(ctx context.Context, slug string) (repository.Post, error) {
	for i, post := range fq.dbMock.posts {
		if post.Slug == slug {
			return fq.dbMock.posts[i], nil
		}
	}
	return repository.Post{}, fmt.Errorf("Post not found")
}

func (fq *queriesMock) CreatePost(ctx context.Context, arg repository.CreatePostParams) (repository.Post, error) {
	newPost := repository.Post{
		Title:         arg.Title,
		Content:       arg.Content,
		Slug:          arg.Slug,
		CreatedAt:     arg.CreatedAt,
		ModifiedAt:    arg.ModifiedAt,
		ParsedContent: arg.ParsedContent,
		Toc:           arg.Toc,
	}
	fq.dbMock.posts = append(fq.dbMock.posts, newPost)
	return newPost, nil
}

func (fq *queriesMock) DeletePostBySlug(ctx context.Context, slug string) error {
	for i, post := range fq.dbMock.posts {
		if post.Slug == slug {
			fq.dbMock.posts = slices.Delete(fq.dbMock.posts, i, i+1)
			return nil
		}
	}
	return fmt.Errorf("Post not found")
}

func TestPostHandler_GetPosts(t *testing.T) {
	fakeQueriesInstance := &queriesMock{
		dbMock: &databaseMock{
			posts: []repository.Post{
				{
					Title:         "Post de Teste",
					Content:       "Conteúdo do post de teste",
					Slug:          "post-de-teste",
					CreatedAt:     pgtype.Timestamp{Time: time.Now(), Valid: true},
					ModifiedAt:    pgtype.Timestamp{Time: time.Now(), Valid: true},
					ParsedContent: "<p>Conteúdo do post de teste</p>",
					Toc:           "<ul><li><a href=\"#post-de-teste\">Post de Teste</a></li></ul>",
					Description:   pgtype.Text{String: "Descrição do post de teste", Valid: true},
				},
			},
		},
	}

	logger := log.New(io.Discard, "", 0)
	location := time.UTC

	type args struct {
		fakeQueriesInstance *queriesMock
		fakeAuthInstance    *authMock
		location            *time.Location
		logger              *log.Logger
		title               string
		pageTitle           string
	}

	test := []struct {
		name                 string
		wantCode             int
		wantBodyContains     []string
		dontWantBodyContains []string
		args                 args
	}{
		{
			name:     "Test Unauthorized",
			wantCode: http.StatusOK,
			wantBodyContains: []string{
				"Página de Teste",
				"Post de Teste",
				"Login",
				"Descrição do post de teste",
			},
			dontWantBodyContains: []string{
				"New Post",
				"delete",
				"edit",
			},
			args: args{
				fakeQueriesInstance: fakeQueriesInstance,
				fakeAuthInstance:    &authMock{cookieName: "session", validToken: false},
				location:            location,
				logger:              logger,
				title:               "Blog de Teste",
				pageTitle:           "Página de Teste",
			},
		},
		{
			name:     "Test Authorized 2",
			wantCode: http.StatusOK,
			wantBodyContains: []string{
				"Página de Teste",
				"Post de Teste",
				"Descrição do post de teste",
				"New Post",
				"delete",
				"edit",
				"Logout",
			},
			dontWantBodyContains: []string{
				"Login",
			},
			args: args{
				fakeQueriesInstance: fakeQueriesInstance,
				fakeAuthInstance:    &authMock{cookieName: "session", validToken: true},
				location:            location,
				logger:              logger,
				title:               "Blog de Teste",
				pageTitle:           "Página de Teste",
			},
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			postHandler := NewPostHandler(
				tt.args.fakeQueriesInstance,
				tt.args.location,
				tt.args.logger,
				tt.args.fakeAuthInstance,
				tt.args.pageTitle,
				tt.args.title,
			)

			req := httptest.NewRequest("GET", "/", nil)
			req.AddCookie(&http.Cookie{
				Name:  tt.args.fakeAuthInstance.GetCookieName(),
				Value: "token",
			})

			rr := httptest.NewRecorder()

			postHandler.GetPosts(rr, req)

			if rr.Code != tt.wantCode {
				t.Errorf("Expected status %d, got %d", tt.wantCode, rr.Code)
			}

			respBody := rr.Body.String()

			for _, want := range tt.wantBodyContains {
				if !strings.Contains(respBody, want) {
					t.Errorf("Expected body content: %v, got: %v", want, respBody)
				}
			}

			for _, dontWant := range tt.dontWantBodyContains {
				if strings.Contains(respBody, dontWant) {
					t.Errorf("Expected body content: %v, got: %v", dontWant, respBody)
				}
			}
		})
	}
}

// func TestPostHandler_ViewPostUnauthorized(t *testing.T) {
// 	fakeAuthInstance := &authMock{
// 		cookieName: "session",
// 		validToken: false,
// 	}
// 	fakeQueriesInstance := &queriesMock{
// 		dbMock: &databaseMock{
// 			posts: []repository.Post{
// 				{
// 					Title:         "Post de Teste",
// 					Content:       "Conteúdo do post de teste",
// 					Slug:          "post-de-teste",
// 					CreatedAt:     pgtype.Timestamp{Time: time.Now(), Valid: true},
// 					ModifiedAt:    pgtype.Timestamp{Time: time.Now(), Valid: true},
// 					ParsedContent: "<p>Conteúdo do post de teste</p>",
// 					Toc:           "<ul><li><a href=\"#post-de-teste\">Post de Teste</a></li></ul>",
// 				},
// 			},
// 		},
// 	}
//
// 	logger := log.New(io.Discard, "", 0)
// 	location := time.UTC
//
// 	postHandler := NewPostHandler(
// 		fakeQueriesInstance,
// 		location,
// 		logger,
// 		fakeAuthInstance,
// 		"Blog de Teste",
// 		"Página de Teste",
// 	)
//
// 	req := httptest.NewRequest("GET", "/post", nil)
// 	req.SetPathValue("slug", "post-de-teste")
// 	req.AddCookie(&http.Cookie{
// 		Name:  fakeAuthInstance.GetCookieName(),
// 		Value: "token",
// 	})
//
// 	rr := httptest.NewRecorder()
//
// 	postHandler.ViewPost(rr, req)
//
// 	if rr.Code != http.StatusOK {
// 		t.Errorf("Expected status 200, got %d", rr.Code)
// 	}
//
// 	respBody := rr.Body.String()
//
// 	if strings.Contains(respBody, "editor") {
// 		t.Errorf("Response body contains edit link. Got: %s", respBody)
// 	}
//
// 	if strings.Contains(respBody, "delete") {
// 		t.Errorf("Response body contains delete link. Got: %s", respBody)
// 	}
//
// 	if strings.Contains(respBody, "New Post") {
// 		t.Errorf("Response body contains editor link. Got: %s", respBody)
// 	}
//
// 	if !strings.Contains(respBody, "Login") {
// 		t.Errorf("Response body does not contain Login link. Got: %s", respBody)
// 	}
//
// 	if !strings.Contains(respBody, "Post de Teste") {
// 		t.Errorf("Response body does not contain expected post title. Got: %s", respBody)
// 	}
//
// 	if !strings.Contains(respBody, "Conteúdo do post de teste") {
// 		t.Errorf("Response body does not contain expected post content. Got: %s", respBody)
// 	}
//
// 	toc := "<ul><li><a href=\"#post-de-teste\">Post de Teste</a></li></ul>"
//
// 	if !strings.Contains(respBody, toc) {
// 		t.Errorf("Response body does not contain expected toc. Got: %s", respBody)
// 	}
// }
//
// func TestPostHandler_ViewPostAuthorized(t *testing.T) {
// 	fakeAuthInstance := &authMock{
// 		cookieName: "session",
// 		validToken: true,
// 	}
// 	fakeQueriesInstance := &queriesMock{
// 		dbMock: &databaseMock{
// 			posts: []repository.Post{
// 				{
// 					Title:         "Post de Teste",
// 					Content:       "Conteúdo do post de teste",
// 					Slug:          "post-de-teste",
// 					CreatedAt:     pgtype.Timestamp{Time: time.Now(), Valid: true},
// 					ModifiedAt:    pgtype.Timestamp{Time: time.Now(), Valid: true},
// 					ParsedContent: "<p>Conteúdo do post de teste</p>",
// 					Toc:           "<ul><li><a href=\"#post-de-teste\">Post de Teste</a></li></ul>",
// 				},
// 			},
// 		},
// 	}
//
// 	logger := log.New(io.Discard, "", 0)
// 	location := time.UTC
//
// 	postHandler := NewPostHandler(
// 		fakeQueriesInstance,
// 		location,
// 		logger,
// 		fakeAuthInstance,
// 		"Blog de Teste",
// 		"Página de Teste",
// 	)
//
// 	req := httptest.NewRequest("GET", "/post", nil)
// 	req.SetPathValue("slug", "post-de-teste")
// 	req.AddCookie(&http.Cookie{
// 		Name:  fakeAuthInstance.GetCookieName(),
// 		Value: "valid-token",
// 	})
//
// 	rr := httptest.NewRecorder()
//
// 	postHandler.ViewPost(rr, req)
//
// 	if rr.Code != http.StatusOK {
// 		t.Errorf("Expected status 200, got %d", rr.Code)
// 	}
//
// 	respBody := rr.Body.String()
//
// 	if !strings.Contains(respBody, "Edit") {
// 		t.Errorf("Response body does not contain edit link. Got: %s", respBody)
// 	}
//
// 	if !strings.Contains(respBody, "Logout") {
// 		t.Errorf("Response body does not contain Login link. Got: %s", respBody)
// 	}
//
// 	if !strings.Contains(respBody, "Post de Teste") {
// 		t.Errorf("Response body does not contain expected post title. Got: %s", respBody)
// 	}
//
// 	if !strings.Contains(respBody, "Conteúdo do post de teste") {
// 		t.Errorf("Response body does not contain expected post content. Got: %s", respBody)
// 	}
//
// 	toc := "<ul><li><a href=\"#post-de-teste\">Post de Teste</a></li></ul>"
//
// 	if !strings.Contains(respBody, toc) {
// 		t.Errorf("Response body does not contain expected toc. Got: %s", respBody)
// 	}
// }
