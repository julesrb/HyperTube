package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"hypertube/api/internal/auth"
	"hypertube/api/internal/comments"
	"hypertube/api/internal/email"
	"hypertube/api/internal/movies"
	"hypertube/api/internal/movies/archive.org"
	"hypertube/api/internal/movies/c411"
	"hypertube/api/internal/movies/tmdb"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	db := connectDB(ctx)
	defer db.Close()

	tokenManager, err := auth.NewTokenManager(os.Getenv("JWT_SECRET"), getEnv("JWT_ISSUER", "hypertube-api"))
	if err != nil {
		log.Fatalf("init JWT manager: %v", err)
	}

	authStore := auth.NewStore(db)
	fortyTwoRedirectURL := getEnv("FORTYTWO_REDIRECT_URL", "http://localhost:8080/api/v1/auth/42/callback")
	authOptions := []auth.HandlerOption{
		auth.WithFrontendAuthCallbackURL(getEnv("FRONTEND_AUTH_CALLBACK_URL", "http://localhost:4200/auth/callback")),
		auth.WithPasswordResetURL(getEnv("PASSWORD_RESET_URL", "http://localhost:4200/{locale}/reset-password")),
		auth.WithPasswordResetTTL(getPasswordResetTTL()),
		auth.WithFortyTwoOAuth(auth.NewFortyTwoOAuth(auth.FortyTwoOAuthConfig{
			ClientID:     os.Getenv("FORTYTWO_CLIENT_ID"),
			ClientSecret: os.Getenv("FORTYTWO_CLIENT_SECRET"),
			RedirectURL:  fortyTwoRedirectURL,
		})),
	}
	if passwordResetMailer := newPasswordResetMailer(); passwordResetMailer != nil {
		authOptions = append(authOptions, auth.WithPasswordResetMailer(passwordResetMailer))
	}
	authHandler := auth.NewHandler(authStore, tokenManager, authOptions...)

	movieStore := movies.NewStore(db)
	commentStore := comments.NewStore(db)

	c411Client, err := c411.NewClient()
	if err != nil {
		log.Fatalf("init C411 client: %v", err)
	}
	archiveClient, err := archiveorg.NewClient()
	if err != nil {
		log.Fatalf("init archive.org client: %v", err)
	}
	tmdbClient, err := tmdb.NewClient()
	if err != nil {
		log.Fatalf("init TMDB client: %v", err)
	}

	seedFeatured(ctx, c411Client, tmdbClient, movieStore)

	searchers := []movies.MovieSearcher{c411Client, archiveClient}
	moviesHandler := movies.NewMoviesHandler(movieStore, searchers, tmdbClient)
	commentsHandler := comments.NewCommentsHandler(commentStore)

	addr := ":" + getEnv("PORT", "8080")
	log.Printf("api listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, newRouter(moviesHandler, commentsHandler, authHandler, tokenManager)))
}

func connectDB(ctx context.Context) *pgxpool.Pool {
	db, err := pgxpool.New(ctx, getEnv("DATABASE_URL", "postgres://hypertube:changeme@localhost:5432/hypertube?sslmode=disable"))
	if err != nil {
		log.Fatalf("connect to db: %v", err)
	}
	if err := db.Ping(ctx); err != nil {
		log.Fatalf("ping db: %v", err)
	}
	log.Println("connected to database")
	return db
}

func seedFeatured(ctx context.Context, c411Client *c411.Client, tmdbClient *tmdb.Client, store *movies.Store) {
	featured, err := c411Client.GetTopMovies(ctx)
	if err != nil {
		log.Printf("startup: failed to fetch top movies: %v", err)
		return
	}
	log.Printf("startup: top %d movies by seeds:", len(featured))
	for _, t := range featured {
		log.Printf("  imdb=%s seeds=%s title=%s", t.ImdbID, t.Seeds, t.Title)
	}
	for i, torrent := range featured {
		movie, err := tmdbClient.FindByIMDBID(ctx, torrent.ImdbID)
		if err != nil {
			log.Printf("TMDB lookup error for IMDb ID %s: %v", torrent.ImdbID, err)
			continue
		}
		if err = store.UpsertMovie(ctx, movie); err != nil {
			log.Fatalf("startup: db err: %v", err)
		}
		if err = store.UpsertTorrent(ctx, torrent); err != nil {
			log.Printf("startup: failed to store torrent %s: %v", torrent.Title, err)
		}
		if err = store.UpsertFeatured(ctx, torrent.ImdbID, i); err != nil {
			log.Printf("startup: failed to store featured torrent %s: %v", torrent.Title, err)
		}
	}
}

func newRouter(
	moviesHandler *movies.MoviesHandler,
	commentsHandler *comments.CommentsHandler,
	authHandler *auth.Handler,
	tokenManager *auth.TokenManager,
) chi.Router {
	r := chi.NewRouter()

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/password-reset", authHandler.RequestPasswordReset)
			r.Post("/reset-password", authHandler.ResetPassword)
			r.Get("/42/login", authHandler.LoginFortyTwo)
			r.Get("/42/callback", authHandler.CallbackFortyTwo)
		})

		r.Post("/oauth/token", authHandler.OAuthToken)

		r.Get("/movies", moviesHandler.GetMovies)

		r.Group(func(r chi.Router) {
			r.Use(auth.RequireAuth(tokenManager))

			r.Get("/movies/watched", moviesHandler.GetWatchedMovies)
			r.Get("/movies/directstream", moviesHandler.GetDirectStreamMovies)
			r.Get("/movies/search", moviesHandler.SearchMovies)
			r.Get("/movies/{id}", moviesHandler.GetMoviesId)
			r.Get("/movies/{id}/torrents", moviesHandler.GetMovieTorrents)
			r.Get("/movies/{id}/comments", moviesHandler.GetComments)
			r.Post("/movies/{id}/comments", moviesHandler.PostComment)

			r.Get("/comments", commentsHandler.List)
			r.Get("/comments/{id}", commentsHandler.Get)
			r.Patch("/comments/{id}", commentsHandler.Update)
			r.Delete("/comments/{id}", commentsHandler.Delete)
		})
	})

	// Backward-compatible callback path for the original environment template.
	r.Get("/oauth/callback/42", authHandler.CallbackFortyTwo)
	r.Post("/oauth/token", authHandler.OAuthToken)

	return r
}

func newPasswordResetMailer() *email.BrevoMailer {
	if os.Getenv("BREVO_API_KEY") == "" {
		return nil
	}
	mailer, err := email.NewBrevoMailer(email.BrevoConfig{
		APIKey:    os.Getenv("BREVO_API_KEY"),
		FromEmail: os.Getenv("MAIL_FROM_EMAIL"),
		FromName:  os.Getenv("MAIL_FROM_NAME"),
	})
	if err != nil {
		log.Fatalf("init Brevo mailer: %v", err)
	}
	return mailer
}

func getPasswordResetTTL() time.Duration {
	rawTTL := os.Getenv("PASSWORD_RESET_TTL")
	if rawTTL == "" {
		return 30 * time.Minute
	}
	ttl, err := time.ParseDuration(rawTTL)
	if err != nil {
		log.Fatalf("parse PASSWORD_RESET_TTL: %v", err)
	}
	return ttl
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
