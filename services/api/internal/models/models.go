package models

type Movie struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Year        int      `json:"year,omitempty"`
	PosterURL   string   `json:"poster_url,omitempty"`
	BackdropURL string   `json:"backdrop_url,omitempty"`
	IMDbRating  float32  `json:"imdb_rating,omitempty"`
	Genres      []string `json:"genres,omitempty"`
	Runtime     int      `json:"runtime_minutes,omitempty"`
	Summary     string   `json:"summary,omitempty"`
	Director    string   `json:"director,omitempty"`
	Cast        []string `json:"cast,omitempty"`
	Watched     bool     `json:"watched"`
	Seeders     int      `json:"seeders,omitempty"`
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
