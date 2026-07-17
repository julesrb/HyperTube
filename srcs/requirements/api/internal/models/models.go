package models

import "time"

type Movie struct {
	ImdbID           string  `json:"imdb_id"`
	TmdbID           string  `json:"tmdb_id"`
	Title            string  `json:"title"`
	Year             string  `json:"year"`
	PosterURL        string  `json:"poster_url"`
	BackdropURL      string  `json:"backdrop_url"`
	Genre            []int   `json:"genres"`
	Note             float32 `json:"note"`
	OriginalLanguage string  `json:"original_language"`
}

type WatchedMovie struct {
	ImdbID           string  `json:"imdb_id" db:"imdbid"`
	TmdbID           string  `json:"tmdb_id" db:"tmdbid"`
	Title            string  `json:"title" db:"title"`
	Year             string  `json:"year" db:"year"`
	PosterURL        string  `json:"poster_url" db:"poster_url"`
	BackdropURL      string  `json:"backdrop_url" db:"backdrop_url"`
	Genre            []int   `json:"genres" db:"genre"`
	Note             float32 `json:"note" db:"note"`
	OriginalLanguage string  `json:"original_language" db:"original_language"`
	Progress         int     `json:"progress" db:"progress"`
	Complete         bool    `json:"complete" db:"complete"`
	Pourcent         int     `json:"pourcent" db:"pourcent"`
}

type Torrent struct {
	Id         string     `json:"id"`
	ImdbID     string     `json:"imdb_id"`
	Title      string     `json:"title"`
	Year       int        `json:"year"`
	Source     string     `json:"source"`
	URL        string     `json:"url"`
	Hash       string     `json:"hash"`
	Quality    string     `json:"quality"`
	Size       float64    `json:"size"`
	Language   string     `json:"language"`
	Seeds      string     `json:"seeds"`
	Status     string     `json:"status"`
	FinishedAt *time.Time `json:"finished_at,omitempty" db:"finished_at"`
}

type Subtitle struct {
	URL      string `json:"url"`
	Language string `json:"language"`
}

type MovieDetails struct {
	Summary        string   `json:"summary"`
	Runtime        int      `json:"runtime"`
	Director       []string `json:"director"`
	Cast           []string `json:"cast"`
	ExtraBackdrops []string `json:"extra_backdrops"`
}

type Comment struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	MovieID   string    `json:"movie_id"`
	Content   string    `json:"content"`
	Edited    bool      `json:"edited" db:"edited"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type MovieSmall struct {
	ImdbID      string `json:"imdb_id"`
	Title       string `json:"title"`
	Year        string `json:"year"`
	BackdropURL string `json:"backdrop_url"`
}

type CommentWithMovie struct {
	ID        int        `json:"id"`
	UserID    int        `json:"user_id"`
	MovieID   string     `json:"movie_id"`
	Movie     MovieSmall `json:"movie"`
	Content   string     `json:"content"`
	Edited    bool       `json:"edited"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type CommentWithUser struct {
	ID        int       `json:"id"`
	User      UserSmall `json:"user"`
	Content   string    `json:"content"`
	Edited    bool      `json:"edited"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type MovieSearchRow struct {
	Query      string    `json:"query"      db:"query"`
	ImdbID     string    `json:"imdb_id"    db:"imdbid"`
	Rank       int       `json:"rank"       db:"rank"`
	SearchedAt time.Time `json:"searched_at" db:"searched_at"`
}

type OAuthApplication struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"owner_id" db:"owner_user_id"`
	Name        string    `json:"name"`
	RedirectURI string    `json:"redirect_uri" db:"redirect_uri"`
	ClientID    string    `json:"client_id" db:"client_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
