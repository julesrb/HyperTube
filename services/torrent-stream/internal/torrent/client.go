package torrent

import (
	"errors"
	"io"
	"path/filepath"
	"strings"
)

// Client coordinates magnet opening and main video file selection.
type Client struct {
	opener magnetOpener
}

var errEmptyMagnetURI = errors.New("empty magnet uri")
var errNoMagnetOpener = errors.New("no magnet opener configured")
var errNoMainVideoFile = errors.New("no main video file found")

type fileCandidate struct {
	Path string
	Size int64
}

// magnetOpener is injected so Add can be tested without talking to a real
// torrent library yet. This keeps the second iteration focused on orchestration.
type magnetOpener interface {
	Open(magnetURI string) (torrentHandle, error)
}

// torrentHandle exposes only the data needed for this iteration: the list of
// files discovered for a magnet.
type torrentHandle interface {
	Files() []torrentFile
}

// torrentFile is intentionally tiny so the selection logic can stay independent
// from whichever torrent implementation we plug in later.
type torrentFile interface {
	Path() string
	Size() int64
	NewReader() (io.Reader, error)
}

// Add starts downloading a magnet link and returns an io.Reader over the
// largest file in the torrent (assumed to be the video file).
func (c *Client) Add(magnetURI string) (io.Reader, error) {
	// Reject empty input before we touch any external dependency.
	if strings.TrimSpace(magnetURI) == "" {
		return nil, errEmptyMagnetURI
	}

	// This guard makes the missing wiring explicit instead of panicking on nil.
	if c.opener == nil {
		return nil, errNoMagnetOpener
	}

	handle, err := c.opener.Open(magnetURI)
	if err != nil {
		return nil, err
	}

	file, err := selectMainVideoFile(handle.Files())
	if err != nil {
		return nil, err
	}

	return file.NewReader()
}

func pickMainVideoFile(files []fileCandidate) (fileCandidate, error) {
	var (
		selected fileCandidate
		found    bool
	)

	for _, file := range files {
		// Keep the first rule set pure and deterministic: only browser-relevant
		// video files compete, and common sample clips are excluded.
		if !isPlayableVideoFile(file.Path) || isSampleFile(file.Path) {
			continue
		}

		// Picking the largest remaining file is a simple heuristic for "main movie"
		// that works well enough for the first iterations.
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
	// Many torrents ship a short "sample" clip next to the real movie.
	// We filter those out early so they never win by size.
	return strings.Contains(strings.ToLower(filepath.Base(path)), "sample")
}

func selectMainVideoFile(files []torrentFile) (torrentFile, error) {
	// Convert rich torrent file objects into plain candidates so the selection
	// rules stay isolated from I/O and easy to unit test.
	candidates := make([]fileCandidate, 0, len(files))
	for _, file := range files {
		candidates = append(candidates, fileCandidate{
			Path: file.Path(),
			Size: file.Size(),
		})
	}

	selected, err := pickMainVideoFile(candidates)
	if err != nil {
		return nil, err
	}

	// Map the selected metadata back to the original torrent file so callers can
	// open a reader on the real object returned by the opener.
	for _, file := range files {
		if file.Path() == selected.Path && file.Size() == selected.Size {
			return file, nil
		}
	}

	return nil, errNoMainVideoFile
}
