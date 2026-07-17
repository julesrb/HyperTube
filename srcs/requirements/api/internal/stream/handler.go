package stream

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"

	"hypertube/api/internal/downloader"
)

// torrentStore resolves a torrent's metadata by its id.
type torrentStore interface {
	GetTorrent(ctx context.Context, id string) (Torrent, error)
	GetTorrentStatus(ctx context.Context, id string) (string, error)
	SetTorrentStatus(ctx context.Context, id, status string) error
}

type StreamHandler struct {
	videoBasePath   string
	torrentBasePath string
	transcodeURL    string
	downloader      *downloader.Downloader
	store           torrentStore
}

func NewStreamHandler(store torrentStore) *StreamHandler {
	torrentBasePath := "/data/torrents"
	return &StreamHandler{
		videoBasePath:   "/data/videos",
		torrentBasePath: torrentBasePath,
		transcodeURL:    "http://vpn:8081",
		downloader:      downloader.New(torrentBasePath),
		store:           store,
	}
}

func (s *StreamHandler) InitStream(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing torrent id", http.StatusBadRequest)
		return
	}

	if status, err := s.store.GetTorrentStatus(r.Context(), id); err == nil && (status == "in_progress" || status == "finished") {
		log.Printf("%s: torrent already %s, skipping download init", id, status)
		w.WriteHeader(http.StatusOK)
		return
	}

	torrent, err := s.store.GetTorrent(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "torrent not found", http.StatusNotFound)
		} else {
			http.Error(w, "db torrent error", http.StatusInternalServerError)
		}
		log.Printf("lookup torrent for %s: %v", id, err)
		return
	}
	log.Printf("%q: initialising stream for %s", torrent.Title, id)

	if _, err := s.downloader.DownloadTorrentFile(id, torrent.URL); err != nil {
		if errors.Is(err, downloader.ErrSourceTimeout) {
			http.Error(w, "torrent source timed out", http.StatusGatewayTimeout)
		} else {
			http.Error(w, "failed to download torrent source file", http.StatusBadGateway)
		}
		log.Printf("%q : %s downloaded torrent file: %v", torrent.Title, id, err)
		return
	}

	// Init the download (peer-to-peer, behind the VPN)
	respDownload, err := http.Post(s.transcodeURL+"/download/"+id, "application/json", nil)
	if err != nil || respDownload.StatusCode != http.StatusOK {
		http.Error(w, "failed to start stream", http.StatusInternalServerError)
		log.Printf("download service error for %s: %v", id, err)
		return
	}
	defer respDownload.Body.Close()

	// Init transcoding — delegate to the torrent-transcode service and wait for an OK;
	transcodeURL := s.transcodeURL + "/transcode/" + id
	if torrent.OriginalLanguage != "" {
		transcodeURL += "?lang=" + url.QueryEscape(torrent.OriginalLanguage)
	}
	respTranscode, err := http.Post(transcodeURL, "application/json", nil)
	if err != nil || respTranscode.StatusCode != http.StatusOK {
		http.Error(w, "failed to start stream", http.StatusInternalServerError)
		log.Printf("transcode service error for %s: %v", id, err)
		return
	}
	defer respTranscode.Body.Close()

	if err := s.store.SetTorrentStatus(r.Context(), id, "in_progress"); err != nil {
		log.Printf("%s: failed to set torrent status to in_progress: %v", id, err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (s *StreamHandler) GetIndex(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing torrent id", http.StatusBadRequest)
		return
	}
	videoPath := s.videoBasePath + "/" + id
	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	if bytes, err := os.ReadFile(videoPath + "/stream.m3u8"); err != nil {
		http.Error(w, "failed to read index file", http.StatusInternalServerError)
		log.Printf("failed to read index for %s: %v", id, err)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	}
}

func (s *StreamHandler) GetSegment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing torrent id", http.StatusBadRequest)
		return
	}
	prefix := s.videoBasePath + "/" + id
	segment := r.PathValue("segment")
	w.Header().Set("Content-Type", "video/mp2t")
	if bytes, err := os.ReadFile(prefix + "/" + segment); err != nil {
		http.Error(w, "failed to read segment file", http.StatusInternalServerError)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	}
}
