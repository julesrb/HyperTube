package movies

import (
	"context"
	"errors"
	"log"
	"net/http"

	"hypertube/api/internal/models"
	"hypertube/api/internal/respond"
)

type movieStore interface {
	listFeatured(ctx context.Context) ([]models.Movie, error)
	findByID(ctx context.Context, id string) (*models.Movie, error)
	UpsertMovie(ctx context.Context, m models.Movie) error
	UpsertTorrent(ctx context.Context, ts models.Torrent) error
	findTorrent(ctx context.Context, imdbID string) ([]models.Torrent, error)
}

type MovieSearcher interface {
	SearchByTitle(ctx context.Context, title string) ([]models.Torrent, error)
	GetTopMovies(ctx context.Context) ([]models.Torrent, error)
}

type tmdbClient interface {
	FindByIMDBID(ctx context.Context, imdbID string) (models.Movie, error)
	GetMovieDetails(ctx context.Context, tmdbID string) (models.MovieDetails, error)
	FindByName(ctx context.Context, title string, year int) (models.Movie, error)
}

type Handler struct {
	store     movieStore
	searchers []MovieSearcher
	// fetchers  []TorrentFetcher
	tmdb tmdbClient
}

func NewHandler(store movieStore, searchers []MovieSearcher, tmdb tmdbClient) *Handler {
	return &Handler{store: store, searchers: searchers, tmdb: tmdb}
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

	// TODO OPTI look for preexisting data in db
	details, err := h.tmdb.GetMovieDetails(r.Context(), movie.TmdbID)
	if err != nil {
		log.Printf("TMDB details error for TmdbID %s: %v", movie.TmdbID, err)
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch movie details")
		return
	}
	movie.Summary = details.Summary
	movie.Director = details.Director
	movie.Cast = details.Cast

	respond.Item(w, http.StatusOK, toMovieDetailResponse(*movie))
}

func (h *Handler) SearchMovies(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	if title == "" {
		respond.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "title query parameter is required")
		return
	}
	log.Printf("searching for movies with title: %s", title)

	var perSource [][]models.Torrent
	for _, s := range h.searchers {
		torrents, err := s.SearchByTitle(r.Context(), title)
		if err != nil {
			log.Println("search err:", err)
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to search movies")
			return
		}
		perSource = append(perSource, torrents)
	}
	maxLen := 0
	for _, t := range perSource {
		if len(t) > maxLen {
			maxLen = len(t)
		}
	}
	torrentsMix := make([]models.Torrent, 0, maxLen*len(perSource))
	for i := range maxLen {
		for _, t := range perSource {
			if i < len(t) {
				torrentsMix = append(torrentsMix, t[i])
			}
		}
	}

	movies := make([]movieResponse, 0)
	imdbIdSeen := make(map[string]bool)
	uniqueMovie := 0

	for _, torrent := range torrentsMix {
		if uniqueMovie >= 8 { // Protect TMDB api call per second limit
			break
		}

		if torrent.ImdbID == "none" {
			movie, err := h.tmdb.FindByName(r.Context(), torrent.Title, torrent.Year)
			if err != nil {
				log.Printf("TMDB lookup error for IMDb ID %s: %v", torrent.ImdbID, err)
				continue
			}
			torrent.ImdbID = movie.ImdbID
			if !imdbIdSeen[torrent.ImdbID] {
				if err = h.store.UpsertMovie(r.Context(), movie); err != nil {
					log.Println("db err:", err)
					respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to store movie")
					return
				}
				movies = append(movies, toMovieResponse(movie))
				imdbIdSeen[movie.ImdbID] = true
				uniqueMovie++
			}
		} else {
			if !imdbIdSeen[torrent.ImdbID] {
				// TODO OPTI look for preexisting data in db
				movie, err := h.tmdb.FindByIMDBID(r.Context(), torrent.ImdbID)
				if err != nil {
					log.Printf("TMDB lookup error for IMDb ID %s: %v", torrent.ImdbID, err)
					continue
				}
				if err = h.store.UpsertMovie(r.Context(), movie); err != nil {
					log.Println("db err:", err)
					respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to store movie")
					return
				}
				movies = append(movies, toMovieResponse(movie))
				imdbIdSeen[movie.ImdbID] = true
				uniqueMovie++
			}
		}
		if err := h.store.UpsertTorrent(r.Context(), torrent); err != nil {
			log.Println("db err:", err)
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to store torrent")
			return
		}
	}
	respond.List(w, http.StatusOK, movies, len(movies))
}

func (h *Handler) GetMovieTorrents(w http.ResponseWriter, r *http.Request) {
	imdbid := r.PathValue("id")
	torrents, err := h.store.findTorrent(r.Context(), imdbid)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			respond.Error(w, http.StatusNotFound, "NOT_FOUND", "no tracker source found for this movie")
		} else {
			log.Println("db err:", err)
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to load tracker source")
		}
		return
	}
	respond.List(w, http.StatusOK, torrents, len(torrents))
}

// ListComments returns comments for a movie.
func (h *Handler) ListComments(w http.ResponseWriter, r *http.Request) {}

// CreateComment posts a new comment on a movie.
func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {}
