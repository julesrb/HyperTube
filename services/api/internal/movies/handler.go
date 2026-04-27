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
	SearchByTitle(ctx context.Context, title string) ([]models.MovieSource, error)
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

	movieResponse := make([]movieResponse, len(movies))
	for i, m := range movies {
		movieResponse[i] = toMovieResponse(m)
	}
	respond.List(w, http.StatusOK, movieResponse, len(movieResponse))
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
	//TODO retrieve the source of the movie and crawl for links. if no match exit
	//TODO a full tmdb fetch to get the actual details
	//TODO fetch torrent info
	respond.Item(w, http.StatusOK, movie)
}

func (h *Handler) SearchMovies(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	if title == "" {
		respond.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "title query parameter is required")
		return
	}
	moviesSources, err := h.searchers[0].SearchByTitle(r.Context(), title) // TODO Nest and add the second torrent source
	log.Printf("searching for movies with title: %s", title)
	if err != nil {
		log.Println("search err:", err)
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to search movies")
		return
	}
	//TODO store each source URL to DB
	movies := make([]movieResponse, 0, 11) // Limit to first 10 results MANUAL SAFE LIMIT FOR TMDB
	for _, moviesSource := range moviesSources {
		movie, err := h.tmdb.FindByIMDBID(r.Context(), moviesSource.ImdbID)
		if err != nil {
			log.Printf("TMDB lookup error for IMDb ID %s: %v", moviesSource.ImdbID, err)
			continue
		}
		movies = append(movies, toMovieResponse(movie))
	}
	respond.List(w, http.StatusOK, movies, len(movies))
}

// ListComments returns comments for a movie.
func (h *Handler) ListComments(w http.ResponseWriter, r *http.Request) {}

// CreateComment posts a new comment on a movie.
func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {}
