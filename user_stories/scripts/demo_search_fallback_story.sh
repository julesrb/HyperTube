#!/usr/bin/env bash
#
# CLI demo walkthrough for a third Hypertube API user story.
#
# User story:
#   As a signed-in user, I want to search for a specific movie, recover from a
#   weak or empty search result, and still reach movie details and torrents.
#
# Usage:
#   ./user_stories/scripts/demo_search_fallback_story.sh
#
# Optional environment variables:
#   BASE_URL        API base URL. Default: http://localhost:8080
#   DEMO_EMAIL      Demo user's email. Default: unique email per run.
#   DEMO_USERNAME   Demo user's username. Default: unique username per run.
#   DEMO_PASSWORD   Demo user's password. Default: DemoPass123!
#   SEARCH_QUERY    Primary movie search query. Default: matrx
#   BACKUP_QUERY    Backup movie search query. Default: matrix
#   FEATURED_INDEX  Zero-based fallback movie index from /movies. Default: 0
#
# Examples:
#   BASE_URL=http://localhost:8080 ./user_stories/scripts/demo_search_fallback_story.sh
#   SEARCH_QUERY=interstelar BACKUP_QUERY=interstellar ./user_stories/scripts/demo_search_fallback_story.sh
#   FEATURED_INDEX=2 ./user_stories/scripts/demo_search_fallback_story.sh

set -o pipefail

TOTAL_STEPS=10
RUN_ID="$(date +%Y%m%d%H%M%S)-$$"

BASE_URL="${BASE_URL:-http://localhost:8080}"
BASE_URL="${BASE_URL%/}"
DEMO_EMAIL="${DEMO_EMAIL:-fallback-demo-${RUN_ID}@example.test}"
DEMO_USERNAME="${DEMO_USERNAME:-fallback_${RUN_ID//[^A-Za-z0-9_]/_}}"
DEMO_PASSWORD="${DEMO_PASSWORD:-DemoPass123!}"
SEARCH_QUERY="${SEARCH_QUERY:-matrx}"
BACKUP_QUERY="${BACKUP_QUERY:-matrix}"
FEATURED_INDEX="${FEATURED_INDEX:-0}"

TOKEN=""
PRIMARY_MOVIE_ID=""
BACKUP_MOVIE_ID=""
FEATURED_MOVIE_ID=""
MOVIE_ID=""
MOVIE_SOURCE=""

TMP_FILES=()
SUMMARY_NAMES=()
SUMMARY_STATUS=()
SUMMARY_DETAILS=()

cleanup() {
  local file
  for file in "${TMP_FILES[@]}"; do
    [[ -n "$file" && -f "$file" ]] && rm -f "$file"
  done
}
trap cleanup EXIT

if [[ -t 1 ]] && command -v tput >/dev/null 2>&1 && [[ "$(tput colors 2>/dev/null || echo 0)" -ge 8 ]]; then
  BOLD="$(tput bold)"
  DIM="$(tput dim)"
  RED="$(tput setaf 1)"
  GREEN="$(tput setaf 2)"
  YELLOW="$(tput setaf 3)"
  CYAN="$(tput setaf 6)"
  RESET="$(tput sgr0)"
else
  BOLD=""
  DIM=""
  RED=""
  GREEN=""
  YELLOW=""
  CYAN=""
  RESET=""
fi

