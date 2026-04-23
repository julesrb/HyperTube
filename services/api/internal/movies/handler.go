package movies

import (
	"context"
	"errors"
	"log"
	"net/http"

	"hypertube/api/internal/models"
	"hypertube/api/internal/movies/tmdb"
	"hypertube/api/internal/respond"
)

type movieStore interface {
	listFeatured(ctx context.Context) ([]models.Movie, error)
	findByID(ctx context.Context, id string) (*models.Movie, error)
}

type MovieSearcher interface {
	SearchByTitle(ctx context.Context, title string) ([]models.MovieTorrents, error)
}

type Handler struct {
	store     movieStore
	searchers []MovieSearcher
	tmdb      *tmdb.Client
}

func NewHandler(store movieStore, searchers []MovieSearcher) *Handler {
	return &Handler{store: store, searchers: searchers, tmdb: tmdb.NewClient()}
}

// GetMovies returns a list of movies.
func (h *Handler) GetMovies(w http.ResponseWriter, r *http.Request) {
	movies, err := h.store.listFeatured(r.Context())
	if err != nil {
		log.Println("db err:", err)
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
func (h *Handler) GetMoviesId(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	movie, err := h.store.findByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			respond.Error(w, http.StatusNotFound, "NOT_FOUND", "movie not found")
		} else {
			log.Println("db err:", err)
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load movie")
		}
		return
	}
	respond.Item(w, http.StatusOK, movie)
}

func (h *Handler) SearchMovies(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	log.Printf("searching for movies with title: %s", title)
	if title == "" {
		respond.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "title query parameter is required")
		return
	}
	moviesAndTorrents, err := h.searchers[0].SearchByTitle(r.Context(), title) // TODO add the second torrent source
	if err != nil {
		log.Println("search err:", err)
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to search movies")
		return
	}
	moviesAndTorrents = moviesAndTorrents[:11] // Limit to first 10 results MANUAL SAFE LIMIT FOR TMDB 
	movies := make([]tmdb.MovieResult, 0)
	for _, movieTorrents := range moviesAndTorrents {
		movie, err := h.tmdb.FindByIMDBID(r.Context(), movieTorrents.ImdbID)
		if err != nil {
			log.Printf("TMDB lookup error for IMDb ID %s: %v", movieTorrents.ImdbID, err)
			continue
		}
		movies = append(movies, movie)
	}
	respond.List(w, http.StatusOK, movies, len(movies))
}

// ListComments returns comments for a movie.
func (h *Handler) ListComments(w http.ResponseWriter, r *http.Request) {}

// CreateComment posts a new comment on a movie.
func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {}
