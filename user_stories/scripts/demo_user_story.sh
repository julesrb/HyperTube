#!/usr/bin/env bash
#
# CLI demo walkthrough for the Hypertube API user story.
#
# Usage:
#   ./user_stories/scripts/demo_user_story.sh
#
# Optional environment variables:
#   BASE_URL       API base URL. Default: http://localhost:8080
#   DEMO_EMAIL     Demo user's email. Default: unique email per run.
#   DEMO_USERNAME  Demo user's username. Default: unique username per run.
#   DEMO_PASSWORD  Demo user's password. Default: DemoPass123!
#   SEARCH_QUERY   Movie search query. Default: matrix
#
# Examples:
#   BASE_URL=http://localhost:8080 ./user_stories/scripts/demo_user_story.sh
#   DEMO_EMAIL=demo@example.test DEMO_PASSWORD=DemoPass123! ./user_stories/scripts/demo_user_story.sh
#   SEARCH_QUERY=inception ./user_stories/scripts/demo_user_story.sh

set -o pipefail

TOTAL_STEPS=10
RUN_ID="$(date +%Y%m%d%H%M%S)-$$"

BASE_URL="${BASE_URL:-http://localhost:8080}"
BASE_URL="${BASE_URL%/}"
DEMO_EMAIL="${DEMO_EMAIL:-demo-${RUN_ID}@example.test}"
DEMO_USERNAME="${DEMO_USERNAME:-demo_${RUN_ID//[^A-Za-z0-9_]/_}}"
DEMO_PASSWORD="${DEMO_PASSWORD:-DemoPass123!}"
SEARCH_QUERY="${SEARCH_QUERY:-matrix}"

TOKEN=""
LIST_MOVIE_ID=""
SEARCH_MOVIE_ID=""
MOVIE_ID=""

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
  BLUE="$(tput setaf 4)"
  CYAN="$(tput setaf 6)"
  RESET="$(tput sgr0)"
else
  BOLD=""
  DIM=""
  RED=""
  GREEN=""
  YELLOW=""
  BLUE=""
  CYAN=""
  RESET=""
fi

