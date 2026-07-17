package downloader

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Downloader struct {
	TorrentDir string
}

func New(torrentDir string) *Downloader {
	return &Downloader{TorrentDir: torrentDir}
}

var ErrSourceTimeout = errors.New("torrent source timed out")

// isTimeout reports whether err is a network timeout.
func isTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

// DownloadTorrentFile fetches the .torrent at torrentURL into a per-torrent
// folder named by id (<TorrentDir>/<id>/source.torrent) and returns its local path.
func (d *Downloader) DownloadTorrentFile(id, torrentURL string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("empty torrent id")
	}
	if torrentURL == "" {
		return "", fmt.Errorf("empty torrent url")
	}

	destDir := filepath.Join(d.TorrentDir, id)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("create torrent dir: %w", err)
	}

	destPath := filepath.Join(destDir, "source.torrent")

	// Only treat the cache as a hit if it's a real file, not a directory.
	if info, err := os.Stat(destPath); err == nil && !info.IsDir() {
		return destPath, nil
	}

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodGet, torrentURL, nil)
	if err != nil {
		return "", fmt.Errorf("build request for %q: %w", torrentURL, err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (HyperTube)")

	resp, err := client.Do(req)
	if err != nil {
		if isTimeout(err) {
			return "", fmt.Errorf("download torrent file %q: %w", torrentURL, ErrSourceTimeout)
		}
		return "", fmt.Errorf("download torrent file %q: %w", torrentURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("download torrent file %q: status %d, content-type %q, body %q",
			torrentURL, resp.StatusCode, resp.Header.Get("Content-Type"), body)
	}

	buf := bufio.NewReader(resp.Body)
	// This catches HTML error/captcha pages served with a 200 status.
	if b, err := buf.Peek(1); err != nil || b[0] != 'd' {
		body, _ := io.ReadAll(io.LimitReader(buf, 512))
		return "", fmt.Errorf("download torrent file %q: not a torrent (content-type %q, body %q)",
			torrentURL, resp.Header.Get("Content-Type"), body)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("create torrent file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, buf); err != nil {
		if isTimeout(err) {
			return "", fmt.Errorf("write torrent file %q: %w", torrentURL, ErrSourceTimeout)
		}
		return "", fmt.Errorf("write torrent file: %w", err)
	}

	return destPath, nil
}
