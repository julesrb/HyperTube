package cleanup

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// fakeStore implements cleanupStore in memory: it tracks when each torrent
// finished and which torrents were reset.
type fakeStore struct {
	mu         sync.Mutex
	finishedAt map[string]time.Time
	inProgress map[string]bool
	reset      []string
}

func (f *fakeStore) FinishedBefore(_ context.Context, cutoff time.Time) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var ids []string
	for id, ts := range f.finishedAt {
		if ts.Before(cutoff) {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func (f *fakeStore) InProgress(_ context.Context) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var ids []string
	for id := range f.inProgress {
		ids = append(ids, id)
	}
	return ids, nil
}

func (f *fakeStore) ResetTorrent(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.finishedAt, id)
	delete(f.inProgress, id)
	f.reset = append(f.reset, id)
	return nil
}

func (f *fakeStore) wasReset(id string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, r := range f.reset {
		if r == id {
			return true
		}
	}
	return false
}

// fakeDropper implements TorrentDropper, recording which torrents were dropped.
type fakeDropper struct {
	mu      sync.Mutex
	dropped []string
}

func (d *fakeDropper) Drop(id string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.dropped = append(d.dropped, id)
}

func (d *fakeDropper) wasDropped(id string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, dr := range d.dropped {
		if dr == id {
			return true
		}
	}
	return false
}

// mustWriteFile creates dir and a file inside it, failing the test on error.
func mustWriteFile(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %q: %v", dir, err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
		t.Fatalf("write %q: %v", name, err)
	}
}

func TestCleanerDeletesAfterRetention(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping time-based cleanup test in -short mode")
	}

	const (
		retention = 10 * time.Second
		interval  = 1 * time.Second
		id        = "torrentid"
	)

	videoBase := t.TempDir()
	torrentBase := t.TempDir()
	videoPath := filepath.Join(videoBase, id)
	torrentPath := filepath.Join(torrentBase, id)
	mustWriteFile(t, videoPath, "stream.m3u8")
	mustWriteFile(t, torrentPath, "source.torrent")

	// The torrent finished "now", so it must survive until the retention window
	// elapses, then be deleted on a subsequent tick.
	store := &fakeStore{finishedAt: map[string]time.Time{id: time.Now()}}
	dropper := &fakeDropper{}

	c := NewCleaner(store, dropper, retention, interval)
	c.videoBasePath = videoBase
	c.torrentBasePath = torrentBase

	go c.Run(context.Background())

	// Well within the retention window: data must still be present.
	time.Sleep(3 * time.Second)
	if _, err := os.Stat(videoPath); err != nil {
		t.Fatalf("video data deleted before retention elapsed: %v", err)
	}
	if store.wasReset(id) {
		t.Fatalf("torrent reset before retention elapsed")
	}

	// Wait past the retention window (plus a tick) for the cleaner to act.
	deadline := time.Now().Add(retention + 3*time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(videoPath); os.IsNotExist(err) {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}

	if _, err := os.Stat(videoPath); !os.IsNotExist(err) {
		t.Fatalf("video data not deleted after retention: err=%v", err)
	}
	if _, err := os.Stat(torrentPath); !os.IsNotExist(err) {
		t.Fatalf("torrent data not deleted after retention: err=%v", err)
	}
	if !store.wasReset(id) {
		t.Fatalf("torrent status not reset after deletion")
	}
	if !dropper.wasDropped(id) {
		t.Fatalf("torrent not dropped from client after deletion")
	}
}

func TestCleanerCleanupInterrupted(t *testing.T) {
	const id = "torrentid"

	videoBase := t.TempDir()
	torrentBase := t.TempDir()
	videoPath := filepath.Join(videoBase, id)
	torrentPath := filepath.Join(torrentBase, id)
	mustWriteFile(t, videoPath, "stream.m3u8")
	mustWriteFile(t, torrentPath, "source.torrent")

	store := &fakeStore{inProgress: map[string]bool{id: true}}
	dropper := &fakeDropper{}

	c := NewCleaner(store, dropper, DefaultRetention, DefaultInterval)
	c.videoBasePath = videoBase
	c.torrentBasePath = torrentBase

	c.CleanupInterrupted(context.Background())

	if _, err := os.Stat(videoPath); !os.IsNotExist(err) {
		t.Fatalf("interrupted video data not deleted: err=%v", err)
	}
	if _, err := os.Stat(torrentPath); !os.IsNotExist(err) {
		t.Fatalf("interrupted torrent data not deleted: err=%v", err)
	}
	if !store.wasReset(id) {
		t.Fatalf("interrupted torrent status not reset")
	}
	if !dropper.wasDropped(id) {
		t.Fatalf("interrupted torrent not dropped from client")
	}
}
