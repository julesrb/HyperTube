package main

import (
	"log"
	"net/http"
	"os"

	"peerstream/torrent-stream/internal/stream"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// POST /stream  body: { "magnet": "magnet:?xt=..." }
	// Starts downloading and streams transcoded video back to the caller.
	mux.HandleFunc("POST /stream", stream.Stream)

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
