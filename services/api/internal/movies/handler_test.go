package movies

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

type fakeStore struct {
	movies            []models.Movie
	err               error
	listWatchedUserID int
	createdComment    models.Comment
}

func (f *fakeStore) listFeatured(_ context.Context) ([]models.Movie, error) {
	return f.movies, f.err
}

func (f *fakeStore) findByID(_ context.Context, id string) (*models.Movie, error) {
	for _, m := range f.movies {
		if m.ImdbID == id {
			return &m, nil
		}
	}
	return nil, ErrNotFound
}

func (f *fakeStore) UpsertTorrent(ctx context.Context, ts models.Torrent) error {
	return nil
}

func (f *fakeStore) UpsertMovie(ctx context.Context, m models.Movie) error {
	return nil
}

func (f *fakeStore) findTorrent(ctx context.Context, imdbID string) ([]models.Torrent, error) {
	return nil, nil
}

func (s *fakeStore) listComments(ctx context.Context, imdbId string) ([]models.Comment, error) {
	return nil, nil
}

func (s *fakeStore) createComment(ctx context.Context, c models.Comment) (models.Comment, error) {
	s.createdComment = c
	c.ID = 1
	return c, nil
}

func (s *fakeStore) countSearchResults(ctx context.Context, query string) (int, error) {
	return 0, nil
}

func (s *fakeStore) upsertSearchResults(ctx context.Context, query string, imdbIDs []string) error {
	return nil
}

func (s *fakeStore) listSearchResults(ctx context.Context, query string, limit, offset int) ([]models.Movie, error) {
	return nil, nil
}

func (s *fakeStore) listWatched(ctx context.Context, user_id int) ([]models.Movie, error) {
	s.listWatchedUserID = user_id
	return nil, nil
}

func (s *fakeStore) listDirectStream(ctx context.Context) ([]models.Movie, error) {
	return nil, nil
}

type fakeTMDB struct{}

func (f *fakeTMDB) FindByIMDBID(_ context.Context, imdbID string) (models.Movie, error) {
	return models.Movie{ImdbID: imdbID}, nil
}

func (f *fakeTMDB) GetMovieDetails(_ context.Context, _ string, _ string) (models.MovieDetails, error) {
	return models.MovieDetails{
		Summary:  "A desert planet epic.",
		Director: []string{"Denis Villeneuve"},
		Cast:     []string{"Timothee Chalamet"},
		Runtime:  166,
	}, nil
}

func (f *fakeTMDB) FindByName(ctx context.Context, title string, year int) (models.Movie, error) {
	return models.Movie{}, nil
}

