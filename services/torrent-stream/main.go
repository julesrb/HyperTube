package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"hypertube/torrent-stream/internal/stream"
)

func handleStream(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// TODO: start torrent download + ffmpeg pipeline for id
	// TODO: return playlist URL or redirect to /hls/{id}/master.m3u8
	http.Error(w, fmt.Sprintf("stream %s: not implemented", id), http.StatusNotImplemented)
}

func addCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		h.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()

	stream := stream.NewStreamHandler()

    mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })

    mux.HandleFunc("GET /player", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "./player.html")
    })

    mux.HandleFunc("GET /stream/{id}", stream.InitStream) // start torrent and prepapre for trancoding and streaming
	mux.HandleFunc("GET /stream/{id}/index", stream.GetIndex) // serve the HLS index 
	mux.HandleFunc("GET /stream/{id}/{segment}", stream.GetSegment) // serve the HLS segments

    addr := ":" + getEnv("PORT", "8081")
    log.Printf("torrent-stream listening on %s", addr)
    log.Fatal(http.ListenAndServe(addr, mux))
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
