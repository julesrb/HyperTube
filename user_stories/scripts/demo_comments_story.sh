#!/usr/bin/env bash
#
# CLI demo walkthrough for the Hypertube comments user story.
#
# User story:
#   As a signed-in user, I want to comment on a movie, edit my comment, and
#   delete it again, so I can take part in a movie discussion while controlling
#   my own contribution. The backend must derive comment ownership from the JWT,
#   not from a user_id sent by the client.
#
# Usage:
#   ./user_stories/scripts/demo_comments_story.sh
#
# Optional environment variables:
#   BASE_URL       API base URL. Default: http://localhost:8080
#   DEMO_EMAIL     Demo user's email. Default: unique email per run.
#   DEMO_USERNAME  Demo user's username. Default: unique username per run.
#   DEMO_PASSWORD  Demo user's password. Default: DemoPass123!
#   MOVIE_ID       Movie IMDB ID to comment on. Default: first movie from list.
#
# Examples:
#   BASE_URL=http://localhost:8080 ./user_stories/scripts/demo_comments_story.sh
#   MOVIE_ID=tt0133093 ./user_stories/scripts/demo_comments_story.sh
#   DEMO_EMAIL=demo@example.test DEMO_PASSWORD=DemoPass123! ./user_stories/scripts/demo_comments_story.sh

set -o pipefail

TOTAL_STEPS=12
RUN_ID="$(date +%Y%m%d%H%M%S)-$$"

BASE_URL="${BASE_URL:-http://localhost:8080}"
BASE_URL="${BASE_URL%/}"
DEMO_EMAIL="${DEMO_EMAIL:-comments-demo-${RUN_ID}@example.test}"
DEMO_USERNAME="${DEMO_USERNAME:-comments_${RUN_ID//[^A-Za-z0-9_]/_}}"
DEMO_PASSWORD="${DEMO_PASSWORD:-DemoPass123!}"
MOVIE_ID="${MOVIE_ID:-}"

TOKEN=""
USER_ID=""
FORGED_USER_ID=""
COMMENT_ID=""

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
  sed -n '2,23p' "$0" | sed 's/^# \{0,1\}//'
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

  heading 12 "Summary"
  explain "The CLI reports which parts of the comment lifecycle walkthrough succeeded, failed, or were skipped."

  printf '\n%sResult:%s\n' "$BOLD" "$RESET"
  for i in "${!SUMMARY_NAMES[@]}"; do
    printf '  %-5s %s - %s\n' "$(status_label "${SUMMARY_STATUS[$i]}")" "${SUMMARY_NAMES[$i]}" "${SUMMARY_DETAILS[$i]}"
  done

  printf '\n%sConfiguration:%s\n' "$BOLD" "$RESET"
  printf '  BASE_URL: %s\n' "$BASE_URL"
  printf '  DEMO_EMAIL: %s\n' "$DEMO_EMAIL"
  printf '  DEMO_USERNAME: %s\n' "$DEMO_USERNAME"
  printf '  USER_ID: %s\n' "${USER_ID:-<not logged in>}"
  printf '  FORGED_BODY_USER_ID: %s\n' "${FORGED_USER_ID:-<not set>}"
  printf '  MOVIE_ID: %s\n' "${MOVIE_ID:-<not selected>}"
  printf '  COMMENT_ID: %s\n' "${COMMENT_ID:-<not created>}"
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

is_conflict_status() {
  [[ "$1" == "409" ]]
}

