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
	upsertMovie(ctx context.Context, m models.Movie) error
	upsertTrackerSource(ctx context.Context, ts models.TrackerSource) error
}

type MovieSearcher interface {
	SearchByTitle(ctx context.Context, title string) ([]models.TrackerSource, error)
}

type tmdbClient interface {
	FindByIMDBID(ctx context.Context, imdbID string) (models.Movie, error)
	GetMovieDetails(ctx context.Context, tmdbID string) (models.MovieDetails, error)
}

type Handler struct {
	store     movieStore
	searchers []MovieSearcher
	tmdb      tmdbClient
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

	details, err := h.tmdb.GetMovieDetails(r.Context(), movie.TmdbID)
	if err != nil {
		log.Printf("TMDB details error for TmdbID %s: %v", movie.TmdbID, err)
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch movie details")
		return
	}
	movie.Summary = details.Summary
	movie.Director = details.Director
	movie.Cast = details.Cast

	//TODO retrieve the source of the movie and crawl for links. if no match exit
	//TODO fetch torrent info
	respond.Item(w, http.StatusOK, toMovieDetailResponse(*movie))
}

func (h *Handler) SearchMovies(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	if title == "" {
		respond.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "title query parameter is required")
		return
	}
	log.Printf("searching for movies with title: %s", title)
	trackerSources, err := h.searchers[0].SearchByTitle(r.Context(), title) // TODO Nest and add the second torrent source
	if err != nil {
		log.Println("search err:", err)
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to search movies")
		return
	}

	movies := make([]movieResponse, 0)
	for i, trackerSource := range trackerSources {
		if i >= 10 { // Protect TMDB api call per second limit
			break
		}
		// TODO OPTI look for preexisting data in db
		movie, err := h.tmdb.FindByIMDBID(r.Context(), trackerSource.ImdbID)
		if err != nil {
			log.Printf("TMDB lookup error for IMDb ID %s: %v", trackerSource.ImdbID, err)
			continue
		}
		if err = h.store.upsertMovie(r.Context(), movie); err != nil {
			log.Println("db err:", err)
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to store movie")
			return
		}
		if err = h.store.upsertTrackerSource(r.Context(), trackerSource); err != nil {
			log.Println("db err:", err)
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to store tracker source")
			return
		}
		movies = append(movies, toMovieResponse(movie))
	}
	respond.List(w, http.StatusOK, movies, len(movies))
}

// ListComments returns comments for a movie.
func (h *Handler) ListComments(w http.ResponseWriter, r *http.Request) {}

// CreateComment posts a new comment on a movie.
func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {}
