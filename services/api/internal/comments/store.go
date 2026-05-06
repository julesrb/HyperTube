package comments

import (
	"context"
	"errors"
	"fmt"
	"hypertube/api/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

func (s *Store) update(ctx context.Context, content string, id string, user_id int) (models.Comment, error) {
	rows, err := s.db.Query(ctx, `
		UPDATE comments
		SET content = $1, updated_at = NOW()
		WHERE id = $2 AND user_id = $3
		RETURNING *
	`, content, id, user_id)

	if err != nil {
		return models.Comment{}, err
	}

	r, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[models.Comment])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Comment{}, ErrNotFound
		}
		return models.Comment{}, err
	}
	return r, nil
}

func (s *Store) delete(ctx context.Context, id string, user_id int) error {
	tag, err := s.db.Exec(ctx, `
		DELETE FROM comments
		WHERE id = $1 AND user_id = $2
	`, id, user_id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("comment not found")
	}
	return nil
}
