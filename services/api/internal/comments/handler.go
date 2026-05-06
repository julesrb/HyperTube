package comments

import (
	"context"
	"encoding/json"
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
	update(ctx context.Context, content string, id string, user_id int) (models.Comment, error)
	delete(ctx context.Context, id string, user_id int) error
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

func (h *CommentsHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var input struct {
		Content string `json:"content"`
		UserID  int    `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println("decode err:", err)
		respond.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	// TODO implement authorization to ensure user can only update their own comments
	comment, err := h.store.update(r.Context(), input.Content, id, input.UserID)
	if err != nil {
		log.Println("db err:", err)
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update comment")
		return
	}
	respond.Item(w, http.StatusOK, comment)
}

func (h *CommentsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var input struct {
		UserID int `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println("decode err:", err)
		respond.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	// TODO implement authorization to ensure user can only update their own comments
	err := h.store.delete(r.Context(), id, input.UserID)
	if err != nil {
		log.Println("db err:", err)
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update comment")
		return
	}
	respond.Item(w, http.StatusOK, nil)
}
