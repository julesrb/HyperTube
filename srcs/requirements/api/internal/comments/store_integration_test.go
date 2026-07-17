package comments

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestIntegrationFindAllByUserIDIncludesMovie(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(db.Close)
	if err := db.Ping(ctx); err != nil {
		t.Skipf("test database is not reachable: %v", err)
	}

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	var userID int64
	err = db.QueryRow(ctx, `
		INSERT INTO users (email, username, first_name, last_name, password_hash, color)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, "comment-test-"+suffix+"@example.test", "comment_test_"+suffix, "Test", "User", "test-hash", "green").Scan(&userID)
	if err != nil {
		t.Fatalf("insert test user: %v", err)
	}
	t.Cleanup(func() {
		if _, err := db.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userID); err != nil {
			t.Errorf("delete test user: %v", err)
		}
	})

	imdbID := "tt-comment-" + suffix
	_, err = db.Exec(ctx, `
		INSERT INTO movies (imdbid, tmdbid, title, year, poster_url, backdrop_url, note, genre)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, imdbID, "tmdb-"+suffix, "Integration Movie", "2026", "poster.jpg", "backdrop.jpg", 8.5, []int{12, 18})
	if err != nil {
		t.Fatalf("insert test movie: %v", err)
	}
	t.Cleanup(func() {
		if _, err := db.Exec(context.Background(), `DELETE FROM movies WHERE imdbid = $1`, imdbID); err != nil {
			t.Errorf("delete test movie: %v", err)
		}
	})

	var commentID int
	err = db.QueryRow(ctx, `
		INSERT INTO comments (user_id, movie_id, content, edited)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, userID, imdbID, "Integration comment", true).Scan(&commentID)
	if err != nil {
		t.Fatalf("insert test comment: %v", err)
	}
	t.Cleanup(func() {
		if _, err := db.Exec(context.Background(), `DELETE FROM comments WHERE id = $1`, commentID); err != nil {
			t.Errorf("delete test comment: %v", err)
		}
	})

	comments, err := NewStore(db).findAllByUserID(ctx, userID, 12, 0)
	if err != nil {
		t.Fatalf("find comments by user: %v", err)
	}
	if len(comments) != 1 {
		t.Fatalf("expected one comment, got %d", len(comments))
	}

	comment := comments[0]
	if comment.ID != commentID || comment.UserID != int(userID) || comment.MovieID != imdbID || comment.Content != "Integration comment" || !comment.Edited || comment.UpdatedAt.IsZero() {
		t.Fatalf("unexpected comment: %+v", comment)
	}
	if comment.Movie.ImdbID != imdbID || comment.Movie.Title != "Integration Movie" || comment.Movie.Year != "2026" || comment.Movie.BackdropURL != "backdrop.jpg" {
		t.Fatalf("unexpected movie: %+v", comment.Movie)
	}
}
