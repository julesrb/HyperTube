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

func (s *Store) listFeatured(ctx context.Context) ([]models.Movie, error) {
	rows, err := s.db.Query(ctx, `
		SELECT m.imdbid, m.tmdbid, m.title, m.year,
		       m.poster_url, m.backdrop_url,
		       m.imdb_rating, m.genres,
		       m.runtime_minutes, m.summary,
		       m.director, m."cast",
		       m.watched, m.progression, m.seeders
		FROM   movies m
		JOIN   featured_movies f ON f.movie_id = m.imdbid
		ORDER  BY f.position
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []models.Movie
	for rows.Next() {
		var m models.Movie
		if err := rows.Scan(
			&m.ImdbID, &m.TmdbID, &m.Title, &m.Year, &m.PosterURL, &m.BackdropURL,
			&m.IMDbRating, &m.Genres, &m.Runtime, &m.Summary,
			&m.Director, &m.Cast, &m.Watched, &m.Progression, &m.Seeders,
		); err != nil {
			return nil, err
		}
		movies = append(movies, m)
	}
	return movies, rows.Err()
}

func (s *Store) findByID(ctx context.Context, id string) (*models.Movie, error) {
	row := s.db.QueryRow(ctx, `
		SELECT imdbid, tmdbid, title, year,
		       poster_url, backdrop_url,
		       imdb_rating, genres,
		       runtime_minutes, summary,
		       director, "cast",
		       watched, progression, seeders
		FROM movies WHERE imdbid = $1`, id)

	var m models.Movie
	if err := row.Scan(
		&m.ImdbID, &m.TmdbID, &m.Title, &m.Year, &m.PosterURL, &m.BackdropURL,
		&m.IMDbRating, &m.Genres, &m.Runtime, &m.Summary,
		&m.Director, &m.Cast, &m.Watched, &m.Progression, &m.Seeders,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &m, nil
}
