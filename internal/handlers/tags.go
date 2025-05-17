package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/luizgustavojunqueira/Blogo/internal/repository"
)

type TagRepository interface {
	GetTags(ctx context.Context) ([]repository.Tag, error)
	SearchTags(context.Context, pgtype.Text) ([]string, error)
	CreateTagIfNotExists(context.Context, repository.CreateTagIfNotExistsParams) ([]repository.CreateTagIfNotExistsRow, error)
	AddTagToPost(ctx context.Context, params repository.AddTagToPostParams) error
	GetTagsByPost(ctx context.Context, slug string) ([]repository.Tag, error)
	ClearPostTagsBySlug(ctx context.Context, slug string) error
}

type TagsHandler struct {
	repository TagRepository
	logger     *log.Logger
}

func NewTagsHandler(repo TagRepository, logger *log.Logger) *TagsHandler {
	return &TagsHandler{
		repository: repo,
		logger:     logger,
	}
}

func (h *TagsHandler) GetTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tags, err := h.repository.GetTags(ctx)
	if err != nil {
		h.logger.Println("Error getting tags:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tags); err != nil {
		h.logger.Println("Error encoding tags to JSON:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *TagsHandler) SearchTag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tag := r.PathValue("tag")
	if tag == "" {
		http.Error(w, "Tag parameter is required", http.StatusBadRequest)
		return
	}

	tags, err := h.repository.SearchTags(ctx, pgtype.Text{String: tag, Valid: true})
	if err != nil {
		h.logger.Println("Error searching for tag:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	h.logger.Println("Tags found:", tags)
	if err := json.NewEncoder(w).Encode(tags); err != nil {
		h.logger.Println("Error encoding tags to JSON:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
