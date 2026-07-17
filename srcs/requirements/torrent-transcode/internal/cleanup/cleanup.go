package cleanup

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Production defaults: delete torrents that have been finished for a month,
// scanning once a day.
const (
	DefaultRetention = 30 * 24 * time.Hour
	DefaultInterval  = 24 * time.Hour
)

// for live tests
// const (
// 	DefaultRetention = 10 * time.Second
// 	DefaultInterval  = 1 * time.Second
// )

// Store is the subset of the torrent store the cleaner needs; an interface so
// tests can supply a fake without a database.
type Store interface {
	FinishedBefore(ctx context.Context, cutoff time.Time) ([]string, error)
	InProgress(ctx context.Context) ([]string, error)
	ResetTorrent(ctx context.Context, id string) error
}

type TorrentDropper interface {
	Drop(id string)
}

type Cleaner struct {
	store           Store
	dropper         TorrentDropper
	retention       time.Duration
	interval        time.Duration
	videoBasePath   string
	torrentBasePath string
}

func NewCleaner(store Store, dropper TorrentDropper, retention, interval time.Duration) *Cleaner {
	return &Cleaner{
		store:           store,
		dropper:         dropper,
		retention:       retention,
		interval:        interval,
		videoBasePath:   "/data/videos",
		torrentBasePath: "/data/torrents",
	}
}

func (c *Cleaner) CleanupInterrupted(ctx context.Context) {
	ids, err := c.store.InProgress(ctx)
	if err != nil {
		log.Printf("cleanup: failed to list interrupted torrents: %v", err)
		return
	}
	for _, id := range ids {
		c.delete(ctx, id)
	}
}

func (c *Cleaner) Run(ctx context.Context) {
	for {
		c.runOnce(ctx)
		time.Sleep(c.interval)
	}
}

func (c *Cleaner) runOnce(ctx context.Context) {
	cutoff := time.Now().Add(-c.retention)
	ids, err := c.store.FinishedBefore(ctx, cutoff)
	if err != nil {
		log.Printf("cleanup: failed to list expired torrents: %v", err)
		return
	}
	for _, id := range ids {
		c.delete(ctx, id)
	}
}

func (c *Cleaner) delete(ctx context.Context, id string) {
	videoPath := filepath.Join(c.videoBasePath, id)
	torrentPath := filepath.Join(c.torrentBasePath, id)

	c.dropper.Drop(id)

	if err := os.RemoveAll(videoPath); err != nil {
		log.Printf("cleanup: %s: failed to remove video data %q: %v", id, videoPath, err)
		return
	}
	if err := os.RemoveAll(torrentPath); err != nil {
		log.Printf("cleanup: %s: failed to remove torrent data %q: %v", id, torrentPath, err)
		return
	}
	if err := c.store.ResetTorrent(ctx, id); err != nil {
		log.Printf("cleanup: %s: failed to reset torrent status: %v", id, err)
		return
	}
	log.Printf("cleanup: %s: removed torrent data and reset status", id)
}
