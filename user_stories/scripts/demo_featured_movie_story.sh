#!/usr/bin/env bash
#
# CLI demo walkthrough for a second Hypertube API user story.
#
# User story:
#   As a visitor, I want to browse featured movies first, then sign in when
#   details and torrents require authentication.
#
# Usage:
#   ./user_stories/scripts/demo_featured_movie_story.sh
#
# Optional environment variables:
#   BASE_URL        API base URL. Default: http://localhost:8080
#   DEMO_EMAIL      Demo user's email. Default: unique email per run.
#   DEMO_USERNAME   Demo user's username. Default: unique username per run.
#   DEMO_PASSWORD   Demo user's password. Default: DemoPass123!
#   FEATURED_INDEX  Zero-based movie index from the featured list. Default: 0
#
# Examples:
#   BASE_URL=http://localhost:8080 ./user_stories/scripts/demo_featured_movie_story.sh
#   FEATURED_INDEX=1 ./user_stories/scripts/demo_featured_movie_story.sh
#   DEMO_EMAIL=demo@example.test DEMO_PASSWORD=DemoPass123! ./user_stories/scripts/demo_featured_movie_story.sh

set -o pipefail

TOTAL_STEPS=9
RUN_ID="$(date +%Y%m%d%H%M%S)-$$"

BASE_URL="${BASE_URL:-http://localhost:8080}"
BASE_URL="${BASE_URL%/}"
DEMO_EMAIL="${DEMO_EMAIL:-featured-demo-${RUN_ID}@example.test}"
DEMO_USERNAME="${DEMO_USERNAME:-featured_${RUN_ID//[^A-Za-z0-9_]/_}}"
DEMO_PASSWORD="${DEMO_PASSWORD:-DemoPass123!}"
FEATURED_INDEX="${FEATURED_INDEX:-0}"

TOKEN=""
MOVIE_ID=""
MOVIE_TITLE=""

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
  sed -n '2,22p' "$0" | sed 's/^# \{0,1\}//'
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

  heading 9 "Summary"
  explain "The CLI reports which parts of the featured movie walkthrough succeeded, failed, or were skipped."

  printf '\n%sResult:%s\n' "$BOLD" "$RESET"
  for i in "${!SUMMARY_NAMES[@]}"; do
    printf '  %-5s %s - %s\n' "$(status_label "${SUMMARY_STATUS[$i]}")" "${SUMMARY_NAMES[$i]}" "${SUMMARY_DETAILS[$i]}"
  done

  printf '\n%sConfiguration:%s\n' "$BOLD" "$RESET"
  printf '  BASE_URL: %s\n' "$BASE_URL"
  printf '  DEMO_EMAIL: %s\n' "$DEMO_EMAIL"
  printf '  DEMO_USERNAME: %s\n' "$DEMO_USERNAME"
  printf '  FEATURED_INDEX: %s\n' "$FEATURED_INDEX"
}

abort_critical() {
  local message="$1"
  printf '\n%sCritical failure:%s %s\n' "$RED" "$RESET" "$message" >&2
  print_summary
  exit 1
}

is_success_status() {
  [[ "$1" =~ ^2[0-9][0-9]$ ]]
}

