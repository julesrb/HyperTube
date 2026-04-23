package models

type Movie struct {
	ImdbID      string   `json:"imdb_id"`
	TmdbID      string   `json:"tmdb_id"`
	Title       string   `json:"title"`
	Year        string   `json:"year"`
	PosterURL   string   `json:"poster_url"`
	BackdropURL string   `json:"backdrop_url"`
	IMDbRating  float32  `json:"imdb_rating"`
	Genres      []string `json:"genres"`
	Runtime     int      `json:"runtime_minutes"`
	Summary     string   `json:"summary"`
	Director    string   `json:"director"`
	Cast        []string `json:"cast"`
	Watched     bool     `json:"watched"`
	Progression float32  `json:"progression"`
	Seeders     int      `json:"seeders"`
}

type Meta struct {
	Total   int `json:"total"`
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

type Response struct {
	Data any   `json:"data"`
	Meta *Meta `json:"meta,omitempty"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type Torrent struct {
	URL      string `json:"url"`
	Quality  string `json:"quality"`
	Size     string `json:"size"`
	Language string `json:"language"`
	Seeds    string `json:"seeds"`
	Source   string `json:"source"`
}

type MovieTorrents struct {
	ImdbID   string    `json:"imdb_id"`
	Torrents []Torrent `json:"torrent"`
}
