package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"hypertube/api/internal/movies"
	"hypertube/api/internal/movies/yts"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	db, err := pgxpool.New(ctx, getEnv("DATABASE_URL", "postgres://hypertube:changeme@localhost:5432/hypertube?sslmode=disable"))
	if err != nil {
		log.Fatalf("connect to db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("ping db: %v", err)
	}
	log.Println("connected to database")

	// prepare dependencies for DB and torrent search clients
	store := movies.NewStore(db)
	searchers := []movies.MovieSearcher{
		yts.NewClient(),
	}

	//inject dependencies into handlers
	moviesHandler := movies.NewHandler(store, searchers)

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// // Auth
	// mux.HandleFunc("POST /api/v1/oauth/token", nil)
	// mux.HandleFunc("GET /api/v1/oauth/callback/42", nil)
	// mux.HandleFunc("GET /api/v1/oauth/callback/github", nil)

	// // Users
	// mux.HandleFunc("GET /api/v1/users", nil)
	// mux.HandleFunc("GET /api/v1/users/{id}", nil)
	// mux.HandleFunc("PATCH /api/v1/users/{id}", nil)

	// Movies
	mux.HandleFunc("GET /api/v1/movies", moviesHandler.GetMovies)
	mux.HandleFunc("GET /api/v1/movies/search", moviesHandler.SearchMovies)
	mux.HandleFunc("GET /api/v1/movies/{id}", moviesHandler.GetMoviesId)
	// mux.HandleFunc("GET /api/v1/movies/{id}/comments", moviesHandler.ListComments)
	// mux.HandleFunc("POST /api/v1/movies/{id}/comments", moviesHandler.CreateComment)

	// // Comments
	// mux.HandleFunc("GET /api/v1/comments", nil)
	// mux.HandleFunc("GET /api/v1/comments/{id}", nil)
	// mux.HandleFunc("POST /api/v1/comments", nil)
	// mux.HandleFunc("PATCH /api/v1/comments/{id}", nil)
	// mux.HandleFunc("DELETE /api/v1/comments/{id}", nil)

	addr := ":" + getEnv("PORT", "8080")
	log.Printf("api listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
