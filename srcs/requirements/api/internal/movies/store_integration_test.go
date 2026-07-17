package movies

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestStoreSaveMovieProgressInsertsAndUpdates(t *testing.T) {
	ctx, db := setupStoreIntegrationDB(t)
	store := NewStore(db)
	userID, imdbID := insertMovieProgressFixture(t, ctx, db)

	if err := store.saveMovieProgress(ctx, userID, imdbID, 120, false, 10); err != nil {
		t.Fatalf("save initial progress: %v", err)
	}
	if err := store.saveMovieProgress(ctx, userID, imdbID, 1804, true, 50); err != nil {
		t.Fatalf("save updated progress: %v", err)
	}

	var count int
	if err := db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM watch_history
		WHERE user_id = $1 AND imdbid = $2
	`, userID, imdbID).Scan(&count); err != nil {
		t.Fatalf("count watch history rows: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one watch history row, got %d", count)
	}

	var progress int
	var complete bool
	if err := db.QueryRow(ctx, `
		SELECT progress, complete
		FROM watch_history
		WHERE user_id = $1 AND imdbid = $2
	`, userID, imdbID).Scan(&progress, &complete); err != nil {
		t.Fatalf("load watch history row: %v", err)
	}
	if progress != 1804 || !complete {
		t.Fatalf("unexpected progress row: progress=%d complete=%v", progress, complete)
	}
}

func TestStoreListWatchedReturnsProgressFields(t *testing.T) {
	ctx, db := setupStoreIntegrationDB(t)
	store := NewStore(db)
	userID, imdbID := insertMovieProgressFixture(t, ctx, db)

	if err := store.saveMovieProgress(ctx, userID, imdbID, 1804, false, 50); err != nil {
		t.Fatalf("save progress: %v", err)
	}

	watched, err := store.listWatched(ctx, userID)
	if err != nil {
		t.Fatalf("list watched: %v", err)
	}
	if len(watched) != 1 {
		t.Fatalf("expected one watched movie, got %d", len(watched))
	}

	movie := watched[0]
	if movie.ImdbID != imdbID || movie.Title != "Progress Integration Movie" || movie.Year != "2026" || movie.PosterURL != "poster.jpg" || movie.BackdropURL != "backdrop.jpg" || movie.Note != 8.5 || len(movie.Genre) != 2 || movie.OriginalLanguage != "en" {
		t.Fatalf("unexpected watched movie: %+v", movie)
	}
	if movie.Progress != 1804 || movie.Complete {
		t.Fatalf("unexpected progress fields: %+v", movie)
	}
}

func TestStoreWatchHistoryUniquePerUserMovie(t *testing.T) {
	ctx, db := setupStoreIntegrationDB(t)
	store := NewStore(db)
	userID, imdbID := insertMovieProgressFixture(t, ctx, db)

	for _, progress := range []int{1, 2, 3} {
		if err := store.saveMovieProgress(ctx, userID, imdbID, progress, false, 84); err != nil {
			t.Fatalf("save progress %d: %v", progress, err)
		}
	}

	var count int
	if err := db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM watch_history
		WHERE user_id = $1 AND imdbid = $2
	`, userID, imdbID).Scan(&count); err != nil {
		t.Fatalf("count watch history rows: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one watch history row, got %d", count)
	}
}

func setupStoreIntegrationDB(t *testing.T) (context.Context, *pgxpool.Pool) {
	t.Helper()

	ctx := context.Background()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://hypertube:changeme@localhost:5432/hypertube?sslmode=disable"
	}

	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(db.Close)
	if err := db.Ping(ctx); err != nil {
		t.Skipf("test database is not reachable: %v", err)
	}
	requireWatchHistoryProgressSchema(t, ctx, db)

	return ctx, db
}

func requireWatchHistoryProgressSchema(t *testing.T, ctx context.Context, db *pgxpool.Pool) {
	t.Helper()

	var columnCount int
	if err := db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_schema = 'public'
		  AND table_name = 'watch_history'
		  AND column_name IN ('progress', 'complete')
	`).Scan(&columnCount); err != nil {
		t.Fatalf("inspect watch_history columns: %v", err)
	}
	if columnCount != 2 {
		t.Skip("watch_history progress schema is not present; reset the database from db/001_schema.sql")
	}

	var hasUniqueKey bool
	if err := db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM pg_constraint
			WHERE conname = 'watch_history_user_id_imdbid_key'
		) OR EXISTS (
			SELECT 1
			FROM pg_indexes
			WHERE schemaname = 'public'
			  AND indexname = 'watch_history_user_id_imdbid_key'
		)
	`).Scan(&hasUniqueKey); err != nil {
		t.Fatalf("inspect watch_history unique key: %v", err)
	}
	if !hasUniqueKey {
		t.Skip("watch_history unique key is not present; reset the database from db/001_schema.sql")
	}
}

func insertMovieProgressFixture(t *testing.T, ctx context.Context, db *pgxpool.Pool) (int, string) {
	t.Helper()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	var userID int
	if err := db.QueryRow(ctx, `
		INSERT INTO users (email, username, first_name, last_name, password_hash, color)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, "movie-progress-"+suffix+"@example.test", "movie_progress_"+suffix, "Test", "User", "test-hash", "green").Scan(&userID); err != nil {
		t.Fatalf("insert test user: %v", err)
	}

	imdbID := "tt-progress-" + suffix
	if _, err := db.Exec(ctx, `
		INSERT INTO movies (imdbid, tmdbid, title, year, poster_url, backdrop_url, note, genre, original_language)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, imdbID, "tmdb-"+suffix, "Progress Integration Movie", "2026", "poster.jpg", "backdrop.jpg", 8.5, []int{12, 18}, "en"); err != nil {
		t.Fatalf("insert test movie: %v", err)
	}

	t.Cleanup(func() {
		cleanupCtx := context.Background()
		if _, err := db.Exec(cleanupCtx, `DELETE FROM watch_history WHERE user_id = $1 OR imdbid = $2`, userID, imdbID); err != nil {
			t.Errorf("delete test watch history: %v", err)
		}
		if _, err := db.Exec(cleanupCtx, `DELETE FROM users WHERE id = $1`, userID); err != nil {
			t.Errorf("delete test user: %v", err)
		}
		if _, err := db.Exec(cleanupCtx, `DELETE FROM movies WHERE imdbid = $1`, imdbID); err != nil {
			t.Errorf("delete test movie: %v", err)
		}
	})

	return userID, imdbID
}
