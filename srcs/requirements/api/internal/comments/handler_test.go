package comments

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"hypertube/api/internal/auth"
	"hypertube/api/internal/models"
)

type fakeCommentStore struct {
	comment           *models.Comment
	comments          []models.Comment
	userComments      []models.CommentWithMovie
	total             int
	createMovieID     string
	createUserID      int
	createdContent    string
	createErr         error
	updateUserID      int
	updatedContent    string
	deleteUserID      int
	findErr           error
	updateErr         error
	deleteErr         error
	countByUserCalled bool
	listByUserCalled  bool
	countByUserID     int64
	listByUserID      int64
	requestedLimit    int
	requestedOffset   int
	countByUserErr    error
	listByUserErr     error
}

type existingAuthUserChecker struct{}

func (existingAuthUserChecker) UserExists(context.Context, int64) (bool, error) {
	return true, nil
}

func (s *fakeCommentStore) create(ctx context.Context, content string, movieID string, userID int) (models.Comment, error) {
	s.createMovieID = movieID
	s.createUserID = userID
	s.createdContent = content
	if s.createErr != nil {
		return models.Comment{}, s.createErr
	}
	return models.Comment{ID: 1, MovieID: movieID, UserID: userID, Content: content}, nil
}

func (s *fakeCommentStore) findByID(ctx context.Context, id int) (*models.Comment, error) {
	if s.findErr != nil {
		return nil, s.findErr
	}
	if s.comment != nil {
		return s.comment, nil
	}
	return &models.Comment{ID: id, UserID: 42, Content: "original"}, nil
}

func (s *fakeCommentStore) findAll(ctx context.Context, limit, offset int) ([]models.Comment, error) {
	return s.comments, nil
}

func (s *fakeCommentStore) countAll(ctx context.Context) (int, error) {
	if s.total != 0 {
		return s.total, nil
	}
	return len(s.comments), nil
}

func (s *fakeCommentStore) countAllByUserID(_ context.Context, userID int64) (int, error) {
	s.countByUserCalled = true
	s.countByUserID = userID
	if s.countByUserErr != nil {
		return 0, s.countByUserErr
	}
	return s.total, nil
}

func (s *fakeCommentStore) findAllByUserID(_ context.Context, userID int64, limit, offset int) ([]models.CommentWithMovie, error) {
	s.listByUserCalled = true
	s.listByUserID = userID
	s.requestedLimit = limit
	s.requestedOffset = offset
	if s.listByUserErr != nil {
		return nil, s.listByUserErr
	}
	return s.userComments, nil
}

func (s *fakeCommentStore) update(ctx context.Context, content string, id int, userID int) (models.Comment, error) {
	s.updateUserID = userID
	s.updatedContent = content
	if s.updateErr != nil {
		return models.Comment{}, s.updateErr
	}
	return models.Comment{ID: 1, UserID: userID, Content: content, Edited: true}, nil
}

func (s *fakeCommentStore) delete(ctx context.Context, id int, userID int) error {
	s.deleteUserID = userID
	return s.deleteErr
}

func TestCreateUsesAuthenticatedUserIDAndPathMovieID(t *testing.T) {
	store := &fakeCommentStore{}
	handler := NewCommentsHandler(store)

	req := httptest.NewRequest(http.MethodPost, "/movies/tt123/comments", strings.NewReader(`{"content":"  hello  ","user_id":999}`))
	req.SetPathValue("id", "tt123")
	rec := httptest.NewRecorder()

	serveWithUser(t, 42, http.HandlerFunc(handler.Create)).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if store.createMovieID != "tt123" {
		t.Fatalf("expected movie id tt123, got %q", store.createMovieID)
	}
	if store.createUserID != 42 {
		t.Fatalf("expected token user id 42, got %d", store.createUserID)
	}
	if store.createdContent != "hello" {
		t.Fatalf("expected trimmed content, got %q", store.createdContent)
	}

	body := decodeCommentItemEnvelope(t, rec)
	if body.Data.UserID != 42 {
		t.Fatalf("expected response user id 42, got %d", body.Data.UserID)
	}
}

