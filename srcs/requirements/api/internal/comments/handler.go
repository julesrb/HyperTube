package comments

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"hypertube/api/internal/auth"
	"hypertube/api/internal/i18n"
	"hypertube/api/internal/models"
	"hypertube/api/internal/respond"
)

type CommentsHandler struct {
	store CommentStore
}

func NewCommentsHandler(store CommentStore) *CommentsHandler {
	return &CommentsHandler{store: store}
}

type CommentStore interface {
	create(ctx context.Context, content string, movieID string, userID int) (models.Comment, error)
	findByID(ctx context.Context, id int) (*models.Comment, error)
	findAll(ctx context.Context, limit, offset int) ([]models.Comment, error)
	findAllByUserID(ctx context.Context, userID int64, limit, offset int) ([]models.CommentWithMovie, error)
	countAll(ctx context.Context) (int, error)
	countAllByUserID(ctx context.Context, userID int64) (int, error)
	update(ctx context.Context, content string, id int, userID int) (models.Comment, error)
	delete(ctx context.Context, id int, userID int) error
}

const commentPageLimit = 12
const maxJSONBodyBytes = 1 << 20

func parsePositiveID(raw string) (int, bool) {
	id, err := strconv.Atoi(raw)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func (h *CommentsHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respond.LocalizedError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", i18n.MsgMissingUserContext)
		return
	}

	rawMovieID := r.PathValue("movie_id")
	if rawMovieID == "" {
		rawMovieID = r.PathValue("id")
	}
	movieID := strings.TrimSpace(rawMovieID)
	if movieID == "" {
		respond.LocalizedError(w, r, http.StatusNotFound, "NOT_FOUND", i18n.MsgMovieNotFound)
		return
	}

	content, ok := decodeCommentContent(w, r)
	if !ok {
		return
	}

	comment, err := h.store.create(r.Context(), content, movieID, int(userID))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			respond.LocalizedError(w, r, http.StatusNotFound, "NOT_FOUND", i18n.MsgMovieNotFound)
			return
		}
		log.Println("db err:", err)
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedCreateComment)
		return
	}

	respond.Item(w, http.StatusCreated, comment)
}

func (h *CommentsHandler) List(w http.ResponseWriter, r *http.Request) {
	page := parsePage(r)
	total, err := h.store.countAll(r.Context())
	if err != nil {
		log.Println("db err:", err)
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedLoadComments)
		return
	}

	comments, err := h.store.findAll(r.Context(), commentPageLimit, page*commentPageLimit)
	if err != nil {
		log.Println("db err:", err)
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedLoadComments)
		return
	}

	respond.ListPaginated(w, http.StatusOK, comments, total, page, commentPageLimit)
}

func (h *CommentsHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respond.LocalizedError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", i18n.MsgMissingUserContext)
		return
	}

	id, ok := parsePositiveID(r.PathValue("id"))
	if !ok {
		respond.LocalizedError(w, r, http.StatusNotFound, "NOT_FOUND", i18n.MsgUserNotFound)
		return
	}

	page := parsePage(r)
	total, err := h.store.countAllByUserID(r.Context(), int64(id))
	if err != nil {
		log.Println("db err:", err)
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedLoadComments)
		return
	}

	comments, err := h.store.findAllByUserID(
		r.Context(),
		int64(id),
		commentPageLimit,
		page*commentPageLimit,
	)
	if err != nil {
		log.Println("db err:", err)
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedLoadComments)
		return
	}

	if comments == nil {
		comments = []models.CommentWithMovie{}
	}

	respond.ListPaginated(w, http.StatusOK, comments, total, page, commentPageLimit)
}

func (h *CommentsHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePositiveID(r.PathValue("id"))
	if !ok {
		respond.LocalizedError(w, r, http.StatusNotFound, "NOT_FOUND", i18n.MsgCommentNotFound)
		return
	}

	comment, err := h.store.findByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			respond.LocalizedError(w, r, http.StatusNotFound, "NOT_FOUND", i18n.MsgCommentNotFound)
		} else {
			log.Println("db err:", err)
			respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedLoadComment)
		}
		return
	}

	respond.Item(w, http.StatusOK, comment)
}

func (h *CommentsHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respond.LocalizedError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", i18n.MsgMissingUserContext)
		return
	}

	id, ok := parsePositiveID(r.PathValue("id"))
	if !ok {
		respond.LocalizedError(w, r, http.StatusNotFound, "NOT_FOUND", i18n.MsgCommentNotFound)
		return
	}

	comment, err := h.store.findByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			respond.LocalizedError(w, r, http.StatusNotFound, "NOT_FOUND", i18n.MsgCommentNotFound)
			return
		}
		log.Println("db err:", err)
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedLoadComment)
		return
	}
	if comment.UserID != int(userID) {
		respond.LocalizedError(w, r, http.StatusNotFound, "NOT_FOUND", i18n.MsgCommentNotFound)
		return
	}

	content, ok := decodeCommentContent(w, r)
	if !ok {
		return
	}

	updatedComment, err := h.store.update(r.Context(), content, id, int(userID))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			respond.LocalizedError(w, r, http.StatusNotFound, "NOT_FOUND", i18n.MsgCommentNotFound)
			return
		}
		log.Println("db err:", err)
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedUpdateComment)
		return
	}

	respond.Item(w, http.StatusOK, updatedComment)
}

func (h *CommentsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		respond.LocalizedError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", i18n.MsgMissingUserContext)
		return
	}

	id, ok := parsePositiveID(r.PathValue("id"))
	if !ok {
		respond.LocalizedError(w, r, http.StatusNotFound, "NOT_FOUND", i18n.MsgCommentNotFound)
		return
	}

	err := h.store.delete(r.Context(), id, int(userID))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			respond.LocalizedError(w, r, http.StatusNotFound, "NOT_FOUND", i18n.MsgCommentNotFound)
			return
		}
		log.Println("db err:", err)
		respond.LocalizedError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", i18n.MsgFailedDeleteComment)
		return
	}

	respond.Item(w, http.StatusOK, nil)
}

func parsePage(r *http.Request) int {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 0 {
		return 0
	}
	return page
}

func decodeCommentContent(w http.ResponseWriter, r *http.Request) (string, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBodyBytes)

	decoder := json.NewDecoder(r.Body)
	var body map[string]json.RawMessage
	if err := decoder.Decode(&body); err != nil || body == nil {
		log.Println("decode err:", err)
		respond.LocalizedFieldValidationError(w, r, http.StatusBadRequest, "body", i18n.MsgInvalidRequestBody)
		return "", false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		respond.LocalizedFieldValidationError(w, r, http.StatusBadRequest, "body", i18n.MsgInvalidRequestBody)
		return "", false
	}

	raw, ok := body["content"]
	if !ok {
		respond.LocalizedFieldValidationError(w, r, http.StatusBadRequest, "content", i18n.MsgInvalidRequestBody)
		return "", false
	}
	if strings.TrimSpace(string(raw)) == "null" {
		respond.LocalizedFieldValidationError(w, r, http.StatusBadRequest, "content", i18n.MsgInvalidRequestBody)
		return "", false
	}

	var content string
	if err := json.Unmarshal(raw, &content); err != nil {
		respond.LocalizedFieldValidationError(w, r, http.StatusBadRequest, "content", i18n.MsgInvalidRequestBody)
		return "", false
	}

	content = strings.TrimSpace(content)
	if content == "" {
		respond.LocalizedFieldValidationError(w, r, http.StatusBadRequest, "content", i18n.MsgInvalidRequestBody)
		return "", false
	}
	return content, true
}
