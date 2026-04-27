package movies

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"hypertube/api/internal/models"
)

type fakeStore struct {
	movies []models.Movie
	err    error
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

func TestGetMovies_OK(t *testing.T) {
	h := &Handler{store: &fakeStore{
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
			Total int `json:"total"`
		} `json:"meta"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.Meta.Total != 2 {
		t.Errorf("expected total=2, got %d", body.Meta.Total)
	}
	if body.Data[0].ImdbID != "1" || body.Data[1].ImdbID != "2" {
		t.Errorf("unexpected order: %+v", body.Data)
	}
}

func TestGetMovies_Empty(t *testing.T) {
	h := &Handler{store: &fakeStore{movies: []models.Movie{}}}

	req := httptest.NewRequest(http.MethodGet, "/movies", nil)
	rec := httptest.NewRecorder()
	h.GetMovies(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body struct {
		Data []movieResponse `json:"data"`
		Meta struct {
			Total int `json:"total"`
		} `json:"meta"`
	}
	json.NewDecoder(rec.Body).Decode(&body)

	if body.Meta.Total != 0 {
		t.Errorf("expected total=0, got %d", body.Meta.Total)
	}
	if len(body.Data) != 0 {
		t.Errorf("expected empty data, got %+v", body.Data)
	}
}

func TestGetMovies_StoreError(t *testing.T) {
	h := &Handler{store: &fakeStore{err: errors.New("db down")}}

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
	h := &Handler{store: &fakeStore{
		movies: []models.Movie{
			{
				ImdbID:      "693134",
				Title:       "Dune: Part Two",
				Year:        "2024",
				PosterURL:   "https://image.tmdb.org/t/p/original/rjmLNTt5tP1obYx4YFzLHpN7KcG.jpg",
				BackdropURL: "https://image.tmdb.org/t/p/original/oBCR7ShGq9ZdnHMK8SGOckGpEgo.jpg",
				IMDbRating:  8.1,
				Genre:       []int{878, 12, 18},
				Runtime:     167,
				Summary:     "Follow the mythic journey of Paul Atreides as he unites with Chani and the Fremen while on a path of revenge against the conspirators who destroyed his family.",
				Director:    "Denis Villeneuve",
				Cast:        []string{"Timothée Chalamet", "Zendaya", "Rebecca Ferguson", "Josh Brolin"},
				Watched:     false,
			},
			{
				ImdbID:      "83533",
				Title:       "Avatar: Fire and Ash",
				Year:        "2025",
				PosterURL:   "https://image.tmdb.org/t/p/original/lE9KpVwgeWHMwgwkNaeH5nEFh20.jpg",
				BackdropURL: "https://image.tmdb.org/t/p/original/u8DU5fkLoM5tTRukzPC31oGPxaQ.jpg",
				IMDbRating:  7.4,
				Genre:       []int{878, 12, 14},
				Runtime:     197,
				Summary:     "Following a devastating conflict with the RDA and the death of their eldest son, Jake Sully and Neytiri face a new threat on Pandora.",
				Director:    "James Cameron",
				Cast:        []string{"Sam Worthington", "Zoe Saldaña", "Sigourney Weaver", "Stephen Lang"},
				Watched:     false,
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
		Data movieResponse `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.Data.Title != "Dune: Part Two" {
		t.Errorf("expected movie Title 'Dune: Part Two', got %q", body.Data.Title)
	}
}

func TestGetMoviesId_NotFound(t *testing.T) {
	h := &Handler{store: &fakeStore{
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
