package main

import (
	"context"
	torrent_transcode "hypertube/torrent-transcode/internal"
	"hypertube/torrent-transcode/internal/cleanup"
	"hypertube/torrent-transcode/internal/torrent"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	db := connectDB(ctx)
	defer db.Close()

	store := torrent_transcode.NewStore(db)

	mux := http.NewServeMux()
	streams := make(map[string]io.ReadSeekCloser)
	mu := &sync.Mutex{}
	torrentClient, err := torrent.NewClient(mu, &streams)
	if err != nil {
		log.Fatal("failed to create torrent client:", err)
	}
	torrentTranscode := torrent_transcode.NewTorrentTranscodeHandler(torrentClient, store, mu, &streams)

	cleaner := cleanup.NewCleaner(store, torrentClient, cleanup.DefaultRetention, cleanup.DefaultInterval)
	cleaner.CleanupInterrupted(ctx)
	go cleaner.Run(ctx)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("POST /download/{id}", torrentTranscode.InitDownload)   // start torrent and prepapre for trancoding and streaming
	mux.HandleFunc("POST /transcode/{id}", torrentTranscode.InitTranscode) // start transcoding for an already downloaded torrent

	addr := ":" + getEnv("PORT", "8081")
	log.Printf("torrent-stream listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func connectDB(ctx context.Context) *pgxpool.Pool {
	db, err := pgxpool.New(ctx, getEnv("DATABASE_URL", "postgres://hypertube:changeme@localhost:5432/hypertube?sslmode=disable"))
	if err != nil {
		log.Fatalf("connect to db: %v", err)
	}
	if err := db.Ping(ctx); err != nil {
		log.Fatalf("ping db: %v", err)
	}
	log.Println("connected to database")
	return db
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
