package main

import (
	"context"
	_ "embed"
	"log"
	"net/http"
	"os"
	"time"

	"hypertube/api/internal/auth"
	"hypertube/api/internal/comments"
	"hypertube/api/internal/email"
	"hypertube/api/internal/movies"
	"hypertube/api/internal/movies/c411"
	"hypertube/api/internal/movies/tmdb"
	"hypertube/api/internal/stream"
	"hypertube/api/internal/users"

	"hypertube/api/internal/movies/archive.org"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
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
	githubRedirectURL := getEnv("GITHUB_REDIRECT_URL", "http://localhost:8080/api/v1/auth/github/callback")
	gitlabRedirectURL := getEnv("GITLAB_REDIRECT_URL", "http://localhost:8080/api/v1/auth/gitlab/callback")
	authOptions := []auth.HandlerOption{
		auth.WithFrontendAuthCallbackURL(getEnv("FRONTEND_AUTH_CALLBACK_URL", "http://localhost:4200/en/auth/callback")),
		auth.WithPasswordResetURL(getEnv("PASSWORD_RESET_URL", "http://localhost:4200/{locale}/password-reset/set-new-password")),
		auth.WithPasswordResetTTL(getPasswordResetTTL()),
		auth.WithFortyTwoOAuth(auth.NewFortyTwoOAuth(auth.FortyTwoOAuthConfig{
			ClientID:     os.Getenv("FORTYTWO_CLIENT_ID"),
			ClientSecret: os.Getenv("FORTYTWO_CLIENT_SECRET"),
			RedirectURL:  fortyTwoRedirectURL,
		})),
		auth.WithGitHubOAuth(auth.NewGitHubOAuth(auth.GitHubOAuthConfig{
			ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
			ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
			RedirectURL:  githubRedirectURL,
		})),
		auth.WithGitLabOAuth(auth.NewGitLabOAuth(auth.GitLabOAuthConfig{
			ClientID:     os.Getenv("GITLAB_CLIENT_ID"),
			ClientSecret: os.Getenv("GITLAB_CLIENT_SECRET"),
			RedirectURL:  gitlabRedirectURL,
		})),
	}
	if passwordResetMailer := newPasswordResetMailer(); passwordResetMailer != nil {
		authOptions = append(authOptions, auth.WithPasswordResetMailer(passwordResetMailer))
	}
	authHandler := auth.NewHandler(authStore, tokenManager, authOptions...)

	movieStore := movies.NewStore(db)
	commentStore := comments.NewStore(db)
	userStore := users.NewStore(db)
	streamStore := stream.NewStore(db)

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
	moviesHandler := movies.NewMoviesHandler(movieStore, userStore, searchers, tmdbClient)
	go moviesHandler.SeedFeaturedTorrents(ctx)
	commentsHandler := comments.NewCommentsHandler(commentStore)
	usersHandler := users.NewHandler(userStore)
	streamHandler := stream.NewStreamHandler(streamStore)

	addr := ":" + getEnv("PORT", "8080")
	log.Printf("api listening on %s", addr)
	allowedOrigin := getEnv("CORS_ALLOWED_ORIGIN", "http://localhost:4200")
	log.Fatal(http.ListenAndServe(addr, newRouter(moviesHandler, commentsHandler, authHandler, usersHandler, tokenManager, authStore, streamHandler, allowedOrigin)))

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
	// log.Printf("startup: top %d movies by seeds:", len(featured))
	// for _, t := range featured {
	// 	log.Printf("  imdb=%s seeds=%s title=%s", t.ImdbID, t.Seeds, t.Title)
	// }
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
		if err = store.UpsertDefault(ctx, torrent.ImdbID, i); err != nil {
			log.Printf("startup: failed to store default torrent %s: %v", torrent.Title, err)
		}
	}
}

