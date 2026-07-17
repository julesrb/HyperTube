package torrent_transcode

import (
	"context"
	"errors"
	"time"

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

type Torrent struct {
	URL   string
	Title string
}

func (s *Store) GetTorrent(ctx context.Context, id string) (Torrent, error) {
	var t Torrent
	err := s.db.QueryRow(ctx, `SELECT url, title FROM torrents WHERE id = $1`, id).Scan(&t.URL, &t.Title)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Torrent{}, ErrNotFound
		}
		return Torrent{}, err
	}
	return t, nil
}

func (s *Store) SetTorrentStatus(ctx context.Context, id, status string) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE torrents
		 SET status = $2,
		     finished_at = CASE WHEN $2 = 'finished' THEN NOW() ELSE NULL END
		 WHERE id = $1`, id, status)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) FinishedBefore(ctx context.Context, cutoff time.Time) ([]string, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id FROM torrents
		 WHERE status = 'finished' AND finished_at IS NOT NULL AND finished_at < $1`, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (s *Store) InProgress(ctx context.Context) ([]string, error) {
	rows, err := s.db.Query(ctx, `SELECT id FROM torrents WHERE status = 'in_progress'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (s *Store) ResetTorrent(ctx context.Context, id string) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE torrents SET status = 'zero', finished_at = NULL WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
