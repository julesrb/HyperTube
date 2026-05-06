package models

import "time"

type Movie struct {
	ImdbID      string   `json:"imdb_id"`
	TmdbID      string   `json:"tmdb_id"`
	Title       string   `json:"title"`
	Year        string   `json:"year"`
	PosterURL   string   `json:"poster_url"`
	BackdropURL string   `json:"backdrop_url"`
	Genre       []int    `json:"genres"`
	Runtime     int      `json:"runtime_minutes"`
	Note        float32  `json:"note"`
	Summary     string   `json:"summary"`
	Director    string   `json:"director"`
	Cast        []string `json:"cast"`
	Watched     bool     `json:"watched"`
	Progression float32  `json:"progression"`
}

type Torrent struct {
	Id       int     `json:"id"`
	ImdbID   string  `json:"imdb_id"`
	Title    string  `json:"title"`
	Year     int     `json:"year"`
	Source   string  `json:"source"`
	URL      string  `json:"url"`
	Quality  string  `json:"quality"`
	Size     float64 `json:"size"`
	Language string  `json:"language"`
	Seeds    string  `json:"seeds"`
}

type Subtitle struct {
	URL      string `json:"url"`
	Language string `json:"language"`
}

type MovieDetails struct {
	Summary  string   `json:"summary"`
	Director string   `json:"director"`
	Cast     []string `json:"cast"`
}

type Comment struct {
	ID         int    `json:"id"`
	UserID     int    `json:"user_id"`
	MovieID    string `json:"movie_id"`
	Content    string `json:"content"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
