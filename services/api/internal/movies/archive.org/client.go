package archiveorg

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
	"sort"
	"strconv"
	"strings"
)

type Client struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
}

func NewClient() (*Client, error) {
	return &Client{
		httpClient: http.DefaultClient,
		apiKey:     "",
		baseURL:    "https://archive.org/advancedsearch.php",
	}, nil
}

type archiveSearchResponse struct {
	Response archiveResponseBody `json:"response"`
}

type archiveResponseBody struct {
	Docs []archiveDoc `json:"docs"`
}

type archiveDoc struct {
	Identifier string `json:"identifier"`
	Title      string `json:"title"`
	Year       int    `json:"year"`
	Downloads  int    `json:"downloads"`
	Btih       string `json:"btih"`
	ItemSize   int64  `json:"item_size"`
}

func (c *Client) SearchByTitle(ctx context.Context, title string) ([]models.Torrent, error) {
	params := url.Values{
		"q":      {fmt.Sprintf("title:(%s) AND collection:feature_films", title)},
		"fl[]":   {"identifier,title,year,downloads,btih,item_size"},
		"sort[]": {"downloads desc"},
		"rows":   {"10"},
		"output": {"json"},
	}
	queryURL := "https://archive.org/advancedsearch.php?" + params.Encode()

	log.Printf("Archive.org query: %s", queryURL)
	return c.fetch(ctx, queryURL)
}

// FetchTop retrieves the top Lastest torrent registered on c411
func (c *Client) FetchTop(ctx context.Context, limit int) ([]models.Torrent, error) {
	params := url.Values{"t": {"movie"}, "q": {""}, "cat": {"2000"}, "limit": {strconv.Itoa(limit)}, "apikey": {c.apiKey}}
	queryURL := c.baseURL + "torznab?" + params.Encode()
	log.Printf("C411 top query: %s", queryURL)
	return c.fetch(ctx, queryURL)
}

var yearInParensRe = regexp.MustCompile(`\((\d{4})\)`)

func stripYear(title string, y int) (string, int) {
	if y == 0 {
		m := yearInParensRe.FindStringSubmatch(title)
		if m == nil {
			return title, 0
		}
		y, _ = strconv.Atoi(m[1])
	}
	year := strconv.Itoa(y)
	re := regexp.MustCompile(`[\s,.()\[\]]+` + regexp.QuoteMeta(year) + `[\s,.()\[\]]*`)
	return strings.Join(strings.Fields(re.ReplaceAllString(title, " ")), " "), y
}

func (c *Client) fetch(ctx context.Context, queryURL string) ([]models.Torrent, error) {
	body, err := c.get(ctx, queryURL)
	if err != nil {
		return nil, err
	}

	var response archiveSearchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	torrents := make([]models.Torrent, 0, len(response.Response.Docs))
	for _, torrent := range response.Response.Docs {
		magnetURL := "magnet:?xt=urn:btih:" + torrent.Btih + "&dn=" + url.QueryEscape(torrent.Title)
		Title, Year := stripYear(torrent.Title, torrent.Year)
		torrents = append(torrents, models.Torrent{
			ImdbID:   "none",
			Title:    Title,
			Year:     Year,
			Source:   "archive.org",
			URL:      magnetURL,
			Quality:  "unknown",
			Size:     formatBytes(torrent.ItemSize),
			Language: "unknown",
			Seeds:    "2+",
		})
	}
	return torrents, nil
}

func (c *Client) GetTopMovies(ctx context.Context) ([]models.Torrent, error) {
	torrents, err := c.FetchTop(ctx, 100) // Get the 100 most recent torrent list
	if err != nil {
		return nil, err
	}

	sort.Slice(torrents, func(i, j int) bool {
		si, _ := strconv.Atoi(torrents[i].Seeds)
		sj, _ := strconv.Atoi(torrents[j].Seeds)
		return si > sj
	})

	seen := make(map[string]bool)
	result := make([]models.Torrent, 0, 9)
	for _, t := range torrents {
		if seen[t.ImdbID] {
			continue
		}
		seen[t.ImdbID] = true
		result = append(result, t)
		if len(result) == 9 {
			break
		}
	}
	return result, nil
}

func (c *Client) get(ctx context.Context, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
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

var qualityRe = regexp.MustCompile(`\b(2160p|1080p|720p|480p)\b`)

func extractQuality(title string) string {
	if m := qualityRe.FindString(title); m != "" {
		return m
	}
	if strings.Contains(title, "4K") || strings.Contains(title, "4KLight") {
		return "2160p"
	}
	return ""
}

var languageRe = regexp.MustCompile(`(?i)\b(MULTI|VFF|VF2|VF|VO|VOSTFR|FRENCH)\b`)

func extractLanguage(title string) string {
	return strings.Join(languageRe.FindAllString(title, -1), "/")
}

func formatBytes(b int64) float64 {
	const unit = 1024
	if b < unit {
		return 0.01
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return float64(b) / float64(div)
}
