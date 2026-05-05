package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"hypertube/api/internal/movies"
	"hypertube/api/internal/movies/c411"
	"hypertube/api/internal/movies/tmdb"

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
	c411Client, err := c411.NewClient()
	if err != nil {
		log.Fatalf("init C411 client: %v", err)
	}
	tmdbClient, err := tmdb.NewClient()
	if err != nil {
		log.Fatalf("init TMDB client: %v", err)
	}
	searchers := []movies.MovieSearcher{c411Client}

	moviesHandler := movies.NewHandler(store, searchers, tmdbClient)

	// Fetch top movies at startup to cache the frontpage content
	featured, err := c411Client.GetTopMovies(ctx)
	if err != nil {
		log.Printf("startup: failed to fetch top movies: %v", err)
	} else {
		log.Printf("startup: top %d movies by seeds:", len(featured))
		for _, t := range featured {
			log.Printf("  imdb=%s seeds=%s title=%s", t.ImdbID, t.Seeds, t.Title)
		}
	}
	for i, torrent := range featured {
		movie, err := tmdbClient.FindByIMDBID(ctx, torrent.ImdbID)
		if err != nil {
			log.Printf("TMDB lookup error for IMDb ID %s: %v", torrent.ImdbID, err)
			continue
		}
		if err = store.UpsertMovie(ctx, movie); err != nil {
			log.Println("db err:", err)
			return
		}
		if err = store.UpsertTorrent(ctx, torrent); err != nil {
			log.Printf("startup: failed to store torrent %s: %v", torrent.Title, err)
		}
		if err = store.UpsertFeatured(ctx, torrent.ImdbID, i); err != nil {
			log.Printf("startup: failed to store featured torrent %s: %v", torrent.Title, err)
		}
	}

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
	mux.HandleFunc("GET /api/v1/movies/{id}/torrents", moviesHandler.GetMovieTorrents)
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