usage() {
  sed -n '2,24p' "$0" | sed 's/^# \{0,1\}//'
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

require_command() {
  local name="$1"
  if ! command -v "$name" >/dev/null 2>&1; then
    printf '%sMissing required command:%s %s\n' "$RED" "$RESET" "$name" >&2
    printf 'Install %s and run this script again.\n' "$name" >&2
    exit 1
  fi
}

require_command curl
require_command jq

if ! [[ "$FEATURED_INDEX" =~ ^[0-9]+$ ]]; then
  printf '%sInvalid FEATURED_INDEX:%s %s\n' "$RED" "$RESET" "$FEATURED_INDEX" >&2
  printf 'FEATURED_INDEX must be a zero-based non-negative integer.\n' >&2
  exit 1
fi

heading() {
  local number="$1"
  local title="$2"
  printf '\n%s[%s/%s] %s%s\n' "$BOLD$CYAN" "$number" "$TOTAL_STEPS" "$title" "$RESET"
}

explain() {
  printf '%sWhat happens:%s\n' "$BOLD" "$RESET"
  printf '  %s\n' "$1"
}

record_result() {
  local name="$1"
  local status="$2"
  local detail="$3"

  SUMMARY_NAMES+=("$name")
  SUMMARY_STATUS+=("$status")
  SUMMARY_DETAILS+=("$detail")
}

status_label() {
  case "$1" in
    OK) printf '%sOK%s' "$GREEN" "$RESET" ;;
    WARN) printf '%sWARN%s' "$YELLOW" "$RESET" ;;
    FAIL) printf '%sFAIL%s' "$RED" "$RESET" ;;
    SKIP) printf '%sSKIP%s' "$DIM" "$RESET" ;;
    *) printf '%s' "$1" ;;
  esac
}

print_summary() {
  local i

  heading 10 "Summary"
  explain "The CLI reports which parts of the search fallback walkthrough succeeded, failed, or were skipped."

  printf '\n%sResult:%s\n' "$BOLD" "$RESET"
  for i in "${!SUMMARY_NAMES[@]}"; do
    printf '  %-5s %s - %s\n' "$(status_label "${SUMMARY_STATUS[$i]}")" "${SUMMARY_NAMES[$i]}" "${SUMMARY_DETAILS[$i]}"
  done

  printf '\n%sConfiguration:%s\n' "$BOLD" "$RESET"
  printf '  BASE_URL: %s\n' "$BASE_URL"
  printf '  DEMO_EMAIL: %s\n' "$DEMO_EMAIL"
  printf '  DEMO_USERNAME: %s\n' "$DEMO_USERNAME"
  printf '  SEARCH_QUERY: %s\n' "$SEARCH_QUERY"
  printf '  BACKUP_QUERY: %s\n' "$BACKUP_QUERY"
  printf '  FEATURED_INDEX: %s\n' "$FEATURED_INDEX"
}

abort_critical() {
  local message="$1"
  printf '\n%sCritical failure:%s %s\n' "$RED" "$RESET" "$message" >&2
  print_summary
  exit 1
}

urlencode() {
  jq -rn --arg value "$1" '$value | @uri'
}

is_success_status() {
  [[ "$1" =~ ^2[0-9][0-9]$ ]]
}

is_conflict_status() {
  [[ "$1" == "409" ]]
}

pretty_json_or_raw() {
  local body="$1"

  if [[ -z "$body" ]]; then
    printf '  %s(empty response body)%s\n' "$DIM" "$RESET"
    return 0
  fi

  if jq -e . >/dev/null 2>&1 <<<"$body"; then
    jq . <<<"$body" | sed 's/^/  /'
  else
    printf '%s\n' "$body" | sed 's/^/  /'
  fi
}

print_request() {
  local method="$1"
  local url="$2"
  local payload="${3:-}"
  local authenticated="${4:-false}"

  printf '\n%sRequest:%s\n' "$BOLD" "$RESET"
  printf '  %s %s\n' "$method" "$url"

  if [[ "$authenticated" == "true" ]]; then
    printf '  Authorization: Bearer %s\n' "${TOKEN:+<stored JWT>}"
  fi

  if [[ -n "$payload" ]]; then
    printf '\n%sPayload:%s\n' "$BOLD" "$RESET"
    jq . <<<"$payload" | sed 's/^/  /'
  fi
}

print_response() {
  local status="$1"
  local body="$2"
  local curl_error="${3:-}"

  printf '\n%sResponse:%s\n' "$BOLD" "$RESET"
  printf '  HTTP %s\n' "$status"

  if [[ -n "$curl_error" ]]; then
    printf '  curl error: %s\n' "$curl_error"
  fi

  pretty_json_or_raw "$body"
}

print_result() {
  local status="$1"
  local detail="$2"

  printf '\n%sResult:%s\n' "$BOLD" "$RESET"
  printf '  %s - %s\n' "$(status_label "$status")" "$detail"
}

