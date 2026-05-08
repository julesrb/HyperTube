package tmdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hypertube/api/internal/models"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Client struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
}

func NewClient() (*Client, error) {
	key := os.Getenv("TMDB_API_KEY")
	if key == "" {
		return nil, errors.New("TMDB API key is required")
	}
	return &Client{
		httpClient: http.DefaultClient,
		apiKey:     key,
		baseURL:    "https://api.themoviedb.org/3/",
	}, nil
}

const tmdbImageBase = "https://image.tmdb.org/t/p/w500"

type movieResult struct {
	ID           int     `json:"id"`
	Title        string  `json:"title"`
	PosterPath   string  `json:"poster_path"`
	BackdropPath string  `json:"backdrop_path"`
	ReleaseDate  string  `json:"release_date"`
	Note         float32 `json:"vote_average"`
	Runtime      int     `json:"runtime"`
	GenreIDs     []int   `json:"genre_ids"`
}

type findResponse struct {
	MovieResults []movieResult `json:"movie_results"`
}

type searchResponse struct {
	Results []movieResult `json:"results"`
}

type movieResponse struct {
	ImdbID string `json:"imdb_id"`
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

type imagesResponse struct {
	Images struct {
		Backdrops []struct {
			FilePath string `json:"file_path"`
		} `json:"backdrops"`
	} `json:"images"`
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
	if res.StatusCode == http.StatusTooManyRequests {
		log.Println("TMDB rate limit exceeded (429) - wait 1 second before retrying")
		time.Sleep(time.Second)
		return c.get(ctx, url, out)
	}
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
		Runtime:     m.Runtime,
	}, nil
}

func (c *Client) FindByName(ctx context.Context, title string, year int) (models.Movie, error) {
	var search searchResponse
	queryURL := "https://api.themoviedb.org/3/search/movie?query=" + url.QueryEscape(title) + "&year=" + fmt.Sprintf("%d", year) + "&language=en-US"
	if err := c.get(ctx, queryURL, &search); err != nil {
		return models.Movie{}, err
	}
	if len(search.Results) == 0 {
		return models.Movie{}, fmt.Errorf("no TMDB movie found for title %s", title)
	}

	m := search.Results[0]
	yearStr := ""
	if len(m.ReleaseDate) >= 4 {
		yearStr = m.ReleaseDate[:4]
	}

	var detail movieResponse
	_ = c.get(ctx, fmt.Sprintf("https://api.themoviedb.org/3/movie/%d", m.ID), &detail)

	return models.Movie{
		ImdbID:      detail.ImdbID,
		TmdbID:      fmt.Sprintf("%d", m.ID),
		Title:       m.Title,
		Year:        yearStr,
		PosterURL:   tmdbImageBase + m.PosterPath,
		BackdropURL: tmdbImageBase + m.BackdropPath,
		Note:        m.Note,
		Genre:       m.GenreIDs,
		Runtime:     m.Runtime,
	}, nil
}

func (c *Client) GetMovieDetails(ctx context.Context, tmdbID string, language string) (models.MovieDetails, error) {
	var creditsResult creditsResponse
	if err := c.get(ctx, "https://api.themoviedb.org/3/movie/"+tmdbID+"?append_to_response=credits&language="+language, &creditsResult); err != nil {
		return models.MovieDetails{}, err
	}

	var imagesResult imagesResponse
	if err := c.get(ctx, "https://api.themoviedb.org/3/movie/"+tmdbID+"?images?include_image_language=null", &imagesResult); err != nil {
		return models.MovieDetails{}, err
	}

	director := make([]string, 0)
	for _, crew := range creditsResult.Credits.Crew {
		if crew.Job == "Director" {
			director = append(director, crew.Name)
			break
		}
	}

	cast := make([]string, 0, 5)
	for _, member := range creditsResult.Credits.Cast {
		if member.Order >= 5 { // Limit the cast to the 5 more important ones
			break
		}
		cast = append(cast, member.Name)
	}

	extraBackdrops := make([]string, 0)
	for _, img := range imagesResult.Images.Backdrops {
		extraBackdrops = append(extraBackdrops, tmdbImageBase+img.FilePath)
	}

	return models.MovieDetails{
		Summary:        creditsResult.Overview,
		Director:       director,
		Cast:           cast,
		ExtraBackdrops: extraBackdrops,
	}, nil
}
