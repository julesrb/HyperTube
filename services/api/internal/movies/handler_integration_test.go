package movies

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

//TODO consider writting new integration test

func setupTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://hypertube:changeme@localhost:5432/hypertube?sslmode=disable"
	}

	db, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("connect to db: %v", err)
	}
	t.Cleanup(db.Close)

	if err := db.Ping(context.Background()); err != nil {
		t.Skip("Database is not present, test aborted! ping db: ", err)
	}

	store := NewStore(db)
	handler := NewMoviesHandler(store, nil, nil)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /movies", handler.GetMovies)
	mux.HandleFunc("GET /movies/{id}", handler.GetMoviesId)

	return httptest.NewServer(mux)
}

func TestIntegration_GetMovies_Returns200WithData(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	log.Printf("test server running at %s", srv.URL)

	resp, err := http.Get(srv.URL + "/movies")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body struct {
		Data []movieResponse `json:"data"`
		Meta struct {
			Total int `json:"total"`
		} `json:"meta"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(body.Data) == 0 {
		t.Fatal("expected data length > 0, got 0")
	}
	log.Printf("got %d movies", body.Meta.Total)
}
