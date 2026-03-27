package auth

import "net/http"

// Token issues a JWT after validating credentials or an OAuth code.
func Token(w http.ResponseWriter, r *http.Request) {}

// Callback42 handles the OAuth2 callback from the 42 intranet.
func Callback42(w http.ResponseWriter, r *http.Request) {}

// CallbackGitHub handles the OAuth2 callback from GitHub.
func CallbackGitHub(w http.ResponseWriter, r *http.Request) {}
