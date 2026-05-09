package stream

import (
	"hypertube/torrent-stream/internal/transcode"
	"log"
	"net/http"
	"os"
)

type StreamHandler struct{}

func NewStreamHandler() *StreamHandler {
	return &StreamHandler{}
}

func (s *StreamHandler) InitStream(w http.ResponseWriter, r *http.Request) {
	// id := r.PathValue("id")
	os.MkdirAll("./internal/transcode/test/output/", 0755)
	err := transcode.TranscodeHLS("./internal/transcode/test/rubber.mp4", "./internal/transcode/test/output/index.m3u8")
	if err != nil {
		http.Error(w, "failed to start stream", http.StatusInternalServerError)
		log.Printf("failed to start stream: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
}

func (s *StreamHandler) GetIndex(w http.ResponseWriter, r *http.Request) {
	prefix := "./internal/transcode/test/output/"
	// id := r.PathValue("id")
	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	if bytes, err := os.ReadFile(prefix + "index.m3u8"); err != nil {
		http.Error(w, "failed to read index file", http.StatusInternalServerError)
		return
	} else {
		w.WriteHeader(200)
		w.Write(bytes)
	}
}

func (s *StreamHandler) GetSegment(w http.ResponseWriter, r *http.Request) {
	prefix := "./internal/transcode/test/output/"
	// id := r.PathValue("id")
	segment := r.PathValue("segment")
	w.Header().Set("Content-Type", "video/mp2t")
	if bytes, err := os.ReadFile(prefix + segment); err != nil {
		http.Error(w, "failed to read segment file", http.StatusInternalServerError)
		return
	} else {
		w.WriteHeader(200)
		w.Write(bytes)
	}
}
