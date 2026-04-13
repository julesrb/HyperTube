package movies

import (
	"net/http"

	"hypertube/api/internal/respond"
)

// GetMovies returns a list of movies.
func GetMovies(w http.ResponseWriter, r *http.Request) {
	movies, err := loadMovies()
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load movies")
		return
	}

	response := make([]movieResponse, len(movies))
	for i, m := range movies {
		response[i] = toMovieResponse(m)
	}
	respond.List(w, http.StatusOK, response, len(response))
}

// Get returns metadata for a single movie.
func Get(w http.ResponseWriter, r *http.Request) {}

// ListComments returns comments for a movie.
func ListComments(w http.ResponseWriter, r *http.Request) {}

// CreateComment posts a new comment on a movie.
func CreateComment(w http.ResponseWriter, r *http.Request) {}
