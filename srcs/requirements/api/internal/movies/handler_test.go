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
	movies                  []models.Movie
	watched                 []models.WatchedMovie
	comments                []models.Comment
	commentTotal            int
	err                     error
	listWatchedUserID       int
	listWatchedCalled       bool
	createdComment          models.Comment
	savedProgressUserID     int
	savedProgressImdbID     string
	savedProgress           int
	savedComplete           bool
	savedPourcent           int
	saveMovieProgressCalled bool
	saveMovieProgressErr    error
}

func (f *fakeStore) listDefault(_ context.Context) ([]models.Movie, error) {
	return f.movies, f.err
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

func (s *fakeStore) listComments(ctx context.Context, imdbId string, limit, offset int) ([]models.Comment, error) {
	return s.comments, nil
}

func (s *fakeStore) countComments(ctx context.Context, imdbId string) (int, error) {
	if s.commentTotal != 0 {
		return s.commentTotal, nil
	}
	return len(s.comments), nil
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

func (s *fakeStore) HasAppState(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (s *fakeStore) MarkAppState(ctx context.Context, key string) error {
	return nil
}

func (s *fakeStore) listWatched(ctx context.Context, user_id int) ([]models.WatchedMovie, error) {
	s.listWatchedCalled = true
	s.listWatchedUserID = user_id
	return s.watched, s.err
}

func (s *fakeStore) saveMovieProgress(ctx context.Context, userID int, imdbID string, progress int, complete bool, pourcent int) error {
	s.saveMovieProgressCalled = true
	s.savedProgressUserID = userID
	s.savedProgressImdbID = imdbID
	s.savedProgress = progress
	s.savedComplete = complete
	s.savedPourcent = pourcent
	return s.saveMovieProgressErr
}

func (s *fakeStore) listDirectStream(ctx context.Context) ([]models.Movie, error) {
	return nil, nil
}

type fakeUserStore struct {
	users map[int64]models.User
}

type existingAuthUserChecker struct{}

func (existingAuthUserChecker) UserExists(context.Context, int64) (bool, error) {
	return true, nil
}

func (f *fakeUserStore) FindUserByID(_ context.Context, id int64) (models.User, error) {
	if u, ok := f.users[id]; ok {
		return u, nil
	}
	return models.User{}, ErrNotFound
}

type fakeTMDB struct {
	lastLanguage string
}

func (f *fakeTMDB) FindByIMDBID(_ context.Context, imdbID string) (models.Movie, error) {
	return models.Movie{ImdbID: imdbID}, nil
}

func (f *fakeTMDB) GetMovieDetails(_ context.Context, _ string, language string) (models.MovieDetails, error) {
	f.lastLanguage = language
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
	h.GetDefaultMovies(rec, req)

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
	h.GetDefaultMovies(rec, req)

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
	h.GetDefaultMovies(rec, req)

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

func TestSearchMoviesMissingTitleReturnsFieldValidationError(t *testing.T) {
	h := &MoviesHandler{}

	req := httptest.NewRequest(http.MethodGet, "/movies/search", nil)
	rec := httptest.NewRecorder()

	h.SearchMovies(rec, req)

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
	if got := body.Error.Fields["title"].Message; got != "Title query parameter is required" {
		t.Fatalf("expected title field validation, got %q", got)
	}
}

func TestGetMoviesId_OK(t *testing.T) {
	tmdb := &fakeTMDB{}
	h := &MoviesHandler{tmdb: tmdb, store: &fakeStore{
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
	req.Header.Set("Accept-Language", "fr")
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
	if tmdb.lastLanguage != "fr" {
		t.Errorf("expected forwarded language fr, got %q", tmdb.lastLanguage)
	}
}

func TestGetMoviesId_LanguageQueryFallback(t *testing.T) {
	tmdb := &fakeTMDB{}
	h := &MoviesHandler{tmdb: tmdb, store: &fakeStore{
		movies: []models.Movie{{ImdbID: "693134"}},
	}}

	req := httptest.NewRequest(http.MethodGet, "/movies/693134?lang=de", nil)
	req.SetPathValue("id", "693134")
	rec := httptest.NewRecorder()
	h.GetMoviesId(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if tmdb.lastLanguage != "de" {
		t.Errorf("expected forwarded language de, got %q", tmdb.lastLanguage)
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

func TestGetCommentsReturnsNotFoundForMissingMovie(t *testing.T) {
	h := &MoviesHandler{store: &fakeStore{}}

	req := httptest.NewRequest(http.MethodGet, "/movies/tt404/comments", nil)
	req.SetPathValue("id", "tt404")
	rec := httptest.NewRecorder()

	serveWithUser(t, 42, http.HandlerFunc(h.GetComments)).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetCommentsReturnsPaginatedComments(t *testing.T) {
	h := &MoviesHandler{
		store: &fakeStore{
			movies:       []models.Movie{{ImdbID: "tt123"}},
			comments:     []models.Comment{{ID: 1, MovieID: "tt123", UserID: 42, Content: "hello"}},
			commentTotal: 25,
		},
		userStore: &fakeUserStore{users: map[int64]models.User{
			42: {ID: 42, Username: "alice", FirstName: "alice", LastName: "gu"},
		}},
	}

	req := httptest.NewRequest(http.MethodGet, "/movies/tt123/comments?page=0", nil)
	req.SetPathValue("id", "tt123")
	rec := httptest.NewRecorder()

	serveWithUser(t, 42, http.HandlerFunc(h.GetComments)).ServeHTTP(rec, req)

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
}

func TestGetCommentsIncludesUserColor(t *testing.T) {
	h := &MoviesHandler{
		store: &fakeStore{
			movies:   []models.Movie{{ImdbID: "tt123"}},
			comments: []models.Comment{{ID: 1, MovieID: "tt123", UserID: 42, Content: "hello", Edited: true}},
		},
		userStore: &fakeUserStore{users: map[int64]models.User{
			42: {ID: 42, Username: "alice", Color: models.UserColorGreen, FirstName: "alice", LastName: "gu"},
		}},
	}

	req := httptest.NewRequest(http.MethodGet, "/movies/tt123/comments?page=0", nil)
	req.SetPathValue("id", "tt123")
	rec := httptest.NewRecorder()

	serveWithUser(t, 42, http.HandlerFunc(h.GetComments)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Data []models.CommentWithUser `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Data) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(body.Data))
	}
	if got := body.Data[0].User.Color; got != models.UserColorGreen {
		t.Fatalf("expected comment user color %q, got %q", models.UserColorGreen, got)
	}
	if !body.Data[0].Edited {
		t.Fatal("expected edited comment in response")
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

func TestPatchMovieProgressCreatesOrUpdatesForAuthenticatedUser(t *testing.T) {
	store := &fakeStore{movies: []models.Movie{{ImdbID: "tt1234567"}}}
	h := &MoviesHandler{store: store}

	req := httptest.NewRequest(http.MethodPatch, "/movies/tt1234567/progress", strings.NewReader(`{"progress":1804,"complete":false,"pourcent":50}`))
	req.SetPathValue("id", "tt1234567")
	rec := httptest.NewRecorder()

	serveWithUser(t, 42, http.HandlerFunc(h.PatchMovieProgress)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Data movieProgressResponse `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Data.Progress != 1804 || body.Data.Complete {
		t.Fatalf("unexpected response: %+v", body.Data)
	}
	if !store.saveMovieProgressCalled {
		t.Fatal("expected store save to be called")
	}
	if store.savedProgressUserID != 42 || store.savedProgressImdbID != "tt1234567" || store.savedProgress != 1804 || store.savedComplete {
		t.Fatalf("unexpected saved progress: user=%d imdb=%q progress=%d complete=%v pourcent=%v", store.savedProgressUserID, store.savedProgressImdbID, store.savedProgress, store.savedComplete, store.savedPourcent)
	}
}

func TestPatchMovieProgressRequiresAuthentication(t *testing.T) {
	store := &fakeStore{movies: []models.Movie{{ImdbID: "tt1234567"}}}
	h := &MoviesHandler{store: store}

	req := httptest.NewRequest(http.MethodPatch, "/movies/tt1234567/progress", strings.NewReader(`{"progress":1804,"complete":false}`))
	req.SetPathValue("id", "tt1234567")
	rec := httptest.NewRecorder()

	h.PatchMovieProgress(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
	if store.saveMovieProgressCalled {
		t.Fatal("store save must not be called without authentication")
	}
}

func TestPatchMovieProgressRejectsUnknownMovie(t *testing.T) {
	store := &fakeStore{}
	h := &MoviesHandler{store: store}

	req := httptest.NewRequest(http.MethodPatch, "/movies/tt404/progress", strings.NewReader(`{"progress":1804,"complete":false,"pourcent":50}`))
	req.SetPathValue("id", "tt404")
	rec := httptest.NewRecorder()

	serveWithUser(t, 42, http.HandlerFunc(h.PatchMovieProgress)).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeHandlerErrorCode(t, rec); got != "NOT_FOUND" {
		t.Fatalf("expected NOT_FOUND, got %q", got)
	}
	if store.saveMovieProgressCalled {
		t.Fatal("store save must not be called for unknown movies")
	}
}

func TestPatchMovieProgressMissingMovieWinsOverInvalidBody(t *testing.T) {
	store := &fakeStore{}
	h := &MoviesHandler{store: store}

	req := httptest.NewRequest(http.MethodPatch, "/movies/tt404/progress", strings.NewReader(`{"progress":`))
	req.SetPathValue("id", "tt404")
	rec := httptest.NewRecorder()

	serveWithUser(t, 42, http.HandlerFunc(h.PatchMovieProgress)).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeHandlerErrorCode(t, rec); got != "NOT_FOUND" {
		t.Fatalf("expected NOT_FOUND, got %q", got)
	}
	if store.saveMovieProgressCalled {
		t.Fatal("store save must not be called for unknown movies")
	}
}

func TestPatchMovieProgressRejectsMalformedJSON(t *testing.T) {
	store := &fakeStore{movies: []models.Movie{{ImdbID: "tt1234567"}}}
	h := &MoviesHandler{store: store}

	req := httptest.NewRequest(http.MethodPatch, "/movies/tt1234567/progress", strings.NewReader(`{"progress":`))
	req.SetPathValue("id", "tt1234567")
	rec := httptest.NewRecorder()

	serveWithUser(t, 42, http.HandlerFunc(h.PatchMovieProgress)).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeHandlerErrorCode(t, rec); got != "BAD_REQUEST" {
		t.Fatalf("expected BAD_REQUEST, got %q", got)
	}
	if store.saveMovieProgressCalled {
		t.Fatal("store save must not be called for malformed JSON")
	}
}

func TestPatchMovieProgressRejectsUnknownField(t *testing.T) {
	store := &fakeStore{movies: []models.Movie{{ImdbID: "tt1234567"}}}
	h := &MoviesHandler{store: store}

	req := httptest.NewRequest(http.MethodPatch, "/movies/tt1234567/progress", strings.NewReader(`{"progress":1804,"complete":false,"pourcent":50,"extra":true}`))
	req.SetPathValue("id", "tt1234567")
	rec := httptest.NewRecorder()

	serveWithUser(t, 42, http.HandlerFunc(h.PatchMovieProgress)).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeHandlerErrorCode(t, rec); got != "BAD_REQUEST" {
		t.Fatalf("expected BAD_REQUEST, got %q", got)
	}
	if store.saveMovieProgressCalled {
		t.Fatal("store save must not be called for unknown fields")
	}
}

func TestPatchMovieProgressRejectsMultipleJSONDocuments(t *testing.T) {
	store := &fakeStore{movies: []models.Movie{{ImdbID: "tt1234567"}}}
	h := &MoviesHandler{store: store}

	req := httptest.NewRequest(http.MethodPatch, "/movies/tt1234567/progress", strings.NewReader(`{"progress":1804,"complete":false,"pourcent":50} {}`))
	req.SetPathValue("id", "tt1234567")
	rec := httptest.NewRecorder()

	serveWithUser(t, 42, http.HandlerFunc(h.PatchMovieProgress)).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := decodeHandlerErrorCode(t, rec); got != "BAD_REQUEST" {
		t.Fatalf("expected BAD_REQUEST, got %q", got)
	}
	if store.saveMovieProgressCalled {
		t.Fatal("store save must not be called for multiple JSON documents")
	}
}

func TestPatchMovieProgressRejectsMissingFields(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		missingField string
	}{
		{name: "missing complete", body: `{"progress":1804,"pourcent":50}`, missingField: "complete"},
		{name: "missing progress", body: `{"complete":false,"pourcent":50}`, missingField: "progress"},
		{name: "missing pourcent", body: `{"progress":1804,"complete":false}`, missingField: "pourcent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeStore{movies: []models.Movie{{ImdbID: "tt1234567"}}}
			h := &MoviesHandler{store: store}

			req := httptest.NewRequest(http.MethodPatch, "/movies/tt1234567/progress", strings.NewReader(tt.body))
			req.SetPathValue("id", "tt1234567")
			rec := httptest.NewRecorder()

			serveWithUser(t, 42, http.HandlerFunc(h.PatchMovieProgress)).ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
			fields := decodeHandlerValidationFields(t, rec)
			if _, ok := fields[tt.missingField]; !ok {
				t.Fatalf("expected field %q, got %+v", tt.missingField, fields)
			}
			if store.saveMovieProgressCalled {
				t.Fatal("store save must not be called for missing fields")
			}
		})
	}
}

func TestPatchMovieProgressRejectsNullOrWrongTypeFields(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "progress null", body: `{"progress":null,"complete":false,"pourcent":50}`},
		{name: "complete null", body: `{"progress":1804,"complete":null,"pourcent":50}`},
		{name: "progress string", body: `{"progress":"1804","complete":false,"pourcent":50}`},
		{name: "complete string", body: `{"progress":1804,"complete":"false","pourcent":50}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &fakeStore{movies: []models.Movie{{ImdbID: "tt1234567"}}}
			h := &MoviesHandler{store: store}

			req := httptest.NewRequest(http.MethodPatch, "/movies/tt1234567/progress", strings.NewReader(tt.body))
			req.SetPathValue("id", "tt1234567")
			rec := httptest.NewRecorder()

			serveWithUser(t, 42, http.HandlerFunc(h.PatchMovieProgress)).ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
			if got := decodeHandlerErrorCode(t, rec); got != "VALIDATION_ERROR" {
				t.Fatalf("expected VALIDATION_ERROR, got %q", got)
			}
			if store.saveMovieProgressCalled {
				t.Fatal("store save must not be called for invalid fields")
			}
		})
	}
}

func TestPatchMovieProgressRejectsNegativeProgress(t *testing.T) {
	store := &fakeStore{movies: []models.Movie{{ImdbID: "tt1234567"}}}
	h := &MoviesHandler{store: store}

	req := httptest.NewRequest(http.MethodPatch, "/movies/tt1234567/progress", strings.NewReader(`{"progress":-1,"complete":false,"pourcent":50}`))
	req.SetPathValue("id", "tt1234567")
	rec := httptest.NewRecorder()

	serveWithUser(t, 42, http.HandlerFunc(h.PatchMovieProgress)).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	fields := decodeHandlerValidationFields(t, rec)
	if _, ok := fields["progress"]; !ok {
		t.Fatalf("expected progress field error, got %+v", fields)
	}
	if store.saveMovieProgressCalled {
		t.Fatal("store save must not be called for negative progress")
	}
}

func TestPatchMovieProgressReturnsInternalError(t *testing.T) {
	store := &fakeStore{
		movies:               []models.Movie{{ImdbID: "tt1234567"}},
		saveMovieProgressErr: errors.New("db down"),
	}
	h := &MoviesHandler{store: store}

	req := httptest.NewRequest(http.MethodPatch, "/movies/tt1234567/progress", strings.NewReader(`{"progress":1804,"complete":false,"pourcent":50}`))
	req.SetPathValue("id", "tt1234567")
	rec := httptest.NewRecorder()

	serveWithUser(t, 42, http.HandlerFunc(h.PatchMovieProgress)).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
	if !store.saveMovieProgressCalled {
		t.Fatal("expected store save to be called")
	}
	if got := decodeHandlerErrorCode(t, rec); got != "INTERNAL_ERROR" {
		t.Fatalf("expected INTERNAL_ERROR, got %q", got)
	}
}

func TestGetUserFilmHistoryAllowsReadingAnotherUsersHistory(t *testing.T) {
	store := &fakeStore{watched: []models.WatchedMovie{{
		ImdbID: "tt1234567", Title: "Example Movie", Year: "2025", PosterURL: "poster.jpg",
		BackdropURL: "backdrop.jpg", Note: 8.1, Genre: []int{12, 18}, Progress: 1804, Complete: false,
	}}}
	h := &MoviesHandler{store: store}

	req := httptest.NewRequest(http.MethodGet, "/users/7/movie-history", nil)
	req.SetPathValue("id", "7")
	rec := httptest.NewRecorder()
	serveWithUser(t, 42, http.HandlerFunc(h.GetUserFilmHistory)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if store.listWatchedUserID != 7 {
		t.Fatalf("expected URL user id 7, got %d", store.listWatchedUserID)
	}
	var body struct {
		Data []movieHistoryResponse `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Data) != 1 || body.Data[0].ImdbID != "tt1234567" || body.Data[0].Title != "Example Movie" || body.Data[0].Year != "2025" || body.Data[0].PosterURL != "poster.jpg" || body.Data[0].BackdropURL != "backdrop.jpg" || body.Data[0].Note != 8.1 || len(body.Data[0].Genre) != 2 {
		t.Fatalf("unexpected history response: %+v", body.Data)
	}
}

func TestGetUserFilmHistoryReturnsUpdatedProgressFields(t *testing.T) {
	store := &fakeStore{watched: []models.WatchedMovie{{
		ImdbID: "tt1234567", Title: "Example Movie", Year: "2025", PosterURL: "poster.jpg",
		BackdropURL: "backdrop.jpg", Note: 8.1, Genre: []int{12, 18}, Progress: 1804, Complete: false, Pourcent: 50,
	}}}
	h := &MoviesHandler{store: store}

	req := httptest.NewRequest(http.MethodGet, "/users/7/movie-history", nil)
	req.SetPathValue("id", "7")
	rec := httptest.NewRecorder()
	serveWithUser(t, 42, http.HandlerFunc(h.GetUserFilmHistory)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Data []movieHistoryResponse `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Data) != 1 {
		t.Fatalf("expected one history entry, got %d", len(body.Data))
	}
	if body.Data[0].Progress != 1804 || body.Data[0].Complete {
		t.Fatalf("unexpected progress fields: %+v", body.Data[0])
	}
}

func TestGetUserFilmHistoryRequiresAuthentication(t *testing.T) {
	store := &fakeStore{}
	h := &MoviesHandler{store: store}
	req := httptest.NewRequest(http.MethodGet, "/users/7/movie-history", nil)
	req.SetPathValue("id", "7")
	rec := httptest.NewRecorder()

	h.GetUserFilmHistory(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
	if store.listWatchedCalled {
		t.Fatal("store must not be called without authentication")
	}
}

func TestGetUserFilmHistoryRejectsInvalidUserID(t *testing.T) {
	for _, value := range []string{"abc", "0", "-1"} {
		t.Run(value, func(t *testing.T) {
			store := &fakeStore{}
			h := &MoviesHandler{store: store}
			req := httptest.NewRequest(http.MethodGet, "/users/"+value+"/movie-history", nil)
			req.SetPathValue("id", value)
			rec := httptest.NewRecorder()

			serveWithUser(t, 42, http.HandlerFunc(h.GetUserFilmHistory)).ServeHTTP(rec, req)

			if rec.Code != http.StatusNotFound {
				t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
			}
			if store.listWatchedCalled {
				t.Fatal("store must not be called for an invalid user ID")
			}
		})
	}
}

func TestGetUserFilmHistoryReturnsInternalError(t *testing.T) {
	store := &fakeStore{err: errors.New("db down")}
	h := &MoviesHandler{store: store}
	req := httptest.NewRequest(http.MethodGet, "/users/7/movie-history", nil)
	req.SetPathValue("id", "7")
	rec := httptest.NewRecorder()

	serveWithUser(t, 42, http.HandlerFunc(h.GetUserFilmHistory)).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetUserFilmHistoryReturnsEmptyJSONList(t *testing.T) {
	store := &fakeStore{}
	h := &MoviesHandler{store: store}
	req := httptest.NewRequest(http.MethodGet, "/users/7/movie-history", nil)
	req.SetPathValue("id", "7")
	rec := httptest.NewRecorder()

	serveWithUser(t, 42, http.HandlerFunc(h.GetUserFilmHistory)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Data []movieHistoryResponse `json:"data"`
		Meta struct {
			Total   int `json:"total"`
			Page    int `json:"page"`
			PerPage int `json:"per_page"`
		} `json:"meta"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Data == nil || len(body.Data) != 0 {
		t.Fatalf("expected data [], got %+v", body.Data)
	}
	if body.Meta.Total != 0 || body.Meta.Page != 0 || body.Meta.PerPage != 0 {
		t.Fatalf("unexpected meta: %+v", body.Meta)
	}
}

func TestPostCommentUsesAuthenticatedUserIDAndPathMovieID(t *testing.T) {
	store := &fakeStore{movies: []models.Movie{{ImdbID: "tt123"}}}
	h := &MoviesHandler{store: store}

	req := httptest.NewRequest(http.MethodPost, "/movies/tt123/comments", strings.NewReader(`{"user_id":999,"movie_id":"tt999","content":"  hello  "}`))
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
	if store.createdComment.Content != "hello" {
		t.Fatalf("expected trimmed content, got %q", store.createdComment.Content)
	}

	var body struct {
		Data struct {
			Edited *bool `json:"edited"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Data.Edited == nil {
		t.Fatal("expected edited field in response")
	}
	if *body.Data.Edited {
		t.Fatal("expected new comment to be unedited")
	}
}

func TestPostCommentMissingMovieWinsOverInvalidBody(t *testing.T) {
	h := &MoviesHandler{store: &fakeStore{}}

	req := httptest.NewRequest(http.MethodPost, "/movies/tt404/comments", strings.NewReader(`{"content":436}`))
	req.SetPathValue("id", "tt404")
	rec := httptest.NewRecorder()

	serveWithUser(t, 42, http.HandlerFunc(h.PostComment)).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPostCommentInvalidContentTypeReturnsContentFieldValidationError(t *testing.T) {
	h := &MoviesHandler{store: &fakeStore{movies: []models.Movie{{ImdbID: "tt123"}}}}

	req := httptest.NewRequest(http.MethodPost, "/movies/tt123/comments", strings.NewReader(`{"content":436}`))
	req.SetPathValue("id", "tt123")
	rec := httptest.NewRecorder()

	serveWithUser(t, 42, http.HandlerFunc(h.PostComment)).ServeHTTP(rec, req)

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
}

func decodeHandlerErrorCode(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return body.Error.Code
}

func decodeHandlerValidationFields(t *testing.T, rec *httptest.ResponseRecorder) map[string]struct {
	Message string `json:"message"`
} {
	t.Helper()

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
	return body.Error.Fields
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
