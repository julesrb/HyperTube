package comments

import (
	"context"
	"errors"
	"hypertube/api/internal/models"
	"hypertube/api/internal/respond"
	"log"
	"net/http"
)

type CommentsHandler struct {
	store CommentStore
}

func NewCommentsHandler(store CommentStore) *CommentsHandler {
	return &CommentsHandler{store: store}
}

type CommentStore interface {
	findByID(ctx context.Context, id string) (*models.Comment, error)
	findAll(ctx context.Context) ([]models.Comment, error)
}

func (h *CommentsHandler) List(w http.ResponseWriter, r *http.Request) {
	comments, err := h.store.findAll(r.Context())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			respond.Error(w, http.StatusNotFound, "NOT_FOUND", "comments not found")
		} else {
			log.Println("db err:", err)
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load comments")
		}
		return
	}
	respond.List(w, http.StatusOK, comments, len(comments))
}

func (h *CommentsHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	comment, err := h.store.findByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			respond.Error(w, http.StatusNotFound, "NOT_FOUND", "comment not found")
		} else {
			log.Println("db err:", err)
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load comment")
		}
		return
	}
	respond.Item(w, http.StatusOK, comment)
}
func (h *CommentsHandler) Create(w http.ResponseWriter, r *http.Request) {}
func (h *CommentsHandler) Update(w http.ResponseWriter, r *http.Request) {}
func (h *CommentsHandler) Delete(w http.ResponseWriter, r *http.Request) {}