is_not_found_status() {
  [[ "$1" == "404" ]]
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

extract_user_id() {
  local body="$1"

  jq -r '
    .data.user.id
    // .user.id
    // .data.id
    // .id
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

extract_comment_id() {
  local body="$1"

  jq -r '
    .data.id
    // .id
    // empty
  ' <<<"$body" 2>/dev/null
}

body_has_comment_id() {
  local body="$1"
  local comment_id="$2"

  jq -e --argjson id "$comment_id" '
    def collection:
      if (.data | type) == "array" then .data
      elif (.comments | type) == "array" then .comments
      elif type == "array" then .
      else []
      end;

    any(collection[]?; .id == $id)
  ' <<<"$body" >/dev/null 2>&1
}

comment_matches_owner_and_movie() {
  local body="$1"
  local user_id="$2"
  local movie_id="$3"

  jq -e --argjson user_id "$user_id" --arg movie_id "$movie_id" '
    .data.user_id == $user_id and .data.movie_id == $movie_id
  ' <<<"$body" >/dev/null 2>&1
}

comment_matches_owner() {
  local body="$1"
  local user_id="$2"

  jq -e --argjson user_id "$user_id" '
    .data.user_id == $user_id
  ' <<<"$body" >/dev/null 2>&1
}

REGISTER_PAYLOAD="$(
  jq -n \
    --arg email "$DEMO_EMAIL" \
    --arg username "$DEMO_USERNAME" \
    --arg password "$DEMO_PASSWORD" \
    '{
      email: $email,
      username: $username,
      first_name: "Comment",
      last_name: "Demo",
      password: $password
    }'
)"

LOGIN_PAYLOAD="$(
  jq -n \
    --arg email "$DEMO_EMAIL" \
    --arg password "$DEMO_PASSWORD" \
    '{email: $email, password: $password}'
)"

COMMENT_TEXT="Comment demo ${RUN_ID}: this movie deserves a discussion thread."
UPDATED_COMMENT_TEXT="Comment demo ${RUN_ID}: edited after a second look."

heading 1 "Health Check"
explain "The CLI checks whether the backend is reachable before running the comment walkthrough."
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
explain "The CLI creates a demo user for the comment lifecycle. Existing users are accepted so repeated runs can continue."
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
explain "The CLI signs in as the demo user because comment routes are protected by the API."
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
  abort_critical "Login failed; authenticated comment requests cannot continue."
fi

heading 4 "Store Auth Context"
explain "The CLI extracts the JWT and user ID from the login response for authenticated comment requests."
TOKEN="$(extract_token "$HTTP_BODY")"
USER_ID="$(extract_user_id "$HTTP_BODY")"

if [[ -n "$TOKEN" && -n "$USER_ID" ]]; then
  FORGED_USER_ID=$((USER_ID + 100000))
  record_result "Store auth context" "OK" "JWT token and user ID $USER_ID were found."
  print_result "OK" "JWT token and user ID $USER_ID were found."
elif [[ -n "$TOKEN" ]]; then
  record_result "Store auth context" "FAIL" "JWT token was found, but no user ID was available."
  print_result "FAIL" "JWT token was found, but no user ID was available."
  abort_critical "No user ID was available for the current comment API contract."
else
  record_result "Store auth context" "FAIL" "Login response did not include a JWT token."
  print_result "FAIL" "Login response did not include a JWT token."
  abort_critical "No JWT token was available for authenticated requests."
fi

heading 5 "Select Movie"
explain "The CLI uses MOVIE_ID when provided, otherwise it selects the first movie from the public movie list."

if [[ -n "$MOVIE_ID" ]]; then
  record_result "Select movie" "OK" "Using configured movie ID $MOVIE_ID."
  print_result "OK" "Using configured movie ID $MOVIE_ID."
else
  MOVIES_URL="$BASE_URL/api/v1/movies"
  print_request "GET" "$MOVIES_URL"
  request "GET" "$MOVIES_URL"
  print_response "$HTTP_STATUS" "$HTTP_BODY" "$CURL_ERROR"

  if is_success_status "$HTTP_STATUS"; then
    MOVIE_ID="$(extract_first_movie_id "$HTTP_BODY")"
    if [[ -n "$MOVIE_ID" ]]; then
      record_result "Select movie" "OK" "Selected first movie ID $MOVIE_ID."
      print_result "OK" "Selected first movie ID $MOVIE_ID."
    else
      record_result "Select movie" "FAIL" "Movie list loaded, but no movie ID was found."
      print_result "FAIL" "Movie list loaded, but no movie ID was found."
      abort_critical "No movie ID was available for the comment walkthrough."
    fi
  else
    record_result "Select movie" "FAIL" "Movie list request failed and MOVIE_ID was not provided."
    print_result "FAIL" "Movie list request failed and MOVIE_ID was not provided."
    abort_critical "No movie ID was available for the comment walkthrough."
  fi