request() {
  local method="$1"
  local url="$2"
  local payload="${3:-}"
  local authenticated="${4:-false}"
  local body_file
  local error_file
  local curl_args

  HTTP_STATUS=""
  HTTP_BODY=""
  CURL_ERROR=""

  body_file="$(mktemp)"
  error_file="$(mktemp)"
  TMP_FILES+=("$body_file" "$error_file")

  curl_args=(
    -sS
    -o "$body_file"
    -w "%{http_code}"
    -X "$method"
    -H "Accept: application/json"
  )

  if [[ -n "$payload" ]]; then
    curl_args+=(-H "Content-Type: application/json" --data "$payload")
  fi

  if [[ "$authenticated" == "true" && -n "$TOKEN" ]]; then
    curl_args+=(-H "Authorization: Bearer $TOKEN")
  fi

  if ! HTTP_STATUS="$(curl "${curl_args[@]}" "$url" 2>"$error_file")"; then
    CURL_ERROR="$(<"$error_file")"
    HTTP_STATUS="${HTTP_STATUS:-000}"
  fi

  HTTP_BODY="$(<"$body_file")"
}

extract_token() {
  local body="$1"

  jq -r '
    .data.access_token
    // .access_token
    // .data.token
    // .token
    // empty
  ' <<<"$body" 2>/dev/null
}

extract_movie_id() {
  local body="$1"
  local index="${2:-0}"

  jq -r \
    --argjson index "$index" \
    '
      def collection:
        if (.data | type) == "array" then .data
        elif (.data.movies | type) == "array" then .data.movies
        elif (.movies | type) == "array" then .movies
        elif (.results | type) == "array" then .results
        elif type == "array" then .
        else []
        end;

      collection[$index] // {}
      | .imdb_id // .imdbID // .imdbId // .id // empty
    ' <<<"$body" 2>/dev/null
}

REGISTER_PAYLOAD="$(
  jq -n \
    --arg email "$DEMO_EMAIL" \
    --arg username "$DEMO_USERNAME" \
    --arg password "$DEMO_PASSWORD" \
    '{
      email: $email,
      username: $username,
      first_name: "Fallback",
      last_name: "Searcher",
      password: $password
    }'
)"

LOGIN_PAYLOAD="$(
  jq -n \
    --arg email "$DEMO_EMAIL" \
    --arg password "$DEMO_PASSWORD" \
    '{email: $email, password: $password}'
)"

heading 1 "Health Check"
explain "The CLI checks whether the backend is reachable before attempting authenticated discovery."
HEALTH_URL="$BASE_URL/api/v1/health"
print_request "GET" "$HEALTH_URL"
request "GET" "$HEALTH_URL"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$CURL_ERROR"

if is_success_status "$HTTP_STATUS"; then
  record_result "Health check" "OK" "Backend is reachable."
  print_result "OK" "Backend is reachable."
else
  record_result "Health check" "FAIL" "Backend did not return a successful health response."
  print_result "FAIL" "Backend did not return a successful health response."
  abort_critical "Health check failed; the API must be running before the demo can continue."
fi

heading 2 "Ensure Demo Account"
explain "The CLI creates a demo account. If it already exists, the script keeps going and verifies access by logging in."
REGISTER_URL="$BASE_URL/api/v1/auth/register"
print_request "POST" "$REGISTER_URL" "$REGISTER_PAYLOAD"
request "POST" "$REGISTER_URL" "$REGISTER_PAYLOAD"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$CURL_ERROR"

if is_success_status "$HTTP_STATUS"; then
  record_result "Ensure demo account" "OK" "Demo account was registered."
  print_result "OK" "Demo account was registered."
elif is_conflict_status "$HTTP_STATUS"; then
  record_result "Ensure demo account" "WARN" "Demo account already exists; continuing to login."
  print_result "WARN" "Demo account already exists; continuing to login."
else
  record_result "Ensure demo account" "WARN" "Registration failed; login will verify whether the account is usable."
  print_result "WARN" "Registration failed; login will verify whether the account is usable."
