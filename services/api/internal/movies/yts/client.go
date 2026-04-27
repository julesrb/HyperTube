package yts

import (
	"context"
	"encoding/json"
	"fmt"
	"hypertube/api/internal/models"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient() *Client {
	return &Client{
		httpClient: http.DefaultClient,
		baseURL:    "https://en.yts-official.top/",
	}
}

type ytsResponse struct {
	Status  string     `json:"status"`
	Message string     `json:"message"`
	Data    []ytsMovie `json:"data"`
}

type ytsMovie struct {
	Title string `json:"title"`
	Year  string `json:"year"`
	Img   string `json:"img"`
	URL   string `json:"url"`
}

// Orchestrates the search flow: queries YTS, fetches movie details, and return the matching IMDD IDs.
func (c *Client) SearchByTitle(ctx context.Context, title string) ([]models.TrackerSource, error) {
	response, err := c.searchMovies(ctx, title)
	if err != nil {
		return nil, err
	}
	var ytsMovies []ytsMovie = response.Data

	TrackerSources := make([]models.TrackerSource, 0)
	for _, m := range ytsMovies {
		movieURL := c.baseURL + m.URL
		log.Printf("Fetching movie details from URL: %s", movieURL)

		body, err := c.fetchMovieDetails(ctx, movieURL)
		if err != nil {
			log.Printf("Error fetching movie details: %v", err)
			continue
		}

		imdbID := extractIMDbID(body)

		TrackerSources = append(TrackerSources, models.TrackerSource{
			ImdbID: imdbID,
			Source: "YTS",
			URL:    movieURL,
		})
	}

	return TrackerSources, nil
}

// searchMovies queries the YTS search endpoint and decodes the JSON response.
func (c *Client) searchMovies(ctx context.Context, query string) (*ytsResponse, error) {
	params := url.Values{}
	params.Set("query", query)
	clientQuery := c.baseURL + "ajax/search?" + params.Encode()

	body, err := c.querySource(ctx, clientQuery)
	if err != nil {
		return nil, err
	}

	var response ytsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

// fetchMovieDetails fetches the HTML page for a movie and returns the raw bytes.
func (c *Client) fetchMovieDetails(ctx context.Context, movieURL string) ([]byte, error) {
	return c.get(ctx, movieURL)
}

// fetchMovieDetails fetches the HTML page for a movie and returns the raw bytes.
func (c *Client) querySource(ctx context.Context, title string) ([]byte, error) {
	return c.get(ctx, title)
}

// get performs an HTTP GET and returns the response body as bytes.
func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

var imdbRe = regexp.MustCompile(`imdb\.com/title/(tt\d+)`)

func extractIMDbID(body []byte) string {
	m := imdbRe.FindSubmatch(body)
	if m == nil {
		return ""
	}
	return string(m[1])
}

// Extracts resolution, file size, and magnet link from each modal-torrent block.
var torrentBlockRe = regexp.MustCompile(
	`(?s)modal-quality"[^>]*id="modal-quality-(\w+)"` + // 1: resolution (e.g. 1080p)
		`.*?<p>File size</p>\s*<p class="quality-size">([^<]+)</p>` + // 2: file size (e.g. 1.7 GB)
		`.*?<a[^>]+href="(magnet:\?[^"]+)"[^>]+class="[^"]*magnet-download`) // 3: magnet link

// Extracts language from each tech-spec-info block (ordered by quality).
var languageRe = regexp.MustCompile(`icon-volume-medium"></span>\s*([^<]+)`)

// Extracts seed/peer count from each tech-spec-info block.
var seedsRe = regexp.MustCompile(`tech-peers-seeds">P/S</span>\s*([^\s<]+)`)

func (c *Client) FetchTorrents(ctx context.Context, pageURL string) ([]models.Torrent, error) {
	body, err := c.fetchMovieDetails(ctx, pageURL)
	if err != nil {
		return nil, err
	}
	return extractTorrents(body), nil
}

func extractTorrents(body []byte) []models.Torrent {
	modalMatches := torrentBlockRe.FindAllSubmatch(body, -1)
	langMatches := languageRe.FindAllSubmatch(body, -1)
	seedMatches := seedsRe.FindAllSubmatch(body, -1)

	torrents := make([]models.Torrent, 0, len(modalMatches))
	for i, m := range modalMatches {
		t := models.Torrent{
			Quality: string(m[1]),
			Size:    strings.TrimSpace(string(m[2])),
			URL:     string(m[3]),
			Source:  "YTS",
		}
		if i < len(langMatches) {
			t.Language = strings.TrimSpace(string(langMatches[i][1]))
		}
		if i < len(seedMatches) {
			t.Seeds = strings.TrimSpace(string(seedMatches[i][1]))
		}
		torrents = append(torrents, t)
	}
	return torrents
}
