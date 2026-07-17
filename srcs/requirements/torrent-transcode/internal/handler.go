package torrent_transcode

import (
	"context"
	"hypertube/torrent-transcode/internal/torrent"
	"hypertube/torrent-transcode/internal/transcode"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type TorrentTranscodeHandler struct {
	videoBasePath   string
	torrentBasePath string
	torrentClient   *torrent.Client
	store           *Store
	streams         *map[string]io.ReadSeekCloser // torrentID → reader
	mu              *sync.Mutex
}

func NewTorrentTranscodeHandler(torrentClient *torrent.Client, store *Store, mu *sync.Mutex, streams *map[string]io.ReadSeekCloser) *TorrentTranscodeHandler {
	return &TorrentTranscodeHandler{
		videoBasePath:   "/data/videos",
		torrentBasePath: "/data/torrents",
		torrentClient:   torrentClient,
		store:           store,
		streams:         streams,
		mu:              mu,
	}
}

func (s *TorrentTranscodeHandler) InitDownload(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing torrent id", http.StatusBadRequest)
		return
	}

	torrentPath, err := s.torrentClient.TorrentFilePath(id)
	if err != nil {
		http.Error(w, "torrent file not available", http.StatusInternalServerError)
		log.Printf("torrent file not available: %v, %s", err, id)
		return
	}
	log.Printf("%s: torrent file ready at %s", id, torrentPath)

	IOReader, err := s.torrentClient.Add(id, torrentPath)
	if err != nil {
		http.Error(w, "failed to add torrent", http.StatusInternalServerError)
		log.Printf("failed to add torrent: %v", err)
		return
	}
	log.Printf("%s: torrent download init successfully", id)

	s.mu.Lock()
	(*s.streams)[id] = IOReader
	s.mu.Unlock()
	w.WriteHeader(http.StatusOK)
}

func (s *TorrentTranscodeHandler) InitTranscode(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	videoPath := s.videoBasePath + "/" + id
	originalLang := r.URL.Query().Get("lang")

	// Check if torrent is ready for transcoding
	var torrentReader io.ReadSeekCloser
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		s.mu.Lock()
		torrentReader = (*s.streams)[id]
		s.mu.Unlock()
		if torrentReader != nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if torrentReader == nil {
		http.Error(w, "torrent not ready", http.StatusServiceUnavailable)
		log.Printf("torrent not ready after timeout: %s", id)
		return
	}
	log.Printf("%s: torrent file pipe ready for transcoding", id)

	if err := os.MkdirAll(videoPath, 0755); err != nil {
		http.Error(w, "failed to create stream directory", http.StatusInternalServerError)
		log.Printf("failed to create stream directory: %v", err)
		return
	}

	log.Printf("%s: ConvertHLS starting", id)
	go func() {
		status := "finished"
		if err := transcode.ConvertPipeHLS(torrentReader, videoPath, originalLang); err != nil {
			log.Printf("ConvertPipeHLS failed: %v", err)
			status = "failed"
		}
		if err := s.store.SetTorrentStatus(context.Background(), id, status); err != nil {
			log.Printf("%s: failed to set torrent status to %q: %v", id, status, err)
		}
	}()

	m3u8Path := videoPath + "/stream.m3u8"
	deadline = time.Now().Add(60 * time.Second)
	var m3u8Ready bool
	for time.Now().Before(deadline) {
		if _, err := os.Stat(m3u8Path); err == nil {
			m3u8Ready = true
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if !m3u8Ready {
		http.Error(w, "transcode timed out", http.StatusGatewayTimeout)
		log.Printf("%s: m3u8 not ready after 60s timeout", id)
		return
	}
	log.Printf("%s: transcoding far enough to stream", id)
	w.WriteHeader(http.StatusOK)
}
