package movies

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

type movieRow struct {
	ImdbID      string  `db:"imdbid"`
	TmdbID      string  `db:"tmdbid"`
	Title       string  `db:"title"`
	Year        string  `db:"year"`
	Note        float32 `db:"note"`
	PosterURL   string  `db:"poster_url"`
	BackdropURL string  `db:"backdrop_url"`
	Genre       []int   `db:"genre"`
	Runtime     int     `db:"runtime_minutes"`
	Summary     string  `db:"summary"`
}

func toMovie(r movieRow) models.Movie {
	return models.Movie{
		ImdbID:      r.ImdbID,
		TmdbID:      r.TmdbID,
		Title:       r.Title,
		Year:        r.Year,
		PosterURL:   r.PosterURL,
		BackdropURL: r.BackdropURL,
		Note:        r.Note,
		Genre:       r.Genre,
		Runtime:     r.Runtime,
		Summary:     r.Summary,
	}
}

func (s *Store) listFeatured(ctx context.Context) ([]models.Movie, error) {
	rows, err := s.db.Query(ctx, `
		SELECT m.imdbid, m.tmdbid, m.title, m.year,
		       m.poster_url, m.backdrop_url, m.note, m.genre,
		       m.runtime_minutes, m.summary
		FROM movies m
		JOIN featured_movies f ON f.imdbid = m.imdbid
		ORDER BY f.position
	`)
	if err != nil {
		return nil, err
	}

	movieRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[movieRow])
	if err != nil {
		return nil, err
	}

	movies := make([]models.Movie, len(movieRows))
	for i, r := range movieRows {
		movies[i] = toMovie(r)
	}
	return movies, nil
}

func (s *Store) findByID(ctx context.Context, id string) (*models.Movie, error) {
	rows, err := s.db.Query(ctx, `
		SELECT imdbid, tmdbid, title, year,
		       poster_url, backdrop_url, note, genre,
		       runtime_minutes, summary
		FROM movies WHERE imdbid = $1`, id)
	if err != nil {
		return nil, err
	}

	r, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[movieRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	m := toMovie(r)
	return &m, nil
}

func (s *Store) UpsertMovie(ctx context.Context, m models.Movie) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO movies (imdbid, tmdbid, title, year, poster_url, backdrop_url, note, genre, runtime_minutes, summary)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (imdbid) DO NOTHING
	`, m.ImdbID, m.TmdbID, m.Title, m.Year, m.PosterURL, m.BackdropURL, m.Note, m.Genre, m.Runtime, m.Summary)
	return err
}

func (s *Store) findTorrent(ctx context.Context, imdbID string) ([]models.Torrent, error) {
	rows, err := s.db.Query(ctx, `
		SELECT * FROM torrents WHERE imdbid = $1
	`, imdbID)
	if err != nil {
		return nil, err
	}
	torrents, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Torrent])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return torrents, nil
}

func (s *Store) UpsertTorrent(ctx context.Context, ts models.Torrent) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO torrents (imdbid, source, year, title, url, quality, size, language, seeds)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (imdbid, url) DO NOTHING
	`, ts.ImdbID, ts.Source, ts.Year, ts.Title, ts.URL, ts.Quality, ts.Size, ts.Language, ts.Seeds)
	return err
}

func (s *Store) UpsertFeatured(ctx context.Context, imdbId string, position int) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO featured_movies (imdbid, position)
		VALUES ($1, $2)
		ON CONFLICT (imdbid, position) DO NOTHING
	`, imdbId, position)
	return err
}

func (s *Store) listComments(ctx context.Context, imdbId string) ([]models.Comment, error) {
	rows, err := s.db.Query(ctx, `
		SELECT * FROM comments
		WHERE movie_id = $1
		ORDER BY updated_at DESC
	`, imdbId)
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

func (s *Store) createComment(ctx context.Context, c models.Comment) (models.Comment, error) {
	rows, err := s.db.Query(ctx, `
		INSERT INTO comments (user_id, movie_id, content)
		VALUES ($1, $2, $3)
		RETURNING *
	`, c.UserID, c.MovieID, c.Content)

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

func (s *Store) countSearchResults(ctx context.Context, query string) (int, error) {
	var count int
	err := s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM movie_searches WHERE query = $1
	`, query).Scan(&count)
	return count, err
}

func (s *Store) upsertSearchResults(ctx context.Context, query string, imdbIDs []string) error {
	for i, id := range imdbIDs {
		_, err := s.db.Exec(ctx, `
			INSERT INTO movie_searches (query, imdbid, rank)
			VALUES ($1, $2, $3)
			ON CONFLICT (query, imdbid) DO UPDATE SET rank = $3, searched_at = NOW()
		`, query, id, i)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) listSearchResults(ctx context.Context, query string, limit, offset int) ([]models.Movie, error) {
	rows, err := s.db.Query(ctx, `
		SELECT m.imdbid, m.tmdbid, m.title, m.year,
		       m.poster_url, m.backdrop_url, m.note, m.genre,
		       m.runtime_minutes, m.summary
		FROM movies m
		JOIN movie_searches ms ON ms.imdbid = m.imdbid
		WHERE ms.query = $1
		ORDER BY ms.rank
		LIMIT $2 OFFSET $3
	`, query, limit, offset)
	if err != nil {
		return nil, err
	}

	movieRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[movieRow])
	if err != nil {
		return nil, err
	}

	movies := make([]models.Movie, len(movieRows))
	for i, r := range movieRows {
		movies[i] = toMovie(r)
	}
	return movies, nil
}
