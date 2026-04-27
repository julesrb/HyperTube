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
		Note         float32 `json:"vote_average"`
		GenreIDs     []int   `json:"genre_ids"`
	} `json:"movie_results"`
}

type creditsResponse struct {
	Overview string `json:"overview"`
	Credits  struct {
		Cast []struct {
			Name  string `json:"name"`
			Order int    `json:"order"`
		} `json:"cast"`
		Crew []struct {
			Name string `json:"name"`
			Job  string `json:"job"`
		} `json:"crew"`
	} `json:"credits"`
}

func (c *Client) get(ctx context.Context, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.apiKey)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, out)
}

func (c *Client) FindByIMDBID(ctx context.Context, imdbID string) (models.Movie, error) {
	var find findResponse
	if err := c.get(ctx, "https://api.themoviedb.org/3/find/"+imdbID+"?external_source=imdb_id&language=en-US", &find); err != nil {
		return models.Movie{}, err
	}
	if len(find.MovieResults) == 0 {
		return models.Movie{}, fmt.Errorf("no TMDB movie found for IMDb ID %s", imdbID)
	}

	m := find.MovieResults[0]
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
		Note:        m.Note,
		Genre:       m.GenreIDs,
	}, nil
}

func (c *Client) GetMovieDetails(ctx context.Context, tmdbID string) (models.MovieDetails, error) {
	var result creditsResponse
	if err := c.get(ctx, "https://api.themoviedb.org/3/movie/"+tmdbID+"?append_to_response=credits&language=en-US", &result); err != nil {
		return models.MovieDetails{}, err
	}

	director := ""
	for _, crew := range result.Credits.Crew {
		if crew.Job == "Director" {
			director = crew.Name
			break
		}
	}

	cast := make([]string, 0, 5)
	for _, member := range result.Credits.Cast {
		if member.Order >= 5 { // Limit the cast to the 5 more important ones
			break
		}
		cast = append(cast, member.Name)
	}

	return models.MovieDetails{
		Summary:  result.Overview,
		Director: director,
		Cast:     cast,
	}, nil
}
