package comments

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"hypertube/api/internal/auth"
	"hypertube/api/internal/models"
)

type fakeCommentStore struct {
	updateUserID int
	deleteUserID int
	updateErr    error
	deleteErr    error
}

func (s *fakeCommentStore) findByID(ctx context.Context, id string) (*models.Comment, error) {
	return nil, nil
}

func (s *fakeCommentStore) findAll(ctx context.Context) ([]models.Comment, error) {
	return nil, nil
}

func (s *fakeCommentStore) update(ctx context.Context, content string, id string, userID int) (models.Comment, error) {
	s.updateUserID = userID
	if s.updateErr != nil {
		return models.Comment{}, s.updateErr
	}
	return models.Comment{ID: 1, UserID: userID, Content: content}, nil
}

func (s *fakeCommentStore) delete(ctx context.Context, id string, userID int) error {
	s.deleteUserID = userID
	return s.deleteErr
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
		auth.RequireAuth(tokens)(next).ServeHTTP(w, r)
	})
}
