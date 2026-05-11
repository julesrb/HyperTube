#!/usr/bin/env bash

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080/api/v1}"
BASE_URL="${BASE_URL%/}"

print_section() {
  echo
  echo "=================================================="
  echo "$1"
  echo "=================================================="
}

pretty_curl() {
  local method="$1"
  local url="$2"
  local data="${3:-}"

  echo
  echo "Request:"
  echo "  $method $url"

  if [[ -n "$data" ]]; then
    echo "Payload:"
    echo "$data" | jq .
  fi

  echo
  echo "Response:"

  if [[ -n "$data" ]]; then
    curl -s -X "$method" \
      "$url" \
      -H "Content-Type: application/json" \
      -d "$data" | jq .
  else
    curl -s -X "$method" "$url" | jq .
  fi
}

print_section "Health Check"

curl -i "$BASE_URL/health"

print_section "List Movies"

pretty_curl "GET" "$BASE_URL/movies"

print_section "Search Movies"

pretty_curl "GET" "$BASE_URL/movies/search?title=gary"

print_section "Get Movie By ID"

MOVIE_ID="693134"

pretty_curl "GET" "$BASE_URL/movies/$MOVIE_ID"

print_section "Get Movie Torrents"

pretty_curl "GET" "$BASE_URL/movies/$MOVIE_ID/torrents"

print_section "Register User"

REGISTER_PAYLOAD='{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123"
}'

pretty_curl "POST" "$BASE_URL/auth/register" "$REGISTER_PAYLOAD"

print_section "Login User"

LOGIN_PAYLOAD='{
  "email": "test@example.com",
  "password": "password123"
}'

pretty_curl "POST" "$BASE_URL/auth/login" "$LOGIN_PAYLOAD"

echo
echo "Done."