func TestGetMovies_OK(t *testing.T) {
	h := &MoviesHandler{store: &fakeStore{
		movies: []models.Movie{
			{ImdbID: "1", Title: "Dune: Part Two", PosterURL: "poster.jpg", BackdropURL: "backdrop.jpg"},
			{ImdbID: "2", Title: "Avatar", PosterURL: "poster2.jpg", BackdropURL: "backdrop2.jpg"},
		},
	}}

	req := httptest.NewRequest(http.MethodGet, "/movies", nil)
	rec := httptest.NewRecorder()
	h.GetMovies(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body struct {
		Data []movieResponse `json:"data"`
		Meta struct {
			Total   int `json:"total"`
			Page    int `json:"page"`
			PerPage int `json:"per_page"`
		} `json:"meta"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.Data[0].ImdbID != "1" || body.Data[1].ImdbID != "2" {
		t.Errorf("unexpected order: %+v", body.Data)
	}

	if body.Meta.Total != 2 || body.Meta.Page != 0 || body.Meta.PerPage != 2 {
		t.Errorf("unexpected meta: %+v", body.Meta)
	}
}

func TestGetMovies_Empty(t *testing.T) {
	h := &MoviesHandler{store: &fakeStore{movies: []models.Movie{}}}

	req := httptest.NewRequest(http.MethodGet, "/movies", nil)
	rec := httptest.NewRecorder()
	h.GetMovies(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body struct {
		Data []movieResponse `json:"data"`
		Meta struct {
			Total   int `json:"total"`
			Page    int `json:"page"`
			PerPage int `json:"per_page"`
		} `json:"meta"`
	}
	json.NewDecoder(rec.Body).Decode(&body)

	if len(body.Data) != 0 {
		t.Errorf("expected empty data, got %+v", body.Data)
	}

	if body.Meta.Total != 0 || body.Meta.Page != 0 || body.Meta.PerPage != 0 {
		t.Errorf("unexpected meta: %+v", body.Meta)
	}
}

func TestGetMovies_StoreError(t *testing.T) {
	h := &MoviesHandler{store: &fakeStore{err: errors.New("db down")}}

	req := httptest.NewRequest(http.MethodGet, "/movies", nil)
	rec := httptest.NewRecorder()
	h.GetMovies(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	json.NewDecoder(rec.Body).Decode(&body)

	if body.Error.Code != "INTERNAL_ERROR" {
		t.Errorf("expected INTERNAL_ERROR, got %q", body.Error.Code)
	}
}

func TestGetMoviesId_OK(t *testing.T) {
	h := &MoviesHandler{tmdb: &fakeTMDB{}, store: &fakeStore{
		movies: []models.Movie{
			{
				ImdbID:      "693134",
				Title:       "Dune: Part Two",
				Year:        "2024",
				PosterURL:   "https://image.tmdb.org/t/p/original/rjmLNTt5tP1obYx4YFzLHpN7KcG.jpg",
				BackdropURL: "https://image.tmdb.org/t/p/original/oBCR7ShGq9ZdnHMK8SGOckGpEgo.jpg",
				Note:        8.1,
				Genre:       []int{878, 12, 18},
			},
			{
				ImdbID:      "83533",
				Title:       "Avatar: Fire and Ash",
				Year:        "2025",
				PosterURL:   "https://image.tmdb.org/t/p/original/lE9KpVwgeWHMwgwkNaeH5nEFh20.jpg",
				BackdropURL: "https://image.tmdb.org/t/p/original/u8DU5fkLoM5tTRukzPC31oGPxaQ.jpg",
				Note:        7.4,
				Genre:       []int{878, 12, 14},
			},
		},
	}}

	req := httptest.NewRequest(http.MethodGet, "/movies/693134", nil)
	req.SetPathValue("id", "693134")
	rec := httptest.NewRecorder()
	h.GetMoviesId(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body struct {
		Data movieDetailResponse `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.Data.Title != "Dune: Part Two" {
		t.Errorf("expected movie Title 'Dune: Part Two', got %q", body.Data.Title)
	}

	if body.Data.Director != "Denis Villeneuve" {
		t.Errorf("expected director 'Denis Villeneuve', got %q", body.Data.Director)
	}
}

func TestGetMoviesId_NotFound(t *testing.T) {
	h := &MoviesHandler{store: &fakeStore{
		movies: []models.Movie{
			{ImdbID: "1", Title: "Dune: Part Two"},
			{ImdbID: "2", Title: "Avatar"},
		},
	}}

	req := httptest.NewRequest(http.MethodGet, "/movies/999", nil)
	req.SetPathValue("id", "999")
	rec := httptest.NewRecorder()
	h.GetMoviesId(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	json.NewDecoder(rec.Body).Decode(&body)

	if body.Error.Code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND, got %q", body.Error.Code)
	}
}

func TestGetWatchedMoviesUsesAuthenticatedUserID(t *testing.T) {
	store := &fakeStore{}
	h := &MoviesHandler{store: store}

	req := httptest.NewRequest(http.MethodGet, "/movies/watched", strings.NewReader(`{"user_id":999}`))
	rec := httptest.NewRecorder()

	serveWithUser(t, 42, http.HandlerFunc(h.GetWatchedMovies)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if store.listWatchedUserID != 42 {
		t.Fatalf("expected token user id 42, got %d", store.listWatchedUserID)
	}
}

func TestPostCommentUsesAuthenticatedUserIDAndPathMovieID(t *testing.T) {
	store := &fakeStore{}
	h := &MoviesHandler{store: store}

	req := httptest.NewRequest(http.MethodPost, "/movies/tt123/comments", strings.NewReader(`{"user_id":999,"movie_id":"tt999","content":"hello"}`))
	req.SetPathValue("id", "tt123")
	rec := httptest.NewRecorder()

	serveWithUser(t, 42, http.HandlerFunc(h.PostComment)).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if store.createdComment.UserID != 42 {
		t.Fatalf("expected token user id 42, got %d", store.createdComment.UserID)
	}
	if store.createdComment.MovieID != "tt123" {
		t.Fatalf("expected path movie id tt123, got %q", store.createdComment.MovieID)
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
