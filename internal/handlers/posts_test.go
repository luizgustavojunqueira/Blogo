package handlers

import (
	"crypto/rand"
	"fmt"
	"strings"
	"testing"

	"github.com/yuin/goldmark/renderer/html"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
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
			fmt.Println(tt.args.src)
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