is_auth_challenge_status() {
  [[ "$1" == "401" || "$1" == "403" ]]
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

extract_movie_field() {
  local body="$1"
  local index="$2"
  local field="$3"

  jq -r \
    --argjson index "$index" \
    --arg field "$field" \
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
      | if $field == "id" then
          .imdb_id // .imdbID // .imdbId // .id // empty
        elif $field == "title" then
          .title // .name // empty
        else
          empty
        end
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
      first_name: "Featured",
      last_name: "Viewer",
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
explain "The CLI checks whether the backend is reachable before browsing the featured catalog."
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

heading 2 "Browse Featured Movies"
explain "The CLI loads the public featured movie list without sending any auth token."
MOVIES_URL="$BASE_URL/api/v1/movies"
print_request "GET" "$MOVIES_URL"
request "GET" "$MOVIES_URL"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$CURL_ERROR"

if is_success_status "$HTTP_STATUS"; then
  MOVIE_ID="$(extract_movie_field "$HTTP_BODY" "$FEATURED_INDEX" "id")"
  MOVIE_TITLE="$(extract_movie_field "$HTTP_BODY" "$FEATURED_INDEX" "title")"

  if [[ -n "$MOVIE_ID" ]]; then
    record_result "Browse featured movies" "OK" "Selected candidate movie ID $MOVIE_ID from the public list."
    print_result "OK" "Selected candidate movie ID $MOVIE_ID from the public list."
  else
    record_result "Browse featured movies" "FAIL" "No movie ID found at FEATURED_INDEX=$FEATURED_INDEX."
    print_result "FAIL" "No movie ID found at FEATURED_INDEX=$FEATURED_INDEX."
    abort_critical "The featured list did not provide a movie ID for the configured index."
  fi
else
  record_result "Browse featured movies" "FAIL" "Featured movie list request failed."
  print_result "FAIL" "Featured movie list request failed."
  abort_critical "The featured movie list is required for this story."
fi

heading 3 "Choose Featured Movie"
explain "The CLI confirms which featured movie will be used for the protected detail and torrent requests."

if [[ -n "$MOVIE_TITLE" ]]; then
  record_result "Choose featured movie" "OK" "Using $MOVIE_TITLE ($MOVIE_ID)."
  print_result "OK" "Using $MOVIE_TITLE ($MOVIE_ID)."
else
  record_result "Choose featured movie" "OK" "Using movie ID $MOVIE_ID."
  print_result "OK" "Using movie ID $MOVIE_ID."
fi

heading 4 "Try Details Before Login"
explain "The CLI intentionally requests protected movie details without a token to show where sign-in is required."
DETAILS_URL="$BASE_URL/api/v1/movies/$MOVIE_ID"
print_request "GET" "$DETAILS_URL"
request "GET" "$DETAILS_URL"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$CURL_ERROR"

if is_auth_challenge_status "$HTTP_STATUS"; then
  record_result "Try details before login" "OK" "API required authentication as expected."
  print_result "OK" "API required authentication as expected."
elif is_success_status "$HTTP_STATUS"; then
  record_result "Try details before login" "WARN" "Movie details were available without authentication."
  print_result "WARN" "Movie details were available without authentication."
else
  record_result "Try details before login" "WARN" "Unauthenticated detail request returned HTTP $HTTP_STATUS."
  print_result "WARN" "Unauthenticated detail request returned HTTP $HTTP_STATUS."
fi

heading 5 "Ensure Demo Account"
explain "The CLI creates a demo account for sign-in. If it already exists, the script keeps going and logs in."
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

heading 6 "Log In And Store Token"
explain "The CLI logs in and extracts the JWT so protected movie endpoints can be called."
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

heading 7 "Fetch Authenticated Details"
explain "The CLI repeats the movie detail request with the stored bearer token."
print_request "GET" "$DETAILS_URL" "" "true"
request "GET" "$DETAILS_URL" "" "true"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$CURL_ERROR"

if is_success_status "$HTTP_STATUS"; then
  record_result "Fetch authenticated details" "OK" "Movie details loaded for $MOVIE_ID."
  print_result "OK" "Movie details loaded for $MOVIE_ID."
else
  record_result "Fetch authenticated details" "WARN" "Authenticated detail request failed for $MOVIE_ID."
  print_result "WARN" "Authenticated detail request failed for $MOVIE_ID."
fi

heading 8 "Fetch Movie Torrents"
explain "The CLI requests available torrents for the same featured movie using the stored token."
TORRENTS_URL="$BASE_URL/api/v1/movies/$MOVIE_ID/torrents"
print_request "GET" "$TORRENTS_URL" "" "true"
request "GET" "$TORRENTS_URL" "" "true"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$CURL_ERROR"

if is_success_status "$HTTP_STATUS"; then
  record_result "Fetch movie torrents" "OK" "Torrent list loaded for $MOVIE_ID."
  print_result "OK" "Torrent list loaded for $MOVIE_ID."
else
  record_result "Fetch movie torrents" "WARN" "Torrent request failed or no torrents were available for $MOVIE_ID."
  print_result "WARN" "Torrent request failed or no torrents were available for $MOVIE_ID."
fi

print_summary
