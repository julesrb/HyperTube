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
	ImdbID      string   `db:"imdbid"`
	TmdbID      string   `db:"tmdbid"`
	Title       string   `db:"title"`
	Year        string   `db:"year"`
	Note        float32  `db:"note"`
	PosterURL   string   `db:"poster_url"`
	BackdropURL string   `db:"backdrop_url"`
	Genre       []int    `db:"genre"`
	Runtime     int      `db:"runtime_minutes"`
	Summary     string   `db:"summary"`
	Director    string   `db:"director"`
	Cast        []string `db:"cast"`
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
		Director:    r.Director,
		Cast:        r.Cast,
	}
}

func (s *Store) listFeatured(ctx context.Context) ([]models.Movie, error) {
	rows, err := s.db.Query(ctx, `
		SELECT m.imdbid, m.tmdbid, m.title, m.year,
		       m.poster_url, m.backdrop_url, m.note, m.genre,
		       m.runtime_minutes, m.summary, m.director, m."cast"
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
		       runtime_minutes, summary, director, "cast"
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
