package movies

import "hypertube/api/internal/models"

type movieResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	PosterURL   string `json:"poster_url"`
	BackdropURL string `json:"backdrop_url"`
}

func toMovieResponse(m models.Movie) movieResponse {
	return movieResponse{
		ID:          m.ID,
		Title:       m.Title,
		PosterURL:   m.PosterURL,
		BackdropURL: m.BackdropURL,
	}
}
