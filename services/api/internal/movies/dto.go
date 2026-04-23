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
