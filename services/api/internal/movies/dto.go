package movies

import "hypertube/api/internal/models"

type movieResponse struct {
	ImdbID      string  `json:"imdb_id"`
	Title       string  `json:"title"`
	Year        string  `json:"year"`
	PosterURL   string  `json:"poster_url"`
	BackdropURL string  `json:"backdrop_url"`
	Note        float32 `json:"note"`
	Genre       []int   `json:"genres"`
}

func toMovieResponse(m models.Movie) movieResponse {
	return movieResponse{
		ImdbID:      m.ImdbID,
		Title:       m.Title,
		Year:        m.Year,
		PosterURL:   m.PosterURL,
		BackdropURL: m.BackdropURL,
		Note:        m.Note,
		Genre:       m.Genre,
	}
}

type movieDetailResponse struct {
	ImdbID           string   `json:"imdb_id"`
	TmdbID           string   `json:"tmdb_id"`
	Title            string   `json:"title"`
	Year             string   `json:"year"`
	PosterURL        string   `json:"poster_url"`
	BackdropURL      string   `json:"backdrop_url"`
	BackdropURLExtra []string `json:"extra_backdrops"`
	Note             float32  `json:"note"`
	Genre            []int    `json:"genres"`
	Runtime          int      `json:"runtime_minutes"`
	Summary          string   `json:"summary"`
	Director         []string `json:"director"`
	Cast             []string `json:"cast"`
	Watched          bool     `json:"watched"`
	Progression      float32  `json:"progression"`
}

func toMovieDetailResponse(m models.Movie, d models.MovieDetails) movieDetailResponse {
	var defaultSummary string
	if d.Summary == "" {
		defaultSummary = m.Summary
	} else {
		defaultSummary = d.Summary
	}
	return movieDetailResponse{
		ImdbID:           m.ImdbID,
		TmdbID:           m.TmdbID,
		Title:            m.Title,
		Year:             m.Year,
		PosterURL:        m.PosterURL,
		BackdropURL:      m.BackdropURL,
		BackdropURLExtra: d.ExtraBackdrops,
		Note:             m.Note,
		Genre:            m.Genre,
		Runtime:          m.Runtime,
		Summary:          defaultSummary,
		Director:         d.Director,
		Cast:             d.Cast,
		Watched:          m.Watched,
		Progression:      m.Progression,
	}
}
