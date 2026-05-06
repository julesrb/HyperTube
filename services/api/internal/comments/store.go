package comments

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"hypertube/api/internal/models"
)

var ErrNotFound = errors.New("not found")

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

func (s *Store) findByID(ctx context.Context, id string) (*models.Comment, error) {
	rows, err := s.db.Query(ctx, `
        SELECT id, user_id, movie_id, content, updated_at
        FROM comments 
        WHERE id = $1`, id)
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

func (s *Store) findAll(ctx context.Context) ([]models.Comment, error) {
	rows, err := s.db.Query(ctx, `
        SELECT id, user_id, movie_id, content, updated_at
        FROM comments 
        ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}

	comments, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Comment])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return comments, nil
}
