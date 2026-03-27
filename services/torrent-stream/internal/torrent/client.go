package torrent

import "io"

// Client wraps anacrolix/torrent and handles sequential piece prioritization.
type Client struct{}

// Add starts downloading a magnet link and returns an io.Reader over the
// largest file in the torrent (assumed to be the video file).
func (c *Client) Add(magnetURI string) (io.Reader, error) {
	return nil, nil
}
