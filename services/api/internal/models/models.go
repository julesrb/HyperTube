package models

type Movie struct {
	ImdbID      string     `json:"imdb_id"`
	TmdbID      string     `json:"tmdb_id"`
	Title       string     `json:"title"`
	Year        string     `json:"year"`
	PosterURL   string     `json:"poster_url"`
	BackdropURL string     `json:"backdrop_url"`
	IMDbRating  float32    `json:"imdb_rating"`
	Genre       []int      `json:"genres"`
	Runtime     int        `json:"runtime_minutes"`
	Note        float32    `json:"note"`
	Summary     string     `json:"summary"`
	Director    string     `json:"director"`
	Cast        []string   `json:"cast"`
	Watched     bool       `json:"watched"`
	Progression float32    `json:"progression"`
	Torrents    []Torrent  `json:"torrent"`
	Subtitles   []Subtitle `json:"subtitles"`
}

type Torrent struct {
	URL      string `json:"url"`
	Quality  string `json:"quality"`
	Size     string `json:"size"`
	Language string `json:"language"`
	Seeds    string `json:"seeds"`
	Source   string `json:"source"`
}

type Subtitle struct {
	URL      string `json:"url"`
	Language string `json:"language"`
}

type MovieSource struct {
	ImdbID string `json:"imdb_id"`
	URL    string `json:"url"`
}
