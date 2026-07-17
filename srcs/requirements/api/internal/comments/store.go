package comments

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"hypertube/api/internal/models"
)

var ErrNotFound = errors.New("not found")

type Store struct {
	db *pgxpool.Pool
}

type commentWithMovieRow struct {
	ID               int       `db:"id"`
	UserID           int       `db:"user_id"`
	MovieID          string    `db:"movie_id"`
	Content          string    `db:"content"`
	Edited           bool      `db:"edited"`
	UpdatedAt        time.Time `db:"updated_at"`
	MovieImdbID      string    `db:"movie_imdb_id"`
	MovieTitle       string    `db:"movie_title"`
	MovieYear        string    `db:"movie_year"`
	MovieBackdropURL string    `db:"movie_backdrop_url"`
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

func (s *Store) create(ctx context.Context, content string, movieID string, userID int) (models.Comment, error) {
	rows, err := s.db.Query(ctx, `
		INSERT INTO comments (user_id, movie_id, content, updated_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING *
	`, userID, movieID, content)
	if err != nil {
		return models.Comment{}, err
	}

	return pgx.CollectOneRow(rows, pgx.RowToStructByName[models.Comment])
}

func (s *Store) findByID(ctx context.Context, id int) (*models.Comment, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, user_id, movie_id, content, edited, updated_at
		FROM comments
		WHERE id = $1
	`, id)
	if err != nil {
		return nil, err
	}

	comment, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[models.Comment])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &comment, nil
}

func (s *Store) findAll(ctx context.Context, limit, offset int) ([]models.Comment, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, user_id, movie_id, content, edited, updated_at
		FROM comments
		ORDER BY updated_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}

	comments, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Comment])
	if err != nil {
		return nil, err
	}

	return comments, nil
}

func (s *Store) findAllByUserID(ctx context.Context, userID int64, limit, offset int) ([]models.CommentWithMovie, error) {
	rows, err := s.db.Query(ctx, `
		SELECT
			c.id,
			c.user_id,
			c.movie_id,
			c.content,
			c.edited,
			c.updated_at,
			m.imdbid AS movie_imdb_id,
			m.title AS movie_title,
			m.year AS movie_year,
			m.backdrop_url AS movie_backdrop_url
		FROM comments c
		JOIN movies m ON m.imdbid = c.movie_id
		WHERE c.user_id = $1
		ORDER BY c.updated_at DESC, c.id DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	dbRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[commentWithMovieRow])
	if err != nil {
		return nil, err
	}

	comments := make([]models.CommentWithMovie, len(dbRows))
	for i, row := range dbRows {
		comments[i] = models.CommentWithMovie{
			ID:        row.ID,
			UserID:    row.UserID,
			MovieID:   row.MovieID,
			Content:   row.Content,
			Edited:    row.Edited,
			UpdatedAt: row.UpdatedAt,
			Movie: models.MovieSmall{
				ImdbID:      row.MovieImdbID,
				Title:       row.MovieTitle,
				Year:        row.MovieYear,
				BackdropURL: row.MovieBackdropURL,
			},
		}
	}

	return comments, nil
}

func (s *Store) countAll(ctx context.Context) (int, error) {
	var total int
	err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM comments`).Scan(&total)
	return total, err
}

func (s *Store) countAllByUserID(ctx context.Context, userID int64) (int, error) {
	var total int
	err := s.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM comments
		WHERE user_id = $1
	`, userID).Scan(&total)
	return total, err
}

func (s *Store) update(ctx context.Context, content string, id int, userID int) (models.Comment, error) {
	rows, err := s.db.Query(ctx, `
		UPDATE comments
		SET content = $1, updated_at = NOW(), edited = TRUE
		WHERE id = $2 AND user_id = $3
		RETURNING *
	`, content, id, userID)
	if err != nil {
		return models.Comment{}, err
	}

	comment, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[models.Comment])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Comment{}, ErrNotFound
		}
		return models.Comment{}, err
	}

	return comment, nil
}

func (s *Store) delete(ctx context.Context, id int, userID int) error {
	tag, err := s.db.Exec(ctx, `
		DELETE FROM comments
		WHERE id = $1 AND user_id = $2
	`, id, userID)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