func newRouter(
	moviesHandler *movies.MoviesHandler,
	commentsHandler *comments.CommentsHandler,
	authHandler *auth.Handler,
	usersHandler *users.Handler,
	tokenManager *auth.TokenManager,
	authUsers auth.UserExistenceChecker,
	streamHandler *stream.StreamHandler,
	allowedOrigin string,
) chi.Router {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{allowedOrigin},
		AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization", "Accept-Language"},
		MaxAge:         600,
	}))

	requireAuth := auth.RequireAuth(tokenManager, authUsers)
	// requireAuth = auth.DevAuthenticateAs(1)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh-token", authHandler.RefreshToken)
			r.Post("/password-reset/send-email", authHandler.RequestPasswordReset)
			r.Post("/password-reset/set-new-password", authHandler.ResetPassword)
			r.Get("/42/login", authHandler.LoginFortyTwo)
			r.Get("/42/callback", authHandler.CallbackFortyTwo)
			r.Get("/github/login", authHandler.LoginGitHub)
			r.Get("/github/callback", authHandler.CallbackGitHub)
			r.Get("/gitlab/login", authHandler.LoginGitLab)
			r.Get("/gitlab/callback", authHandler.CallbackGitLab)
		})

		r.Post("/oauth/token", authHandler.OAuthToken)
		r.With(requireAuth).Post("/oauth/applications", authHandler.CreateOAuthApplication)
		r.With(requireAuth).Get("/oauth/applications", authHandler.ListOAuthApplications)
		r.With(requireAuth).Patch("/oauth/applications/{id}", authHandler.UpdateOAuthApplication)
		r.With(requireAuth).Delete("/oauth/applications/{id}", authHandler.DeleteOAuthApplication)

		r.Get("/movies", moviesHandler.GetDefaultMovies)
		r.Get("/movies/featured", moviesHandler.GetFeaturedMovies)
		r.Get("/movies/search", moviesHandler.SearchMovies)

		r.With(requireAuth).Get("/movies/watched", moviesHandler.GetWatchedMovies)
		r.With(requireAuth).Get("/movies/directstream", moviesHandler.GetDirectStreamMovies)
		r.With(requireAuth).Patch("/movies/{id}/progress", moviesHandler.PatchMovieProgress)
		r.With(requireAuth).Get("/movies/{id}", moviesHandler.GetMoviesId)
		r.With(requireAuth).Get("/movies/{id}/torrents", moviesHandler.GetMovieTorrents)
		r.With(requireAuth).Get("/movies/{id}/comments", moviesHandler.GetComments)
		r.With(requireAuth).Post("/movies/{id}/comments", moviesHandler.PostComment)

		r.With(requireAuth).Get("/comments", commentsHandler.List)
		r.With(requireAuth).Get("/comments/{id}", commentsHandler.Get)
		r.With(requireAuth).Patch("/comments/{id}", commentsHandler.Update)
		r.With(requireAuth).Delete("/comments/{id}", commentsHandler.Delete)

		r.With(requireAuth).Get("/users", usersHandler.ListUsers)
		r.With(requireAuth).Get("/users/{id}", usersHandler.GetUser)
		r.With(requireAuth).Get("/users/{id}/comments", commentsHandler.ListByUser)
		r.With(requireAuth).Get("/users/{id}/movie-history", moviesHandler.GetUserFilmHistory)
		r.With(requireAuth).Patch("/users/new-password", usersHandler.SetPassword)
		r.With(requireAuth).Patch("/users/{id}", usersHandler.UpdateUser)

		r.With(requireAuth).Get("/stream/{id}", streamHandler.InitStream)           // start torrent and prepapre for trancoding and streaming
		r.With(requireAuth).Get("/stream/{id}/index", streamHandler.GetIndex)       // serve the HLS index
		r.With(requireAuth).Get("/stream/{id}/{segment}", streamHandler.GetSegment) // serve the HLS segments
	})

	// Backward-compatible callback path for the original environment template.
	r.Get("/oauth/callback/42", authHandler.CallbackFortyTwo)
	r.Get("/oauth/callback/github", authHandler.CallbackGitHub)
	r.Get("/oauth/callback/gitlab", authHandler.CallbackGitLab)

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
