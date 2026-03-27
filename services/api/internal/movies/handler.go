package movies

import "net/http"

// List searches torrent sources and enriches results with TMDb metadata.
func List(w http.ResponseWriter, r *http.Request) {}

// Get returns metadata for a single movie.
func Get(w http.ResponseWriter, r *http.Request) {}

// ListComments returns comments for a movie.
func ListComments(w http.ResponseWriter, r *http.Request) {}

// CreateComment posts a new comment on a movie.
func CreateComment(w http.ResponseWriter, r *http.Request) {}
