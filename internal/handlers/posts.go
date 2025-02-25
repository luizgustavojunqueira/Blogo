package handlers

import (
	"Blog/internal/repository"
	"Blog/internal/templates"
	"net/http"
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
