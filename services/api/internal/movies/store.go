package movies

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"hypertube/api/internal/models"
)

//go:embed movies.json
var listData []byte

func loadMovies() ([]models.Movie, error) {
	var movies []models.Movie
	if err := json.Unmarshal(listData, &movies); err != nil {
		return nil, fmt.Errorf("parse list.json: %w", err)
	}
	return movies, nil
}
