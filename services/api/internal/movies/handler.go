package movies

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"hypertube/api/internal/models"
	"hypertube/api/internal/respond"
)

type MoviesHandler struct {
	store     movieStore
	searchers []MovieSearcher
	tmdb tmdbClient
}

type movieStore interface {
	listFeatured(ctx context.Context) ([]models.Movie, error)
	findByID(ctx context.Context, id string) (*models.Movie, error)
	UpsertMovie(ctx context.Context, m models.Movie) error
	UpsertTorrent(ctx context.Context, ts models.Torrent) error
	findTorrent(ctx context.Context, imdbID string) ([]models.Torrent, error)
	listComments(ctx context.Context, imdbID string) ([]models.Comment, error)
	createComment(ctx context.Context, c models.Comment) (models.Comment, error)
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


func NewMoviesHandler(store movieStore, searchers []MovieSearcher, tmdb tmdbClient) *MoviesHandler {
	return &MoviesHandler{store: store, searchers: searchers, tmdb: tmdb}
}

// GetMovies returns a list of movies.
func (h *MoviesHandler) GetMovies(w http.ResponseWriter, r *http.Request) {
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
func (h *MoviesHandler) GetMoviesId(w http.ResponseWriter, r *http.Request) {
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

func (h *MoviesHandler) collectTorrents(ctx context.Context, title string) ([]models.Torrent, error) {
	var perSource [][]models.Torrent
	for _, s := range h.searchers {
		torrents, err := s.SearchByTitle(ctx, title)
		if err != nil {
			return nil, err
		}
		perSource = append(perSource, torrents)
	}
	maxLen := 0
	for _, t := range perSource {
		if len(t) > maxLen {
			maxLen = len(t)
		}
	}
	mixed := make([]models.Torrent, 0, maxLen*len(perSource))
	for i := range maxLen {
		for _, t := range perSource {
			if i < len(t) {
				mixed = append(mixed, t[i])
			}
		}
	}
	return mixed, nil
}

func (h *MoviesHandler) resolveMovie(ctx context.Context, torrent models.Torrent) (models.Movie, models.Torrent, error) {
	if torrent.ImdbID == "none" {
		movie, err := h.tmdb.FindByName(ctx, torrent.Title, torrent.Year)
		if err != nil {
			return models.Movie{}, torrent, err
		}
		torrent.ImdbID = movie.ImdbID
		return movie, torrent, nil
	}
	movie, err := h.tmdb.FindByIMDBID(ctx, torrent.ImdbID)
	return movie, torrent, err
}

func (h *MoviesHandler) SearchMovies(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	if title == "" {
		respond.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "title query parameter is required")
		return
	}
	log.Printf("searching for movies with title: %s", title)

	torrents, err := h.collectTorrents(r.Context(), title)
	if err != nil {
		log.Println("search err:", err)
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to search movies")
		return
	}

	movies := make([]movieResponse, 0)
	imdbIdSeen := make(map[string]bool)

	for _, torrent := range torrents {
		if len(movies) >= 8 { // Protect TMDB api call per second limit
			break
		}
		if imdbIdSeen[torrent.ImdbID] {
			h.store.UpsertTorrent(r.Context(), torrent)
			continue
		}
		movie, torrent, err := h.resolveMovie(r.Context(), torrent)
		if err != nil {
			log.Printf("TMDB lookup error for %q: %v", torrent.Title, err)
			continue
		}
		if !imdbIdSeen[movie.ImdbID] {
			if err = h.store.UpsertMovie(r.Context(), movie); err != nil {
				log.Println("db err:", err)
				respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to store movie")
				return
			}
			movies = append(movies, toMovieResponse(movie))
			imdbIdSeen[movie.ImdbID] = true
		}
		if err = h.store.UpsertTorrent(r.Context(), torrent); err != nil {
			log.Println("db err:", err)
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to store torrent")
			return
		}
	}
	respond.List(w, http.StatusOK, movies, len(movies))
}

func (h *MoviesHandler) GetMovieTorrents(w http.ResponseWriter, r *http.Request) {
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
func (h *MoviesHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	imdbid := r.PathValue("id")
	comments, err := h.store.listComments(r.Context(), imdbid)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			respond.Error(w, http.StatusNotFound, "NOT_FOUND", "no comments")
		} else {
			log.Println("db err:", err)
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to acess comments")
		}
		return
	}
	respond.List(w, http.StatusOK, comments, len(comments))
}

// CreateComment posts a new comment on a movie.
func (h *MoviesHandler) PostComment(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID  int    `json:"user_id"`
		MovieID string `json:"movie_id"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println("decode err:", err)
		respond.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	comment := models.Comment{
		UserID:  input.UserID,
		MovieID: input.MovieID,
		Content: input.Content,
	}
	if comment, err := h.store.createComment(r.Context(), comment); err != nil {
		log.Println("db err:", err)
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create comment")
		return
	} else {
		respond.Item(w, http.StatusCreated, comment)
	}
}