fi

heading 6 "Load Existing Movie Comments"
explain "The CLI loads existing comments for the selected movie. An empty comment list is acceptable for a fresh database."
MOVIE_COMMENTS_URL="$BASE_URL/api/v1/movies/$MOVIE_ID/comments"
print_request "GET" "$MOVIE_COMMENTS_URL" "" "true"
request "GET" "$MOVIE_COMMENTS_URL" "" "true"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$CURL_ERROR"

if is_success_status "$HTTP_STATUS"; then
  record_result "Load movie comments" "OK" "Existing comments were loaded for $MOVIE_ID."
  print_result "OK" "Existing comments were loaded for $MOVIE_ID."
elif is_not_found_status "$HTTP_STATUS"; then
  record_result "Load movie comments" "WARN" "No comments exist yet for $MOVIE_ID."
  print_result "WARN" "No comments exist yet for $MOVIE_ID."
else
  record_result "Load movie comments" "WARN" "Existing comments could not be loaded."
  print_result "WARN" "Existing comments could not be loaded."
fi

heading 7 "Post Comment"
explain "The CLI intentionally sends a forged user_id and movie_id in the JSON body. The backend should ignore both and use the JWT user plus URL movie."
CREATE_COMMENT_PAYLOAD="$(
  jq -n \
    --argjson user_id "$FORGED_USER_ID" \
    --arg movie_id "tt-forged-body" \
    --arg content "$COMMENT_TEXT" \
    '{user_id: $user_id, movie_id: $movie_id, content: $content}'
)"
print_request "POST" "$MOVIE_COMMENTS_URL" "$CREATE_COMMENT_PAYLOAD" "true"
request "POST" "$MOVIE_COMMENTS_URL" "$CREATE_COMMENT_PAYLOAD" "true"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$CURL_ERROR"

if is_success_status "$HTTP_STATUS"; then
  COMMENT_ID="$(extract_comment_id "$HTTP_BODY")"
  if [[ -n "$COMMENT_ID" ]] && comment_matches_owner_and_movie "$HTTP_BODY" "$USER_ID" "$MOVIE_ID"; then
    record_result "Post comment" "OK" "Created comment ID $COMMENT_ID using JWT user $USER_ID and URL movie $MOVIE_ID."
    print_result "OK" "Created comment ID $COMMENT_ID using JWT user $USER_ID and URL movie $MOVIE_ID."
  elif [[ -n "$COMMENT_ID" ]]; then
    record_result "Post comment" "FAIL" "Comment was created, but ownership did not match the JWT user or URL movie."
    print_result "FAIL" "Comment was created, but ownership did not match the JWT user or URL movie."
    abort_critical "Comment ownership did not come from the JWT user and URL movie."
  else
    record_result "Post comment" "FAIL" "Comment was created, but no comment ID was returned."
    print_result "FAIL" "Comment was created, but no comment ID was returned."
    abort_critical "No comment ID was available for read, update, and delete steps."
  fi
else
  record_result "Post comment" "FAIL" "Comment creation failed."
  print_result "FAIL" "Comment creation failed."
  abort_critical "Comment creation failed; read, update, and delete steps cannot continue."
fi

heading 8 "Verify Comment Appears"
explain "The CLI reloads movie comments and checks that the newly created comment appears in the list."
print_request "GET" "$MOVIE_COMMENTS_URL" "" "true"
request "GET" "$MOVIE_COMMENTS_URL" "" "true"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$CURL_ERROR"

