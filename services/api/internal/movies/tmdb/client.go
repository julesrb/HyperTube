package tmdb

import (
	"context"
	"encoding/json"
	"fmt"
	"hypertube/api/internal/models"
	"io"
	"net/http"
	"os"
)

type Client struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
}

func NewClient() *Client {
	key := os.Getenv("TMDB_API_KEY")
	if key == "" {
		panic("TMDB API key is required")
	}
	return &Client{
		httpClient: http.DefaultClient,
		apiKey:     key,
		baseURL:    "https://api.themoviedb.org/3/",
	}
}

const tmdbImageBase = "https://image.tmdb.org/t/p/w500"

type findResponse struct {
	MovieResults []struct {
		ID           int     `json:"id"`
		Title        string  `json:"title"`
		PosterPath   string  `json:"poster_path"`
		BackdropPath string  `json:"backdrop_path"`
		ReleaseDate  string  `json:"release_date"`
		Genre        []int   `json:"genre_ids"`
		Note         float32 `json:"vote_average"`
	} `json:"movie_results"`
}

func (c *Client) FindByIMDBID(ctx context.Context, imdbID string) (models.Movie, error) {
	url := "https://api.themoviedb.org/3/find/" + imdbID + "?external_source=imdb_id&language=en-US"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return models.Movie{}, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.apiKey)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return models.Movie{}, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return models.Movie{}, err
	}

	var result findResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return models.Movie{}, err
	}

	if len(result.MovieResults) == 0 {
		return models.Movie{}, fmt.Errorf("no TMDB movie found for IMDb ID %s", imdbID)
	}

	m := result.MovieResults[0]
	year := ""
	if len(m.ReleaseDate) >= 4 {
		year = m.ReleaseDate[:4]
	}

	return models.Movie{
		ImdbID:      imdbID,
		TmdbID:      fmt.Sprintf("%d", m.ID),
		Title:       m.Title,
		Year:        year,
		PosterURL:   tmdbImageBase + m.PosterPath,
		BackdropURL: tmdbImageBase + m.BackdropPath,
		Genre:       m.Genre,
		Note:        m.Note,
	}, nil
}
