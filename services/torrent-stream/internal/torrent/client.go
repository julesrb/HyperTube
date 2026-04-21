package torrent

import (
	"errors"
	"io"
	"path/filepath"
	"strings"
)

// Client wraps anacrolix/torrent and handles sequential piece prioritization.
type Client struct{}

var errNoMainVideoFile = errors.New("no main video file found")

type fileCandidate struct {
	Path string
	Size int64
}

// Add starts downloading a magnet link and returns an io.Reader over the
// largest file in the torrent (assumed to be the video file).
func (c *Client) Add(magnetURI string) (io.Reader, error) {
	return nil, nil
}

func pickMainVideoFile(files []fileCandidate) (fileCandidate, error) {
	var (
		selected fileCandidate
		found    bool
	)

	for _, file := range files {
		if !isPlayableVideoFile(file.Path) || isSampleFile(file.Path) {
			continue
		}

		if !found || file.Size > selected.Size {
			selected = file
			found = true
		}
	}

	if !found {
		return fileCandidate{}, errNoMainVideoFile
	}

	return selected, nil
}

func isPlayableVideoFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mkv", ".mp4", ".webm":
		return true
	default:
		return false
	}
}

func isSampleFile(path string) bool {
	return strings.Contains(strings.ToLower(filepath.Base(path)), "sample")
}