usage() {
  sed -n '2,18p' "$0" | sed 's/^# \{0,1\}//'
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
  explain "The CLI reports which parts of the end-to-end walkthrough succeeded, failed, or were skipped."

  printf '\n%sResult:%s\n' "$BOLD" "$RESET"
  for i in "${!SUMMARY_NAMES[@]}"; do
    printf '  %-5s %s - %s\n' "$(status_label "${SUMMARY_STATUS[$i]}")" "${SUMMARY_NAMES[$i]}" "${SUMMARY_DETAILS[$i]}"
  done

  printf '\n%sConfiguration:%s\n' "$BOLD" "$RESET"
  printf '  BASE_URL: %s\n' "$BASE_URL"
  printf '  DEMO_EMAIL: %s\n' "$DEMO_EMAIL"
  printf '  DEMO_USERNAME: %s\n' "$DEMO_USERNAME"
  printf '  SEARCH_QUERY: %s\n' "$SEARCH_QUERY"
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

extract_first_movie_id() {
  local body="$1"

  jq -r '
    def collection:
      if (.data | type) == "array" then .data
      elif (.data.movies | type) == "array" then .data.movies
      elif (.movies | type) == "array" then .movies
      elif (.results | type) == "array" then .results
      elif type == "array" then .
      else []
      end;

    collection[0] // {}
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
      first_name: "Demo",
      last_name: "User",
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
explain "The CLI checks whether the backend is reachable before running the rest of the walkthrough."
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

heading 2 "Register Demo User"
explain "The CLI creates a demo user. If that user already exists, the script keeps going and attempts login with the configured password."
REGISTER_URL="$BASE_URL/api/v1/auth/register"
print_request "POST" "$REGISTER_URL" "$REGISTER_PAYLOAD"
request "POST" "$REGISTER_URL" "$REGISTER_PAYLOAD"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$CURL_ERROR"

if is_success_status "$HTTP_STATUS"; then
  record_result "Register demo user" "OK" "Demo user was registered."
  print_result "OK" "Demo user was registered."
elif is_conflict_status "$HTTP_STATUS"; then
  record_result "Register demo user" "WARN" "Demo user already exists; continuing to login."
  print_result "WARN" "Demo user already exists; continuing to login."
else
  record_result "Register demo user" "WARN" "Registration failed; login will verify whether the demo can continue."
  print_result "WARN" "Registration failed; login will verify whether the demo can continue."
fi

heading 3 "Log In"
explain "The CLI signs in as the demo user so protected movie endpoints can be called with a bearer token."
LOGIN_URL="$BASE_URL/api/v1/auth/login"
print_request "POST" "$LOGIN_URL" "$LOGIN_PAYLOAD"
request "POST" "$LOGIN_URL" "$LOGIN_PAYLOAD"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$CURL_ERROR"

if is_success_status "$HTTP_STATUS"; then
  record_result "Log in" "OK" "Login request succeeded."
  print_result "OK" "Login request succeeded."
else
  record_result "Log in" "FAIL" "Login request failed."
  print_result "FAIL" "Login request failed."
  abort_critical "Login failed; authenticated movie requests cannot continue."
fi

heading 4 "Store Auth Token"
explain "The CLI extracts the JWT from the login response and stores it for authenticated requests."
TOKEN="$(extract_token "$HTTP_BODY")"

printf '\n%sResult:%s\n' "$BOLD" "$RESET"
if [[ -n "$TOKEN" ]]; then
  record_result "Store auth token" "OK" "JWT token was found and stored."
  printf '  %s - JWT token was found and stored.\n' "$(status_label OK)"
else
  record_result "Store auth token" "FAIL" "Login response did not include a JWT token."
  printf '  %s - Login response did not include a JWT token.\n' "$(status_label FAIL)"
  abort_critical "No JWT token was available for authenticated requests."
fi

heading 5 "List Movies"
explain "The CLI loads the public movie list and keeps the first movie ID as a fallback for later steps."
MOVIES_URL="$BASE_URL/api/v1/movies"
print_request "GET" "$MOVIES_URL"
request "GET" "$MOVIES_URL"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$CURL_ERROR"

if is_success_status "$HTTP_STATUS"; then
  LIST_MOVIE_ID="$(extract_first_movie_id "$HTTP_BODY")"
  if [[ -n "$LIST_MOVIE_ID" ]]; then
    record_result "List movies" "OK" "Movie list loaded; fallback movie ID is $LIST_MOVIE_ID."
    print_result "OK" "Movie list loaded; fallback movie ID is $LIST_MOVIE_ID."
  else
    record_result "List movies" "WARN" "Movie list loaded, but no movie ID was found."
    print_result "WARN" "Movie list loaded, but no movie ID was found."
  fi
else
  record_result "List movies" "WARN" "Movie list request failed; search may still provide a movie."
  print_result "WARN" "Movie list request failed; search may still provide a movie."
fi

heading 6 "Search Movies"
explain "The CLI searches for movies using the configured query and prefers the first search result for the detail walkthrough."
SEARCH_URL="$BASE_URL/api/v1/movies/search?title=$(urlencode "$SEARCH_QUERY")"
print_request "GET" "$SEARCH_URL" "" "true"
request "GET" "$SEARCH_URL" "" "true"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$CURL_ERROR"

if is_success_status "$HTTP_STATUS"; then
  SEARCH_MOVIE_ID="$(extract_first_movie_id "$HTTP_BODY")"
  if [[ -n "$SEARCH_MOVIE_ID" ]]; then
    record_result "Search movies" "OK" "Search returned movie ID $SEARCH_MOVIE_ID."
    print_result "OK" "Search returned movie ID $SEARCH_MOVIE_ID."
  else
    record_result "Search movies" "WARN" "Search succeeded, but no movie ID was found."
    print_result "WARN" "Search succeeded, but no movie ID was found."
  fi
else
  record_result "Search movies" "WARN" "Search request failed; falling back to the movie list if possible."
  print_result "WARN" "Search request failed; falling back to the movie list if possible."
fi

heading 7 "Select First Movie"
explain "The CLI selects the first movie from search results, or falls back to the first movie from the list."

if [[ -n "$SEARCH_MOVIE_ID" ]]; then
  MOVIE_ID="$SEARCH_MOVIE_ID"
  record_result "Select first movie" "OK" "Selected $MOVIE_ID from search results."
  print_result "OK" "Selected $MOVIE_ID from search results."
elif [[ -n "$LIST_MOVIE_ID" ]]; then
  MOVIE_ID="$LIST_MOVIE_ID"
  record_result "Select first movie" "OK" "Selected $MOVIE_ID from the movie list fallback."
  print_result "OK" "Selected $MOVIE_ID from the movie list fallback."
else
  record_result "Select first movie" "WARN" "No movie ID was available from search results or the movie list."
  print_result "WARN" "No movie ID was available from search results or the movie list."
fi

heading 8 "Fetch Movie Details"
explain "The CLI requests details for the selected movie using the stored JWT token."

if [[ -n "$MOVIE_ID" ]]; then
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
else
  record_result "Fetch movie details" "SKIP" "Skipped because no movie ID was selected."
  print_result "SKIP" "Skipped because no movie ID was selected."
fi

heading 9 "Fetch Torrents"
explain "The CLI requests available torrents for the selected movie using the same movie ID."

if [[ -n "$MOVIE_ID" ]]; then
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
else
  record_result "Fetch torrents" "SKIP" "Skipped because no movie ID was selected."
  print_result "SKIP" "Skipped because no movie ID was selected."
fi

print_summary
