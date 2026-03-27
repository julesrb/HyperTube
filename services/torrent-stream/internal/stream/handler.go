package stream

import "net/http"

// Stream starts the torrent download (if not cached), pipes it through ffmpeg,
// and writes the fragmented MP4 to the HTTP response.
func Stream(w http.ResponseWriter, r *http.Request) {}
