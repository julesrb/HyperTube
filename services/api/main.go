package main

import (
	"log"
	"net/http"
	"os"
	"hypertube/api/internal/movies"

)

func main() {
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

	// // Movies
	mux.HandleFunc("GET /movies", movies.GetMovies)
	// mux.HandleFunc("GET /movies/{id}", nil)
	// mux.HandleFunc("GET /movies/{id}/comments", nil)
	// mux.HandleFunc("POST /movies/{id}/comments", nil)

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