fi

heading 3 "Log In And Store Token"
explain "The CLI logs in and extracts the JWT so search, details, and torrents can use bearer authentication."
LOGIN_URL="$BASE_URL/api/v1/auth/login"
print_request "POST" "$LOGIN_URL" "$LOGIN_PAYLOAD"
request "POST" "$LOGIN_URL" "$LOGIN_PAYLOAD"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$CURL_ERROR"

if is_success_status "$HTTP_STATUS"; then
  TOKEN="$(extract_token "$HTTP_BODY")"
  if [[ -n "$TOKEN" ]]; then
    record_result "Log in and store token" "OK" "Login succeeded and JWT token was stored."
    print_result "OK" "Login succeeded and JWT token was stored."
  else
    record_result "Log in and store token" "FAIL" "Login response did not include a JWT token."
    print_result "FAIL" "Login response did not include a JWT token."
    abort_critical "No JWT token was available for authenticated requests."
  fi
else
  record_result "Log in and store token" "FAIL" "Login request failed."
  print_result "FAIL" "Login request failed."
  abort_critical "Login failed; authenticated movie requests cannot continue."
fi

heading 4 "Primary Search"
explain "The CLI searches for the user's first query and tries to select the first movie from those results."
PRIMARY_URL="$BASE_URL/api/v1/movies/search?title=$(urlencode "$SEARCH_QUERY")"
print_request "GET" "$PRIMARY_URL" "" "true"
request "GET" "$PRIMARY_URL" "" "true"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$CURL_ERROR"

if is_success_status "$HTTP_STATUS"; then
  PRIMARY_MOVIE_ID="$(extract_movie_id "$HTTP_BODY")"
  if [[ -n "$PRIMARY_MOVIE_ID" ]]; then
    record_result "Primary search" "OK" "Primary search returned movie ID $PRIMARY_MOVIE_ID."
    print_result "OK" "Primary search returned movie ID $PRIMARY_MOVIE_ID."
  else
    record_result "Primary search" "WARN" "Primary search succeeded, but returned no usable movie ID."
    print_result "WARN" "Primary search succeeded, but returned no usable movie ID."
  fi
else
  record_result "Primary search" "WARN" "Primary search failed with HTTP $HTTP_STATUS."
  print_result "WARN" "Primary search failed with HTTP $HTTP_STATUS."
fi

heading 5 "Backup Search"
explain "If the first query did not produce a movie, the CLI retries with a cleaner backup query."

if [[ -z "$PRIMARY_MOVIE_ID" ]]; then
  BACKUP_URL="$BASE_URL/api/v1/movies/search?title=$(urlencode "$BACKUP_QUERY")"
  print_request "GET" "$BACKUP_URL" "" "true"
  request "GET" "$BACKUP_URL" "" "true"
  print_response "$HTTP_STATUS" "$HTTP_BODY" "$CURL_ERROR"

  if is_success_status "$HTTP_STATUS"; then
    BACKUP_MOVIE_ID="$(extract_movie_id "$HTTP_BODY")"
    if [[ -n "$BACKUP_MOVIE_ID" ]]; then
      record_result "Backup search" "OK" "Backup search returned movie ID $BACKUP_MOVIE_ID."
      print_result "OK" "Backup search returned movie ID $BACKUP_MOVIE_ID."
    else
      record_result "Backup search" "WARN" "Backup search succeeded, but returned no usable movie ID."
      print_result "WARN" "Backup search succeeded, but returned no usable movie ID."
    fi
  else
    record_result "Backup search" "WARN" "Backup search failed with HTTP $HTTP_STATUS."
    print_result "WARN" "Backup search failed with HTTP $HTTP_STATUS."
  fi
else
  record_result "Backup search" "SKIP" "Skipped because the primary search already returned a movie."
  print_result "SKIP" "Skipped because the primary search already returned a movie."
fi

heading 6 "Featured Fallback"
explain "If both searches miss, the CLI falls back to the public featured movie list so the user can still inspect available content."

