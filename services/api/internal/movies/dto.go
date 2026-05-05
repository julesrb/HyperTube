package movies

import "hypertube/api/internal/models"

type movieResponse struct {
	ImdbID      string `json:"imdb_id"`
	Title       string `json:"title"`
	Year        string `json:"year"`
	PosterURL   string `json:"poster_url"`
	BackdropURL string `json:"backdrop_url"`
}

func toMovieResponse(m models.Movie) movieResponse {
	return movieResponse{
		ImdbID:      m.ImdbID,
		Title:       m.Title,
		Year:        m.Year,
		PosterURL:   m.PosterURL,
		BackdropURL: m.BackdropURL,
	}
}

type movieDetailResponse struct {
	ImdbID      string   `json:"imdb_id"`
	TmdbID      string   `json:"tmdb_id"`
	Title       string   `json:"title"`
	Year        string   `json:"year"`
	PosterURL   string   `json:"poster_url"`
	BackdropURL string   `json:"backdrop_url"`
	Note        float32  `json:"note"`
	Genre       []int    `json:"genres"`
	Runtime     int      `json:"runtime_minutes"`
	Summary     string   `json:"summary"`
	Director    string   `json:"director"`
	Cast        []string `json:"cast"`
	Watched     bool     `json:"watched"`
	Progression float32  `json:"progression"`
}

func toMovieDetailResponse(m models.Movie) movieDetailResponse {
	return movieDetailResponse{
		ImdbID:      m.ImdbID,
		TmdbID:      m.TmdbID,
		Title:       m.Title,
		Year:        m.Year,
		PosterURL:   m.PosterURL,
		BackdropURL: m.BackdropURL,
		Note:        m.Note,
		Genre:       m.Genre,
		Runtime:     m.Runtime,
		Summary:     m.Summary,
		Director:    m.Director,
		Cast:        m.Cast,
		Watched:     m.Watched,
		Progression: m.Progression,
	}
}
