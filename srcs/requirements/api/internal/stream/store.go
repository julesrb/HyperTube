package stream

import (
	"context"
	"errors"

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

// Torrent holds the torrent metadata needed to start a stream.
type Torrent struct {
	URL              string
	Title            string
	OriginalLanguage string
}

func (s *Store) GetTorrent(ctx context.Context, id string) (Torrent, error) {
	var t Torrent
	err := s.db.QueryRow(ctx, `
		SELECT t.url, t.title, m.original_language
		FROM torrents t
		JOIN movies m ON m.imdbid = t.imdbid
		WHERE t.id = $1`, id).Scan(&t.URL, &t.Title, &t.OriginalLanguage)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Torrent{}, ErrNotFound
		}
		return Torrent{}, err
	}
	return t, nil
}

func (s *Store) GetTorrentStatus(ctx context.Context, id string) (string, error) {
	var status string
	err := s.db.QueryRow(ctx, `SELECT status FROM torrents WHERE id = $1`, id).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	return status, nil
}

func (s *Store) SetTorrentStatus(ctx context.Context, id, status string) error {
	tag, err := s.db.Exec(ctx, `UPDATE torrents SET status = $2 WHERE id = $1`, id, status)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
