package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"hypertube/api/internal/movies"
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

	moviesHandler := movies.NewHandler(db)

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// // Auth
	// mux.HandleFunc("POST /oauth/token", nil)
	// mux.HandleFunc("GET /oauth/callback/42", nil)
	// mux.HandleFunc("GET /oauth/callback/github", nil)

	// // Users
	// mux.HandleFunc("GET /users", nil)
	// mux.HandleFunc("GET /users/{id}", nil)
	// mux.HandleFunc("PATCH /users/{id}", nil)

	// Movies
	mux.HandleFunc("GET /movies", moviesHandler.GetMovies)
	mux.HandleFunc("GET /movies/{id}", moviesHandler.GetMoviesId)
	// mux.HandleFunc("GET /movies/{id}/comments", moviesHandler.ListComments)
	// mux.HandleFunc("POST /movies/{id}/comments", moviesHandler.CreateComment)

	// // Comments
	// mux.HandleFunc("GET /comments", nil)
	// mux.HandleFunc("GET /comments/{id}", nil)
	// mux.HandleFunc("POST /comments", nil)
	// mux.HandleFunc("PATCH /comments/{id}", nil)
	// mux.HandleFunc("DELETE /comments/{id}", nil)

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
