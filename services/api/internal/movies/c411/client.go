package c411

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"hypertube/api/internal/models"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
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
	key := os.Getenv("C411_API_KEY")
	if key == "" {
		return nil, errors.New("C411 API key is required")
	}
	return &Client{
		httpClient: http.DefaultClient,
		apiKey:     key,
		baseURL:    "https://c411.org/api/",
	}, nil
}

type torznabFeed struct {
	XMLName xml.Name       `xml:"rss"`
	Channel torznabChannel `xml:"channel"`
}

type torznabChannel struct {
	Items []torznabItem `xml:"item"`
}

type torznabItem struct {
	Title     string           `xml:"title"`
	GUID      string           `xml:"guid"`
	Link      string           `xml:"link"`
	PubDate   string           `xml:"pubDate"`
	Size      int64            `xml:"size"`
	Enclosure torznabEnclosure `xml:"enclosure"`
	Attrs     []torznabAttr    `xml:"http://torznab.com/schemas/2015/feed attr"`
}

type torznabEnclosure struct {
	URL    string `xml:"url,attr"`
	Length int64  `xml:"length,attr"`
}

type torznabAttr struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

func (item *torznabItem) attr(name string) string {
	for _, a := range item.Attrs {
		if a.Name == name {
			return a.Value
		}
	}
	return ""
}

func (c *Client) SearchByTitle(ctx context.Context, title string) ([]models.Torrent, error) {
	params := url.Values{"t": {"movie"}, "cat": {"2000"}, "q": {title}, "apikey": {c.apiKey}}
	queryURL := c.baseURL + "torznab?" + params.Encode()
	log.Printf("C411 query: %s", queryURL)
	return c.fetch(ctx, queryURL)
}

// FetchTop retrieves the top Lastest torrent registered on c411
func (c *Client) FetchTop(ctx context.Context, limit int) ([]models.Torrent, error) {
	params := url.Values{"t": {"movie"}, "q": {""}, "cat": {"2000"}, "limit": {strconv.Itoa(limit)}, "apikey": {c.apiKey}}
	queryURL := c.baseURL + "torznab?" + params.Encode()
	log.Printf("C411 top query: %s", queryURL)
	return c.fetch(ctx, queryURL)
}

func (c *Client) fetch(ctx context.Context, queryURL string) ([]models.Torrent, error) {
	body, err := c.get(ctx, queryURL)
	if err != nil {
		return nil, err
	}

	var feed torznabFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, err
	}

	torrents := make([]models.Torrent, 0, len(feed.Channel.Items))
	for _, item := range feed.Channel.Items {
		imdbID := item.attr("imdbid")
		if imdbID == "" {
			continue
		}
		infohash := item.attr("infohash")
		magnetURL := "magnet:?xt=urn:btih:" + infohash + "&dn=" + url.QueryEscape(item.Title)
		torrents = append(torrents, models.Torrent{
			ImdbID:   imdbID,
			Title:    item.Title,
			Source:   "C411",
			URL:      magnetURL,
			Quality:  extractQuality(item.Title),
			Size:     formatBytes(item.Size),
			Language: extractLanguage(item.Title),
			Seeds:    item.attr("seeders"),
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