if is_success_status "$HTTP_STATUS" && body_has_comment_id "$HTTP_BODY" "$COMMENT_ID"; then
  record_result "Verify movie comments" "OK" "New comment ID $COMMENT_ID appears for movie $MOVIE_ID."
  print_result "OK" "New comment ID $COMMENT_ID appears for movie $MOVIE_ID."
elif is_success_status "$HTTP_STATUS"; then
  record_result "Verify movie comments" "WARN" "Movie comments loaded, but ID $COMMENT_ID was not found."
  print_result "WARN" "Movie comments loaded, but ID $COMMENT_ID was not found."
else
  record_result "Verify movie comments" "WARN" "Movie comments could not be reloaded."
  print_result "WARN" "Movie comments could not be reloaded."
fi

heading 9 "Edit Comment"
explain "The CLI edits the comment while again sending a forged user_id. The backend should still use the JWT user as the owner check."
UPDATE_COMMENT_URL="$BASE_URL/api/v1/comments/$COMMENT_ID"
UPDATE_COMMENT_PAYLOAD="$(
  jq -n \
    --argjson user_id "$FORGED_USER_ID" \
    --arg content "$UPDATED_COMMENT_TEXT" \
    '{user_id: $user_id, content: $content}'
)"
print_request "PATCH" "$UPDATE_COMMENT_URL" "$UPDATE_COMMENT_PAYLOAD" "true"
request "PATCH" "$UPDATE_COMMENT_URL" "$UPDATE_COMMENT_PAYLOAD" "true"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$CURL_ERROR"

if is_success_status "$HTTP_STATUS" && comment_matches_owner "$HTTP_BODY" "$USER_ID"; then
  record_result "Edit comment" "OK" "Comment ID $COMMENT_ID was updated as JWT user $USER_ID."
  print_result "OK" "Comment ID $COMMENT_ID was updated as JWT user $USER_ID."
elif is_success_status "$HTTP_STATUS"; then
  record_result "Edit comment" "FAIL" "Comment update succeeded, but response owner did not match JWT user $USER_ID."
  print_result "FAIL" "Comment update succeeded, but response owner did not match JWT user $USER_ID."
else
  record_result "Edit comment" "WARN" "Comment update failed."
  print_result "WARN" "Comment update failed."
fi

heading 10 "Fetch Comment By ID"
explain "The CLI fetches the single comment to show the updated content returned by the API."
print_request "GET" "$UPDATE_COMMENT_URL" "" "true"
request "GET" "$UPDATE_COMMENT_URL" "" "true"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$CURL_ERROR"

if is_success_status "$HTTP_STATUS"; then
  record_result "Fetch comment" "OK" "Comment ID $COMMENT_ID was fetched."
  print_result "OK" "Comment ID $COMMENT_ID was fetched."
else
  record_result "Fetch comment" "WARN" "Comment ID $COMMENT_ID could not be fetched."
  print_result "WARN" "Comment ID $COMMENT_ID could not be fetched."
fi

heading 11 "Delete Comment"
explain "The CLI deletes the comment as cleanup while sending a forged body user_id. The backend should authorize the delete from the JWT."
DELETE_COMMENT_PAYLOAD="$(
  jq -n \
    --argjson user_id "$FORGED_USER_ID" \
    '{user_id: $user_id}'
)"
print_request "DELETE" "$UPDATE_COMMENT_URL" "$DELETE_COMMENT_PAYLOAD" "true"
request "DELETE" "$UPDATE_COMMENT_URL" "$DELETE_COMMENT_PAYLOAD" "true"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$CURL_ERROR"

if is_success_status "$HTTP_STATUS"; then
  record_result "Delete comment" "OK" "Comment ID $COMMENT_ID was deleted."
  print_result "OK" "Comment ID $COMMENT_ID was deleted."
else
  record_result "Delete comment" "WARN" "Comment delete failed; manual cleanup may be needed."
  print_result "WARN" "Comment delete failed; manual cleanup may be needed."
fi

print_summary