if [[ -z "$PRIMARY_MOVIE_ID" && -z "$BACKUP_MOVIE_ID" ]]; then
  FEATURED_URL="$BASE_URL/api/v1/movies"
  print_request "GET" "$FEATURED_URL"
  request "GET" "$FEATURED_URL"
  print_response "$HTTP_STATUS" "$HTTP_BODY" "$CURL_ERROR"

  if is_success_status "$HTTP_STATUS"; then
    FEATURED_MOVIE_ID="$(extract_movie_id "$HTTP_BODY" "$FEATURED_INDEX")"
    if [[ -n "$FEATURED_MOVIE_ID" ]]; then
      record_result "Featured fallback" "OK" "Featured fallback returned movie ID $FEATURED_MOVIE_ID."
      print_result "OK" "Featured fallback returned movie ID $FEATURED_MOVIE_ID."
    else
      record_result "Featured fallback" "WARN" "Featured list loaded, but no movie ID was found at index $FEATURED_INDEX."
      print_result "WARN" "Featured list loaded, but no movie ID was found at index $FEATURED_INDEX."
    fi
  else
    record_result "Featured fallback" "WARN" "Featured fallback failed with HTTP $HTTP_STATUS."
    print_result "WARN" "Featured fallback failed with HTTP $HTTP_STATUS."
  fi
else
  record_result "Featured fallback" "SKIP" "Skipped because a search result was already available."
  print_result "SKIP" "Skipped because a search result was already available."
fi

heading 7 "Select Movie"
explain "The CLI chooses the best available movie ID in order: primary search, backup search, then featured fallback."

if [[ -n "$PRIMARY_MOVIE_ID" ]]; then
  MOVIE_ID="$PRIMARY_MOVIE_ID"
  MOVIE_SOURCE="primary search"
elif [[ -n "$BACKUP_MOVIE_ID" ]]; then
  MOVIE_ID="$BACKUP_MOVIE_ID"
  MOVIE_SOURCE="backup search"
elif [[ -n "$FEATURED_MOVIE_ID" ]]; then
  MOVIE_ID="$FEATURED_MOVIE_ID"
  MOVIE_SOURCE="featured fallback"
fi

if [[ -n "$MOVIE_ID" ]]; then
  record_result "Select movie" "OK" "Selected $MOVIE_ID from $MOVIE_SOURCE."
  print_result "OK" "Selected $MOVIE_ID from $MOVIE_SOURCE."
else
  record_result "Select movie" "FAIL" "No movie ID was available from any discovery path."
  print_result "FAIL" "No movie ID was available from any discovery path."
  abort_critical "No movie could be selected for details or torrents."
fi

heading 8 "Fetch Movie Details"
explain "The CLI fetches details for the selected movie with the stored bearer token."
DETAILS_URL="$BASE_URL/api/v1/movies/$MOVIE_ID"
print_request "GET" "$DETAILS_URL" "" "true"
request "GET" "$DETAILS_URL" "" "true"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$CURL_ERROR"

if is_success_status "$HTTP_STATUS"; then
  record_result "Fetch movie details" "OK" "Movie details loaded for $MOVIE_ID."
  print_result "OK" "Movie details loaded for $MOVIE_ID."
else
  record_result "Fetch movie details" "WARN" "Movie detail request failed for $MOVIE_ID."
  print_result "WARN" "Movie detail request failed for $MOVIE_ID."
fi

heading 9 "Fetch Torrents"
explain "The CLI fetches available torrents for the selected movie with the stored bearer token."
TORRENTS_URL="$BASE_URL/api/v1/movies/$MOVIE_ID/torrents"
print_request "GET" "$TORRENTS_URL" "" "true"
request "GET" "$TORRENTS_URL" "" "true"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$CURL_ERROR"

if is_success_status "$HTTP_STATUS"; then
  record_result "Fetch torrents" "OK" "Torrent list loaded for $MOVIE_ID."
  print_result "OK" "Torrent list loaded for $MOVIE_ID."
else
  record_result "Fetch torrents" "WARN" "Torrent request failed or no torrents were available for $MOVIE_ID."
  print_result "WARN" "Torrent request failed or no torrents were available for $MOVIE_ID."
fi

print_summary
