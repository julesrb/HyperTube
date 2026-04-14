package movies

import (
	"context"
	"errors"
	"log"
	"net/http"

	"hypertube/api/internal/models"
	"hypertube/api/internal/respond"

	"github.com/jackc/pgx/v5/pgxpool"
)

type movieStore interface {
	listFeatured(ctx context.Context) ([]models.Movie, error)
	findByID(ctx context.Context, id string) (*models.Movie, error)
}

type Handler struct {
	store movieStore
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{store: NewStore(db)}
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
	respond.Item(w, http.StatusOK, toMovieResponse(*movie))
}

// ListComments returns comments for a movie.
func (h *Handler) ListComments(w http.ResponseWriter, r *http.Request) {}

// CreateComment posts a new comment on a movie.
func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {}