func TestCreateRejectsMissingAuthenticatedUser(t *testing.T) {
	handler := NewCommentsHandler(&fakeCommentStore{})

	req := httptest.NewRequest(http.MethodPost, "/movies/tt123/comments", strings.NewReader(`{"content":"hello"}`))
	req.SetPathValue("id", "tt123")
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeCommentErrorEnvelope(t, rec).Error.Code; got != "UNAUTHORIZED" {
		t.Fatalf("expected UNAUTHORIZED, got %q", got)
	}
}

func TestCreateInvalidBodyReturnsFieldValidationError(t *testing.T) {
	handler := NewCommentsHandler(&fakeCommentStore{})

	req := httptest.NewRequest(http.MethodPost, "/movies/tt123/comments", strings.NewReader(`{"content":`))
	req.SetPathValue("id", "tt123")
	rec := httptest.NewRecorder()

	serveWithUser(t, 42, http.HandlerFunc(handler.Create)).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}

	errorBody := decodeCommentErrorEnvelope(t, rec).Error
	if errorBody.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR, got %q", errorBody.Code)
	}
	if got := errorBody.Fields["body"].Message; got != "Invalid request body" {
		t.Fatalf("expected body field validation, got %q", got)
	}
}

func TestListReturnsPaginatedComments(t *testing.T) {
	store := &fakeCommentStore{
		comments: []models.Comment{
			{ID: 1, UserID: 42, MovieID: "tt123", Content: "hello", Edited: true},
		},
		total: 25,
	}
	handler := NewCommentsHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/comments?page=0", nil)
	rec := httptest.NewRecorder()

	serveWithUser(t, 42, http.HandlerFunc(handler.List)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Data []models.Comment `json:"data"`
		Meta struct {
			Total   int `json:"total"`
			Page    int `json:"page"`
			PerPage int `json:"per_page"`
		} `json:"meta"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Meta.Total != 25 || body.Meta.Page != 0 || body.Meta.PerPage != commentPageLimit {
		t.Fatalf("unexpected meta: %+v", body.Meta)
	}
	if len(body.Data) != 1 || !body.Data[0].Edited {
		t.Fatalf("expected edited comment in response, got %+v", body.Data)
	}
}

func TestListByUserReturnsAuthenticatedUsersPaginatedComments(t *testing.T) {
	store := &fakeCommentStore{
		userComments: []models.CommentWithMovie{
			{ID: 1, UserID: 7, MovieID: "tt123", Content: "hello", Movie: models.MovieSmall{ImdbID: "tt123", Title: "Example", Year: "2025", BackdropURL: "backdrop.jpg"}},
		},
		total: 25,
	}
	handler := NewCommentsHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/users/7/comments?page=1", nil)
	req.SetPathValue("id", "7")
	rec := httptest.NewRecorder()

	serveWithUser(t, 7, http.HandlerFunc(handler.ListByUser)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !store.countByUserCalled || !store.listByUserCalled {
		t.Fatalf("expected count and list calls, got count=%v list=%v", store.countByUserCalled, store.listByUserCalled)
	}
	if store.countByUserID != 7 || store.listByUserID != 7 {
		t.Fatalf("expected token user id 7, got count=%d list=%d", store.countByUserID, store.listByUserID)
	}
	if store.requestedLimit != commentPageLimit || store.requestedOffset != commentPageLimit {
		t.Fatalf("expected limit and offset %d, got limit=%d offset=%d", commentPageLimit, store.requestedLimit, store.requestedOffset)
	}

	body := decodeUserCommentListEnvelope(t, rec)
	if body.Meta.Total != 25 || body.Meta.Page != 1 || body.Meta.PerPage != commentPageLimit {
		t.Fatalf("unexpected meta: %+v", body.Meta)
	}
	if len(body.Data) != 1 || body.Data[0].UserID != 7 || body.Data[0].MovieID != "tt123" {
		t.Fatalf("unexpected comments: %+v", body.Data)
	}
}

func TestListByUserAllowsAuthenticatedUserToReadAnotherUsersComments(t *testing.T) {
	store := &fakeCommentStore{userComments: []models.CommentWithMovie{
		{
			ID:      17,
			UserID:  7,
			MovieID: "tt1234567",
			Content: "A very good movie.",
			Movie: models.MovieSmall{
				ImdbID:      "tt1234567",
				Title:       "Example Movie",
				Year:        "2025",
				BackdropURL: "https://example.test/backdrop.jpg",
			},
		},
	}, total: 1}
	handler := NewCommentsHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/users/7/comments", nil)
	req.SetPathValue("id", "7")
	rec := httptest.NewRecorder()

	serveWithUser(t, 42, http.HandlerFunc(handler.ListByUser)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if store.countByUserID != 7 || store.listByUserID != 7 {
		t.Fatalf("expected URL user id 7, got count=%d list=%d", store.countByUserID, store.listByUserID)
	}

	responseBytes := rec.Body.Bytes()
	var body struct {
		Data []models.CommentWithMovie `json:"data"`
	}
	if err := json.Unmarshal(responseBytes, &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Data) != 1 || body.Data[0].MovieID != "tt1234567" || body.Data[0].Movie.ImdbID != "tt1234567" || body.Data[0].Movie.Title != "Example Movie" || body.Data[0].Movie.Year != "2025" || body.Data[0].Movie.BackdropURL != "https://example.test/backdrop.jpg" {
		t.Fatalf("unexpected user comments: %+v", body.Data)
	}

	var raw struct {
		Data []map[string]json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(responseBytes, &raw); err != nil {
		t.Fatalf("decode raw response: %v", err)
	}
	if len(raw.Data) != 1 {
		t.Fatalf("expected one raw comment, got %d", len(raw.Data))
	}
	for _, key := range []string{"movie_id", "movie"} {
		if _, ok := raw.Data[0][key]; !ok {
			t.Fatalf("missing JSON key %q in %+v", key, raw.Data[0])
		}
	}
	var movie map[string]json.RawMessage
	if err := json.Unmarshal(raw.Data[0]["movie"], &movie); err != nil {
		t.Fatalf("decode raw movie: %v", err)
	}
	for _, key := range []string{"imdb_id", "title", "year", "backdrop_url"} {
		if _, ok := movie[key]; !ok {
			t.Fatalf("missing movie JSON key %q in %+v", key, movie)
		}
	}
}

func TestListByUserRejectsMissingAuthContext(t *testing.T) {
	store := &fakeCommentStore{}
	handler := NewCommentsHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/users/7/comments", nil)
	req.SetPathValue("id", "7")
	rec := httptest.NewRecorder()

	handler.ListByUser(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeCommentErrorEnvelope(t, rec).Error.Code; got != "UNAUTHORIZED" {
		t.Fatalf("expected UNAUTHORIZED, got %q", got)
	}
	if store.countByUserCalled || store.listByUserCalled {
		t.Fatalf("store must not be called, got count=%v list=%v", store.countByUserCalled, store.listByUserCalled)
	}
}

func TestListByUserRejectsInvalidUserID(t *testing.T) {
	for _, value := range []string{"not-an-id", "0", "-3"} {
		t.Run(value, func(t *testing.T) {
			store := &fakeCommentStore{}
			handler := NewCommentsHandler(store)

			req := httptest.NewRequest(http.MethodGet, "/users/"+value+"/comments", nil)
			req.SetPathValue("id", value)
			rec := httptest.NewRecorder()

			serveWithUser(t, 7, http.HandlerFunc(handler.ListByUser)).ServeHTTP(rec, req)

			if rec.Code != http.StatusNotFound {
				t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
			}
			if got := decodeCommentErrorEnvelope(t, rec).Error.Code; got != "NOT_FOUND" {
				t.Fatalf("expected NOT_FOUND, got %q", got)
			}
			if store.countByUserCalled || store.listByUserCalled {
				t.Fatalf("store must not be called, got count=%v list=%v", store.countByUserCalled, store.listByUserCalled)
			}
		})
	}
}

func TestListByUserReturnsEmptyJSONList(t *testing.T) {
	store := &fakeCommentStore{}
	handler := NewCommentsHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/users/7/comments", nil)
	req.SetPathValue("id", "7")
	rec := httptest.NewRecorder()

	serveWithUser(t, 7, http.HandlerFunc(handler.ListByUser)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	body := decodeUserCommentListEnvelope(t, rec)
	if body.Data == nil {
		t.Fatal("expected data to be an empty array, got null")
	}
	if len(body.Data) != 0 {
		t.Fatalf("expected no comments, got %+v", body.Data)
	}
	if body.Meta.Total != 0 || body.Meta.Page != 0 || body.Meta.PerPage != commentPageLimit {
		t.Fatalf("unexpected meta: %+v", body.Meta)
	}
}

func TestListByUserInvalidPageFallsBackToZero(t *testing.T) {
	for _, query := range []string{"page=abc", "page=-1"} {
		t.Run(query, func(t *testing.T) {
			store := &fakeCommentStore{userComments: []models.CommentWithMovie{}}
			handler := NewCommentsHandler(store)

			req := httptest.NewRequest(http.MethodGet, "/users/7/comments?"+query, nil)
			req.SetPathValue("id", "7")
			rec := httptest.NewRecorder()

			serveWithUser(t, 7, http.HandlerFunc(handler.ListByUser)).ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
			}
			if !store.countByUserCalled || !store.listByUserCalled {
				t.Fatalf("expected count and list calls, got count=%v list=%v", store.countByUserCalled, store.listByUserCalled)
			}
			if store.requestedOffset != 0 {
				t.Fatalf("expected offset 0, got %d", store.requestedOffset)
			}
			if body := decodeUserCommentListEnvelope(t, rec); body.Meta.Page != 0 {
				t.Fatalf("expected page 0, got %d", body.Meta.Page)
			}
		})
	}
}

func TestListByUserReturnsInternalErrorWhenCountFails(t *testing.T) {
	store := &fakeCommentStore{countByUserErr: errors.New("count failed")}
	handler := NewCommentsHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/users/7/comments", nil)
	req.SetPathValue("id", "7")
	rec := httptest.NewRecorder()

	serveWithUser(t, 7, http.HandlerFunc(handler.ListByUser)).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeCommentErrorEnvelope(t, rec).Error.Code; got != "INTERNAL_ERROR" {
		t.Fatalf("expected INTERNAL_ERROR, got %q", got)
	}
	if !store.countByUserCalled || store.listByUserCalled {
		t.Fatalf("expected only count call, got count=%v list=%v", store.countByUserCalled, store.listByUserCalled)
	}
}

func TestListByUserReturnsInternalErrorWhenListFails(t *testing.T) {
	store := &fakeCommentStore{listByUserErr: errors.New("list failed")}
	handler := NewCommentsHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/users/7/comments", nil)
	req.SetPathValue("id", "7")
	rec := httptest.NewRecorder()

	serveWithUser(t, 7, http.HandlerFunc(handler.ListByUser)).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeCommentErrorEnvelope(t, rec).Error.Code; got != "INTERNAL_ERROR" {
		t.Fatalf("expected INTERNAL_ERROR, got %q", got)
	}
	if !store.countByUserCalled || !store.listByUserCalled {
		t.Fatalf("expected count and list calls, got count=%v list=%v", store.countByUserCalled, store.listByUserCalled)
	}
}

func TestGetInvalidIDReturnsNotFound(t *testing.T) {
	handler := NewCommentsHandler(&fakeCommentStore{})

	req := httptest.NewRequest(http.MethodGet, "/comments/not-an-id", nil)
	req.SetPathValue("id", "not-an-id")
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeCommentErrorEnvelope(t, rec).Error.Code; got != "NOT_FOUND" {
		t.Fatalf("expected NOT_FOUND, got %q", got)
	}
}

func TestGetReturnsItemEnvelope(t *testing.T) {
	store := &fakeCommentStore{
		comment: &models.Comment{ID: 7, UserID: 42, MovieID: "tt123", Content: "hello", Edited: true},
	}
	handler := NewCommentsHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/comments/7", nil)
	req.SetPathValue("id", "7")
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	body := decodeCommentItemEnvelope(t, rec)
	if body.Data.ID != 7 || body.Data.MovieID != "tt123" || body.Data.Content != "hello" {
		t.Fatalf("unexpected comment envelope: %+v", body.Data)
	}
	if !body.Data.Edited {
		t.Fatal("expected edited comment in response")
	}
}

func TestGetMissingCommentReturnsNotFound(t *testing.T) {
	handler := NewCommentsHandler(&fakeCommentStore{findErr: ErrNotFound})

	req := httptest.NewRequest(http.MethodGet, "/comments/404", nil)
	req.SetPathValue("id", "404")
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeCommentErrorEnvelope(t, rec).Error.Code; got != "NOT_FOUND" {
		t.Fatalf("expected NOT_FOUND, got %q", got)
	}
}

func TestUpdateUsesAuthenticatedUserID(t *testing.T) {
	store := &fakeCommentStore{}
	handler := NewCommentsHandler(store)

	req := httptest.NewRequest(http.MethodPatch, "/comments/1", strings.NewReader(`{"content":"edited","user_id":999}`))
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	serveWithUser(t, 42, http.HandlerFunc(handler.Update)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if store.updateUserID != 42 {
		t.Fatalf("expected token user id 42, got %d", store.updateUserID)
	}
	body := decodeCommentItemEnvelope(t, rec)
	if !body.Data.Edited {
		t.Fatal("expected updated comment to be marked as edited")
	}
}

func TestUpdateTrimsContent(t *testing.T) {
	store := &fakeCommentStore{}
	handler := NewCommentsHandler(store)

	req := httptest.NewRequest(http.MethodPatch, "/comments/1", strings.NewReader(`{"content":"  edited  "}`))
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	serveWithUser(t, 42, http.HandlerFunc(handler.Update)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if store.updatedContent != "edited" {
		t.Fatalf("expected trimmed content, got %q", store.updatedContent)
	}
}

func TestUpdateInvalidBodyReturnsFieldValidationError(t *testing.T) {
	store := &fakeCommentStore{}
	handler := NewCommentsHandler(store)

	req := httptest.NewRequest(http.MethodPatch, "/comments/1", strings.NewReader(`{"content":`))
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	serveWithUser(t, 42, http.HandlerFunc(handler.Update)).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Error struct {
			Code   string `json:"code"`
			Fields map[string]struct {
				Message string `json:"message"`
			} `json:"fields"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR, got %q", body.Error.Code)
	}
	if got := body.Error.Fields["body"].Message; got != "Invalid request body" {
		t.Fatalf("expected body field validation, got %q", got)
	}
}

func TestUpdateInvalidContentTypeReturnsContentFieldValidationError(t *testing.T) {
	store := &fakeCommentStore{}
	handler := NewCommentsHandler(store)

	req := httptest.NewRequest(http.MethodPatch, "/comments/1", strings.NewReader(`{"content":436}`))
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	serveWithUser(t, 42, http.HandlerFunc(handler.Update)).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Error struct {
			Code   string `json:"code"`
			Fields map[string]struct {
				Message string `json:"message"`
			} `json:"fields"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR, got %q", body.Error.Code)
	}
	if got := body.Error.Fields["content"].Message; got != "Invalid request body" {
		t.Fatalf("expected content field validation, got %q", got)
	}
	if _, ok := body.Error.Fields["body"]; ok {
		t.Fatalf("did not expect body field validation, got %+v", body.Error.Fields)
	}
}

func TestUpdateMissingCommentWinsOverInvalidBody(t *testing.T) {
	store := &fakeCommentStore{findErr: ErrNotFound}
	handler := NewCommentsHandler(store)

	req := httptest.NewRequest(http.MethodPatch, "/comments/1", strings.NewReader(`{"content":436}`))
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	serveWithUser(t, 42, http.HandlerFunc(handler.Update)).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteUsesAuthenticatedUserID(t *testing.T) {
	store := &fakeCommentStore{}
	handler := NewCommentsHandler(store)

	req := httptest.NewRequest(http.MethodDelete, "/comments/1", strings.NewReader(`{"user_id":999}`))
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	serveWithUser(t, 42, http.HandlerFunc(handler.Delete)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if store.deleteUserID != 42 {
		t.Fatalf("expected token user id 42, got %d", store.deleteUserID)
	}
}

func TestUpdateReturnsNotFoundWhenCommentIsNotOwnedByUser(t *testing.T) {
	store := &fakeCommentStore{updateErr: ErrNotFound}
	handler := NewCommentsHandler(store)

	req := httptest.NewRequest(http.MethodPatch, "/comments/1", strings.NewReader(`{"content":"edited","user_id":999}`))
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	serveWithUser(t, 42, http.HandlerFunc(handler.Update)).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
	if store.updateUserID != 42 {
		t.Fatalf("expected token user id 42, got %d", store.updateUserID)
	}
}

func TestDeleteReturnsNotFoundWhenCommentIsNotOwnedByUser(t *testing.T) {
	store := &fakeCommentStore{deleteErr: ErrNotFound}
	handler := NewCommentsHandler(store)

	req := httptest.NewRequest(http.MethodDelete, "/comments/1", strings.NewReader(`{"user_id":999}`))
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	serveWithUser(t, 42, http.HandlerFunc(handler.Delete)).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
	if store.deleteUserID != 42 {
		t.Fatalf("expected token user id 42, got %d", store.deleteUserID)
	}
}

func serveWithUser(t *testing.T, userID int64, next http.Handler) http.Handler {
	t.Helper()

	tokens, err := auth.NewTokenManager("0123456789abcdef0123456789abcdef", "hypertube-test")
	if err != nil {
		t.Fatalf("new token manager: %v", err)
	}
	token, _, err := tokens.CreateAccessToken(userID)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Authorization", "Bearer "+token)
		auth.RequireAuth(tokens, existingAuthUserChecker{})(next).ServeHTTP(w, r)
	})
}

func decodeCommentItemEnvelope(t *testing.T, rec *httptest.ResponseRecorder) struct {
	Data models.Comment `json:"data"`
} {
	t.Helper()

	var body struct {
		Data models.Comment `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return body
}

func decodeCommentListEnvelope(t *testing.T, rec *httptest.ResponseRecorder) struct {
	Data []models.Comment `json:"data"`
	Meta struct {
		Total   int `json:"total"`
		Page    int `json:"page"`
		PerPage int `json:"per_page"`
	} `json:"meta"`
} {
	t.Helper()

	var body struct {
		Data []models.Comment `json:"data"`
		Meta struct {
			Total   int `json:"total"`
			Page    int `json:"page"`
			PerPage int `json:"per_page"`
		} `json:"meta"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return body
}

func decodeUserCommentListEnvelope(t *testing.T, rec *httptest.ResponseRecorder) struct {
	Data []models.CommentWithMovie `json:"data"`
	Meta struct {
		Total   int `json:"total"`
		Page    int `json:"page"`
		PerPage int `json:"per_page"`
	} `json:"meta"`
} {
	t.Helper()

	var body struct {
		Data []models.CommentWithMovie `json:"data"`
		Meta struct {
			Total   int `json:"total"`
			Page    int `json:"page"`
			PerPage int `json:"per_page"`
		} `json:"meta"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return body
}

func decodeCommentErrorEnvelope(t *testing.T, rec *httptest.ResponseRecorder) struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Fields  map[string]struct {
			Message string `json:"message"`
		} `json:"fields"`
	} `json:"error"`
} {
	t.Helper()

	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
			Fields  map[string]struct {
				Message string `json:"message"`
			} `json:"fields"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return body
}
