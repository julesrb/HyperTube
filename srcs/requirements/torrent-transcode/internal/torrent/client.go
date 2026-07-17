package torrent

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	torrentlib "github.com/anacrolix/torrent"
)

type Client struct {
	cl         *torrentlib.Client
	streams    *map[string]io.ReadSeekCloser  // torrentID → reader
	torrents   map[string]*torrentlib.Torrent // torrentID → torrent handle
	mu         *sync.Mutex
	torrentDir string
}

func NewClient(mu *sync.Mutex, streams *map[string]io.ReadSeekCloser) (*Client, error) {
	torrentDir := "/data/torrents"
	cl, err := torrentlib.NewClient(nil)
	if err != nil {
		return nil, fmt.Errorf("torrent client: %w", err)
	}
	log.Printf("torrent client listening on: %v", cl.ListenAddrs())
	return &Client{cl: cl, streams: streams, torrents: make(map[string]*torrentlib.Torrent), mu: mu, torrentDir: torrentDir}, nil
}

func (c *Client) Close() {
	c.cl.Close()
}

func (c *Client) Drop(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if r, ok := (*c.streams)[id]; ok {
		r.Close()
		delete(*c.streams, id)
	}
	if t, ok := c.torrents[id]; ok {
		t.Drop()
		delete(c.torrents, id)
		log.Printf("%s: dropped torrent from client", id)
	}
}

// Add downloads the .torrent file at torrentURL to /data/torrent/, then starts
// streaming and returns a reader over the largest file (assumed to be the video).
func (c *Client) Add(id, torrentPath string) (io.ReadSeekCloser, error) {

	torrent, err := c.cl.AddTorrentFromFile(torrentPath)
	if err != nil {
		return nil, fmt.Errorf("add torrent: %w", err)
	}

	c.mu.Lock()
	c.torrents[id] = torrent
	c.mu.Unlock()

	<-torrent.GotInfo()

	var largest *torrentlib.File
	for _, f := range torrent.Files() {
		if largest == nil || f.Length() > largest.Length() {
			largest = f
		}
	}
	if largest == nil {
		return nil, fmt.Errorf("torrent has no files")
	}

	largest.Download()

	const minBuffer = 10 * 1024 * 1024 // 10 MB before making reader available for probing
	for largest.BytesCompleted() < minBuffer {
		time.Sleep(500 * time.Millisecond)
	}
	log.Printf("buffer ready, ready for transcoding")

	reader := largest.NewReader()
	reader.SetReadahead(largest.Length())

	return reader, nil
}

// torrentFilename is the fixed name the api service saves each .torrent under,
// inside its per-torrent folder (<torrentDir>/<id>/). Must match the api side.
const torrentFilename = "source.torrent"

// TorrentFilePath resolves the local path of the .torrent file for the torrent
// with the given id. The file itself is fetched by the api service (which runs
// outside the VPN) and written to <torrentDir>/<id>/source.torrent; this only
// derives the path and verifies the file is present.
func (c *Client) TorrentFilePath(id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("empty torrent id")
	}

	destPath := filepath.Join(c.torrentDir, id, torrentFilename)

	info, err := os.Stat(destPath)
	if err != nil {
		return "", fmt.Errorf("torrent file not found at %q (downloader did not run?): %w", destPath, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("expected torrent file but %q is a directory", destPath)
	}

	return destPath, nil
}
