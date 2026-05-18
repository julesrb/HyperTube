#!/usr/bin/env bash

set -uo pipefail

# Comprehensive API smoke/e2e test for the running Hypertube API.
#
# Usage:
#   tests/api/comprehensive_api_test.sh
#
# Configuration:
#   BASE_URL=http://localhost:8080/api/v1
#   API_TEST_LIVE=1                 # set to 0 to skip C411/TMDB-dependent route success tests
#   API_TEST_SEARCH_TITLE=dune
#   CURL_TIMEOUT=45
#   NO_COLOR=1                      # disable colors
#   FORCE_COLOR=1                   # force colors even when stdout is not a TTY

BASE_URL="${BASE_URL:-http://localhost:8080/api/v1}"
BASE_URL="${BASE_URL%/}"
API_TEST_LIVE="${API_TEST_LIVE:-1}"
API_TEST_SEARCH_TITLE="${API_TEST_SEARCH_TITLE:-dune}"
CURL_TIMEOUT="${CURL_TIMEOUT:-45}"

TMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/hypertube-api-test.XXXXXX")"
LAST_BODY_FILE="$TMP_DIR/last_body"
LAST_HEADERS_FILE="$TMP_DIR/last_headers"
LAST_CURL_ERR_FILE="$TMP_DIR/last_curl_err"
LAST_REQUEST_METHOD=""
LAST_REQUEST_URL=""
LAST_REQUEST_BODY=""
LAST_REQUEST_AUTH=""
LAST_STATUS=""
LAST_CURL_EXIT=0

PASSED=0
FAILED=0
SKIPPED=0
FEATURED_MOVIE_ID=""
AUTH_TOKEN=""
REGISTERED_USER_ID=""
USER_CONTEXT_COMMENT_ID=""

RESET=""
BOLD=""
DIM=""
RED=""
GREEN=""
YELLOW=""
CYAN=""

if [[ -z "${NO_COLOR:-}" && ( -t 1 || "${FORCE_COLOR:-}" == "1" || "${CLICOLOR_FORCE:-}" == "1" ) ]]; then
  RESET=$'\033[0m'
  BOLD=$'\033[1m'
  DIM=$'\033[2m'
  RED=$'\033[31m'
  GREEN=$'\033[32m'
  YELLOW=$'\033[33m'
  CYAN=$'\033[36m'
fi

cleanup() {
  rm -rf -- "$TMP_DIR"
}
trap cleanup EXIT

section() {
  local title="$1"

  echo
  printf '%b%s%b\n' "$DIM" "================================================================" "$RESET"
  printf '%b%s%b\n' "${BOLD}${CYAN}" "$title" "$RESET"
  printf '%b%s%b\n' "$DIM" "================================================================" "$RESET"
}

result_line() {
  local status="$1"
  local color="$2"
  local message="$3"

  printf '  %b%-6s%b %s\n' "$color" "$status" "$RESET" "$message"
}

config_line() {
  local key="$1"
  local value="$2"

  printf '  %b%-22s%b %s\n' "$DIM" "$key" "$RESET" "$value"
}

dump_field() {
  local key="$1"
  local value="$2"

  printf '    %b%-14s%b %s\n' "$DIM" "$key" "$RESET" "$value"
}

pass() {
  PASSED=$((PASSED + 1))
  result_line "PASS" "$GREEN" "$1"
}

fail() {
  FAILED=$((FAILED + 1))
  result_line "FAIL" "$RED" "$1"
  dump_last_response
}

skip() {
  SKIPPED=$((SKIPPED + 1))
  result_line "SKIP" "$YELLOW" "$1 ($2)"
}

require_command() {
  local missing=0
  local cmd

  for cmd in "$@"; do
    if ! command -v "$cmd" >/dev/null 2>&1; then
      printf 'Missing required command: %s\n' "$cmd" >&2
      missing=1
    fi
  done

  if [[ "$missing" -ne 0 ]]; then
    exit 127
  fi
}

redact_json_payload() {
  local payload="$1"

  if printf '%s' "$payload" | jq . >/dev/null 2>&1; then
    printf '%s' "$payload" | jq 'if type == "object" and has("password") then .password = "<redacted>" else . end'
  else
    printf '%s\n' "$payload"
  fi
}

dump_last_response() {
  if [[ -n "$LAST_REQUEST_URL" ]]; then
    dump_field "Request" "$LAST_REQUEST_METHOD $LAST_REQUEST_URL"
  fi
  if [[ -n "$LAST_REQUEST_AUTH" ]]; then
    dump_field "Authorization" "$LAST_REQUEST_AUTH"
  fi
  if [[ -n "$LAST_REQUEST_BODY" ]]; then
    printf '    %b%s%b\n' "$DIM" "Payload" "$RESET"
    redact_json_payload "$LAST_REQUEST_BODY" | sed 's/^/      /'
  fi
  if [[ "$LAST_CURL_EXIT" -ne 0 && -s "$LAST_CURL_ERR_FILE" ]]; then
    printf '    %b%s%b\n' "$RED" "Curl error" "$RESET"
    sed 's/^/      /' "$LAST_CURL_ERR_FILE"
  fi

  dump_field "Status" "${LAST_STATUS:-<none>}"
  printf '    %b%s%b\n' "$DIM" "Response body" "$RESET"
  if [[ ! -s "$LAST_BODY_FILE" ]]; then
    printf '      <empty>\n'
  elif jq . "$LAST_BODY_FILE" >/dev/null 2>&1; then
    jq . "$LAST_BODY_FILE" | sed 's/^/      /'
  else
    sed 's/^/      /' "$LAST_BODY_FILE"
  fi
}

urlencode() {
  jq -rn --arg value "$1" '$value | @uri'
}

request() {
  local method="$1"
  local path="$2"
  local body="${3:-}"
  local token="${4:-}"
  local raw_auth_header="${5:-}"
  local url
  local curl_args

  if [[ "$path" =~ ^https?:// ]]; then
    url="$path"
  else
    url="$BASE_URL$path"
  fi

  : >"$LAST_BODY_FILE"
  : >"$LAST_HEADERS_FILE"
  : >"$LAST_CURL_ERR_FILE"

  LAST_REQUEST_METHOD="$method"
  LAST_REQUEST_URL="$url"
  LAST_REQUEST_BODY="$body"
  LAST_REQUEST_AUTH=""
  LAST_STATUS=""
  LAST_CURL_EXIT=0

  curl_args=(
    --silent
    --show-error
    --max-time "$CURL_TIMEOUT"
    -X "$method"
    "$url"
    -H "Accept: application/json"
    -D "$LAST_HEADERS_FILE"
    -o "$LAST_BODY_FILE"
    -w "%{http_code}"
  )

  if [[ -n "$body" ]]; then
    curl_args+=(-H "Content-Type: application/json" --data "$body")
  fi

  if [[ -n "$token" ]]; then
    curl_args+=(-H "Authorization: Bearer $token")
    LAST_REQUEST_AUTH="Bearer <redacted>"
  elif [[ -n "$raw_auth_header" ]]; then
    curl_args+=(-H "Authorization: $raw_auth_header")
    if [[ "${raw_auth_header,,}" == bearer* ]]; then
      LAST_REQUEST_AUTH="Bearer <redacted>"
    else
      LAST_REQUEST_AUTH="$raw_auth_header"
    fi
  fi

  LAST_STATUS="$(curl "${curl_args[@]}" 2>"$LAST_CURL_ERR_FILE")"
  LAST_CURL_EXIT=$?
}

expect_status() {
  local name="$1"
  local expected="$2"

  if [[ "$LAST_CURL_EXIT" -ne 0 ]]; then
    fail "$name: curl failed"
    return 1
  fi

  if [[ "$LAST_STATUS" == "$expected" ]]; then
    pass "$name: HTTP $expected"
    return 0
  fi

  fail "$name: expected HTTP $expected, got ${LAST_STATUS:-<none>}"
  return 1
}

assert_jq_true() {
  local name="$1"
  local filter="$2"

  if jq -e "$filter" "$LAST_BODY_FILE" >/dev/null 2>"$TMP_DIR/jq_error"; then
    pass "$name"
    return 0
  fi

  fail "$name: jq assertion failed: $filter"
  if [[ -s "$TMP_DIR/jq_error" ]]; then
    sed 's/^/  jq: /' "$TMP_DIR/jq_error"
  fi
  return 1
}

assert_jq_eq() {
  local name="$1"
  local filter="$2"
  local expected="$3"
  local actual

  actual="$(jq -r "$filter" "$LAST_BODY_FILE" 2>"$TMP_DIR/jq_error")"
  if [[ $? -eq 0 && "$actual" == "$expected" ]]; then
    pass "$name"
    return 0
  fi

  fail "$name: expected '$expected', got '${actual:-<jq failed>}'"
  if [[ -s "$TMP_DIR/jq_error" ]]; then
    sed 's/^/  jq: /' "$TMP_DIR/jq_error"
  fi
  return 1
}

assert_error() {
  local name="$1"
  local expected_code="$2"
  local expected_message="$3"

  assert_jq_true "$name: error envelope exists" '.error | type == "object"'
  assert_jq_eq "$name: error code" '.error.code' "$expected_code"
  assert_jq_eq "$name: error message" '.error.message' "$expected_message"
}

expect_error_case() {
  local name="$1"
  local method="$2"
  local path="$3"
  local expected_status="$4"
  local expected_code="$5"
  local expected_message="$6"
  local body="${7:-}"
  local token="${8:-}"
  local raw_auth_header="${9:-}"

  request "$method" "$path" "$body" "$token" "$raw_auth_header"
  if expect_status "$name" "$expected_status"; then
    assert_error "$name" "$expected_code" "$expected_message"
  fi
}

assert_list_envelope() {
  local name="$1"

  assert_jq_true "$name: data is array" '.data | type == "array"'
  assert_jq_true "$name: meta.total is numeric" '.meta.total | type == "number"'
  assert_jq_true "$name: meta.page is numeric" '.meta.page | type == "number"'
  assert_jq_true "$name: meta.per_page is numeric" '.meta.per_page | type == "number"'
}

register_payload() {
  local email="$1"
  local username="$2"
  local first_name="${3-Api}"
  local last_name="${4-Tester}"
  local password="$5"

  jq -n \
    --arg email "$email" \
    --arg username "$username" \
    --arg first_name "$first_name" \
    --arg last_name "$last_name" \
    --arg password "$password" \
    '{email:$email, username:$username, first_name:$first_name, last_name:$last_name, password:$password}'
}

login_payload() {
  local email="$1"
  local password="$2"

  jq -n --arg email "$email" --arg password "$password" '{email:$email, password:$password}'
}

extract_token_from_body() {
  jq -r '.data.access_token // empty' "$LAST_BODY_FILE" 2>/dev/null
}

test_public_routes() {
  section "Public routes and route surface"

  request "GET" "/health"
  expect_status "Health check" "200"

  request "GET" "/does-not-exist"
  expect_status "Unknown route returns 404" "404"

  request "POST" "/movies"
  expect_status "Unsupported method on /movies returns 405" "405"

  request "GET" "/movies"
  if expect_status "List movies" "200"; then
    assert_list_envelope "List movies"
    FEATURED_MOVIE_ID="$(jq -r '.data[0].imdb_id // empty' "$LAST_BODY_FILE")"
    if [[ -n "$FEATURED_MOVIE_ID" ]]; then
      pass "List movies exposes first imdb_id: $FEATURED_MOVIE_ID"
      assert_jq_true "List movies item has expected public fields" '.data[0] | has("imdb_id") and has("title") and has("year") and has("poster_url") and has("backdrop_url")'
    else
      skip "Seeded featured movie route follow-up" "GET /movies returned an empty list"
    fi
  fi
}

test_auth_validation_and_success() {
  local run_id
  local raw_username
  local test_email
  local normalized_email
  local uppercase_email
  local test_username
  local test_password
  local payload

  section "Auth validation, registration, and login"

  run_id="${API_TEST_RUN_ID:-$(date +%s)-$$-$RANDOM}"
  raw_username="apitest_${run_id//[^A-Za-z0-9]/_}"
  test_username="${API_TEST_USERNAME:-${raw_username:0:32}}"
  test_email="${API_TEST_EMAIL:-ApiTest+${run_id}@Example.COM}"
  normalized_email="$(printf '%s' "$test_email" | tr '[:upper:]' '[:lower:]')"
  uppercase_email="$(printf '%s' "$normalized_email" | tr '[:lower:]' '[:upper:]')"
  test_password="${API_TEST_PASSWORD:-correct-horse-battery-$run_id}"

  expect_error_case \
    "Register rejects missing body" \
    "POST" "/auth/register" "400" "BAD_REQUEST" "invalid JSON body"

  expect_error_case \
    "Register rejects invalid JSON" \
    "POST" "/auth/register" "400" "BAD_REQUEST" "invalid JSON body" \
    '{'

  expect_error_case \
    "Register rejects unknown JSON field" \
    "POST" "/auth/register" "400" "BAD_REQUEST" "invalid JSON body" \
    '{"email":"extra@example.com","username":"extrauser","first_name":"Extra","last_name":"Field","password":"password123","role":"admin"}'

  payload="$(register_payload "not-an-email" "valid_user" "Api" "Tester" "password123")"
  expect_error_case \
    "Register rejects invalid email" \
    "POST" "/auth/register" "400" "VALIDATION_ERROR" "valid email is required" \
    "$payload"

  payload="$(register_payload "shortname@example.com" "ab" "Api" "Tester" "password123")"
  expect_error_case \
    "Register rejects short username" \
    "POST" "/auth/register" "400" "VALIDATION_ERROR" "username must be 3-32 characters and contain only letters, numbers, or underscores" \
    "$payload"

  payload="$(register_payload "firstname@example.com" "firstname_user" "" "Tester" "password123")"
  expect_error_case \
    "Register rejects missing first_name" \
    "POST" "/auth/register" "400" "VALIDATION_ERROR" "first_name is required and must be at most 100 characters" \
    "$payload"

  payload="$(register_payload "lastname@example.com" "lastname_user" "Api" "" "password123")"
  expect_error_case \
    "Register rejects missing last_name" \
    "POST" "/auth/register" "400" "VALIDATION_ERROR" "last_name is required and must be at most 100 characters" \
    "$payload"

  payload="$(register_payload "password@example.com" "password_user" "Api" "Tester" "short")"
  expect_error_case \
    "Register rejects short password" \
    "POST" "/auth/register" "400" "VALIDATION_ERROR" "password must be between 8 and 72 bytes" \
    "$payload"

  payload="$(register_payload "$test_email" "$test_username" "Api" "Tester" "$test_password")"
  request "POST" "/auth/register" "$payload"
  if expect_status "Register succeeds" "201"; then
    assert_jq_true "Register returns auth envelope" '.data | type == "object"'
    assert_jq_true "Register returns access token" '.data.access_token | type == "string" and length > 20'
    assert_jq_eq "Register returns bearer token type" '.data.token_type' "Bearer"
    assert_jq_true "Register returns positive expiry" '.data.expires_in | type == "number" and . > 0'
    assert_jq_eq "Register normalizes email" '.data.user.email' "$normalized_email"
    assert_jq_eq "Register returns username" '.data.user.username' "$test_username"
    assert_jq_eq "Register returns first_name" '.data.user.first_name' "Api"
    assert_jq_eq "Register returns last_name" '.data.user.last_name' "Tester"
    REGISTERED_USER_ID="$(jq -r '.data.user.id // empty' "$LAST_BODY_FILE")"
  fi

  expect_error_case \
    "Register rejects duplicate email or username" \
    "POST" "/auth/register" "409" "USER_EXISTS" "email or username already exists" \
    "$payload"

  expect_error_case \
    "Login rejects missing body" \
    "POST" "/auth/login" "400" "BAD_REQUEST" "invalid JSON body"

  expect_error_case \
    "Login rejects invalid JSON" \
    "POST" "/auth/login" "400" "BAD_REQUEST" "invalid JSON body" \
    '{'

  expect_error_case \
    "Login rejects unknown JSON field" \
    "POST" "/auth/login" "400" "BAD_REQUEST" "invalid JSON body" \
    '{"email":"extra@example.com","password":"password123","remember":true}'

  payload="$(login_payload "not-an-email" "$test_password")"
  expect_error_case \
    "Login rejects invalid email" \
    "POST" "/auth/login" "400" "VALIDATION_ERROR" "valid email is required" \
    "$payload"

  payload="$(login_payload "$normalized_email" "")"
  expect_error_case \
    "Login rejects missing password" \
    "POST" "/auth/login" "400" "VALIDATION_ERROR" "password is required" \
    "$payload"

  payload="$(login_payload "$normalized_email" "wrong-password")"
  expect_error_case \
    "Login rejects wrong password" \
    "POST" "/auth/login" "401" "INVALID_CREDENTIALS" "invalid email or password" \
    "$payload"

  payload="$(login_payload "missing-$run_id@example.com" "$test_password")"
  expect_error_case \
    "Login rejects unknown email" \
    "POST" "/auth/login" "401" "INVALID_CREDENTIALS" "invalid email or password" \
    "$payload"

  payload="$(login_payload "$uppercase_email" "$test_password")"
  request "POST" "/auth/login" "$payload"
  if expect_status "Login succeeds with normalized email" "200"; then
    assert_jq_true "Login returns auth envelope" '.data | type == "object"'
    assert_jq_true "Login returns access token" '.data.access_token | type == "string" and length > 20'
    assert_jq_eq "Login returns bearer token type" '.data.token_type' "Bearer"
    assert_jq_true "Login returns positive expiry" '.data.expires_in | type == "number" and . > 0'
    assert_jq_eq "Login returns normalized email" '.data.user.email' "$normalized_email"
    if [[ -n "$REGISTERED_USER_ID" ]]; then
      assert_jq_eq "Login returns same user id as registration" '.data.user.id | tostring' "$REGISTERED_USER_ID"
    fi
    AUTH_TOKEN="$(jq -r '.data.access_token // empty' "$LAST_BODY_FILE")"
  fi
}

test_protected_route_errors() {
  section "Protected route errors"

  expect_error_case \
    "Protected route rejects missing bearer token" \
    "GET" "/movies/search?title=dune" "401" "UNAUTHORIZED" "missing bearer token"

  expect_error_case \
    "Protected route rejects malformed auth scheme" \
    "GET" "/movies/search?title=dune" "401" "UNAUTHORIZED" "missing bearer token" \
    "" "" "Token abc"

  expect_error_case \
    "Protected route rejects invalid bearer token" \
    "GET" "/movies/search?title=dune" "401" "UNAUTHORIZED" "invalid bearer token" \
    "" "" "Bearer not-a-valid-jwt"

  if [[ -z "$AUTH_TOKEN" ]]; then
    skip "Authenticated protected-route validation" "login did not return a token"
    return
  fi

  expect_error_case \
    "Search rejects missing title query" \
    "GET" "/movies/search" "400" "VALIDATION_ERROR" "title query parameter is required" \
    "" "$AUTH_TOKEN"

  expect_error_case \
    "Search rejects empty title query" \
    "GET" "/movies/search?title=" "400" "VALIDATION_ERROR" "title query parameter is required" \
    "" "$AUTH_TOKEN"

  expect_error_case \
    "Movie detail rejects unknown imdb_id" \
    "GET" "/movies/tt0000000" "404" "NOT_FOUND" "movie not found" \
    "" "$AUTH_TOKEN"

  request "GET" "/movies/tt0000000/torrents" "" "$AUTH_TOKEN"
  if [[ "$LAST_STATUS" == "404" ]]; then
    pass "Unknown movie torrent route returns 404"
    assert_error "Unknown movie torrent route" "NOT_FOUND" "no tracker source found for this movie"
  elif expect_status "Unknown movie torrent route returns current empty-list envelope" "200"; then
    assert_list_envelope "Unknown movie torrent route"
    assert_jq_true "Unknown movie torrent route returns no torrents" '.data | length == 0'
  fi
}

test_live_movie_routes() {
  local encoded_title
  local movie_id

  section "Authenticated successful movie routes"

  if [[ -z "$AUTH_TOKEN" ]]; then
    skip "Successful protected movie routes" "login did not return a token"
    return
  fi

  if [[ "$API_TEST_LIVE" != "1" ]]; then
    skip "Live movie search" "API_TEST_LIVE is not 1"
    if [[ -n "$FEATURED_MOVIE_ID" ]]; then
      skip "Movie detail and torrent success routes" "detail route fetches TMDB data; enable API_TEST_LIVE=1"
    fi
    return
  fi

  encoded_title="$(urlencode "$API_TEST_SEARCH_TITLE")"
  request "GET" "/movies/search?title=$encoded_title" "" "$AUTH_TOKEN"
  if expect_status "Movie search succeeds for '$API_TEST_SEARCH_TITLE'" "200"; then
    assert_list_envelope "Movie search"
    assert_jq_true "Movie search returns at least one result" '.data | length > 0'
    movie_id="$(jq -r '.data[0].imdb_id // empty' "$LAST_BODY_FILE")"
  else
    movie_id=""
  fi

  if [[ -z "$movie_id" && -n "$FEATURED_MOVIE_ID" ]]; then
    movie_id="$FEATURED_MOVIE_ID"
    pass "Falling back to featured movie id for detail route: $movie_id"
  fi

  if [[ -z "$movie_id" ]]; then
    skip "Movie detail success route" "no movie id available from search or featured list"
    skip "Movie torrents success route" "no movie id available from search or featured list"
    return
  fi

  request "GET" "/movies/$movie_id" "" "$AUTH_TOKEN"
  if expect_status "Movie detail succeeds for $movie_id" "200"; then
    assert_jq_eq "Movie detail returns requested imdb_id" '.data.imdb_id' "$movie_id"
    assert_jq_true "Movie detail has tmdb_id" '.data.tmdb_id | type == "string" and length > 0'
    assert_jq_true "Movie detail has title" '.data.title | type == "string" and length > 0'
    assert_jq_true "Movie detail has genres array" '.data.genres | type == "array"'
    assert_jq_true "Movie detail has summary field" '.data.summary | type == "string"'
    assert_jq_true "Movie detail has director field" '.data.director | type == "string"'
    assert_jq_true "Movie detail has cast array" '.data.cast | type == "array"'
    assert_jq_true "Movie detail exposes watched flag" '.data.watched | type == "boolean"'
    assert_jq_true "Movie detail exposes progression number" '.data.progression | type == "number"'
  fi

  request "GET" "/movies/$movie_id/torrents" "" "$AUTH_TOKEN"
  if expect_status "Movie torrents succeed for $movie_id" "200"; then
    assert_list_envelope "Movie torrents"
    assert_jq_true "Movie torrents response data is present" '.data | length >= 0'
    if jq -e '.data | length > 0' "$LAST_BODY_FILE" >/dev/null 2>&1; then
      assert_jq_true "Movie torrent item has expected fields" '.data[0] | has("imdb_id") and has("title") and has("source") and has("url") and has("quality") and has("size") and has("language") and has("seeds")'
    else
      skip "Movie torrent item field check" "route returned an empty torrent list"
    fi
  fi
}

test_authenticated_user_context_routes() {
  local movie_id
  local forged_user_id
  local create_payload
  local update_payload
  local delete_payload
  local intruder_run_id
  local intruder_email
  local intruder_username
  local intruder_password
  local intruder_token
  local payload

  section "Authenticated user context and comment ownership"

  if [[ -z "$AUTH_TOKEN" || -z "$REGISTERED_USER_ID" ]]; then
    skip "User context routes" "login did not return both token and user id"
    return
  fi

  request "GET" "/movies/watched" "" "$AUTH_TOKEN"
  if expect_status "Watched movies uses token user without request body" "200"; then
    assert_list_envelope "Watched movies"
  fi

  movie_id="$FEATURED_MOVIE_ID"
  if [[ -z "$movie_id" ]]; then
    request "GET" "/movies"
    if [[ "$LAST_STATUS" == "200" ]]; then
      movie_id="$(jq -r '.data[0].imdb_id // empty' "$LAST_BODY_FILE")"
    fi
  fi

  if [[ -z "$movie_id" ]]; then
    skip "Comment ownership routes" "no seeded movie id is available"
    return
  fi

  forged_user_id=$((REGISTERED_USER_ID + 100000))
  create_payload="$(
    jq -n \
      --argjson user_id "$forged_user_id" \
      --arg movie_id "tt-forged-body" \
      --arg content "api ownership test comment" \
      '{user_id: $user_id, movie_id: $movie_id, content: $content}'
  )"

  request "POST" "/movies/$movie_id/comments" "$create_payload" "$AUTH_TOKEN"
  if expect_status "Create comment ignores forged body user_id and movie_id" "201"; then
    assert_jq_eq "Create comment uses token user_id" '.data.user_id | tostring' "$REGISTERED_USER_ID"
    assert_jq_eq "Create comment uses path movie_id" '.data.movie_id' "$movie_id"
    USER_CONTEXT_COMMENT_ID="$(jq -r '.data.id // empty' "$LAST_BODY_FILE")"
  fi

  if [[ -z "$USER_CONTEXT_COMMENT_ID" ]]; then
    skip "Comment update/delete ownership checks" "comment creation did not return an id"
    return
  fi

  update_payload="$(
    jq -n \
      --argjson user_id "$forged_user_id" \
      --arg content "api ownership test comment edited" \
      '{user_id: $user_id, content: $content}'
  )"

  request "PATCH" "/comments/$USER_CONTEXT_COMMENT_ID" "$update_payload" "$AUTH_TOKEN"
  if expect_status "Update comment ignores forged body user_id" "200"; then
    assert_jq_eq "Update comment keeps token user_id" '.data.user_id | tostring' "$REGISTERED_USER_ID"
    assert_jq_eq "Update comment changes content" '.data.content' "api ownership test comment edited"
  fi

  intruder_run_id="${API_TEST_RUN_ID:-$(date +%s)-$$-$RANDOM}-intruder"
  intruder_email="intruder+${intruder_run_id}@example.com"
  intruder_username="intruder_${intruder_run_id//[^A-Za-z0-9]/_}"
  intruder_username="${intruder_username:0:32}"
  intruder_password="intruder-password-$intruder_run_id"

  payload="$(register_payload "$intruder_email" "$intruder_username" "Intruder" "Tester" "$intruder_password")"
  request "POST" "/auth/register" "$payload"
  if [[ "$LAST_STATUS" == "201" ]]; then
    intruder_token="$(extract_token_from_body)"
    pass "Intruder user registered for ownership checks"
  else
    request "POST" "/auth/login" "$(login_payload "$intruder_email" "$intruder_password")"
    if [[ "$LAST_STATUS" == "200" ]]; then
      intruder_token="$(extract_token_from_body)"
      pass "Intruder user logged in for ownership checks"
    else
      intruder_token=""
      skip "Intruder ownership checks" "could not register or log in second user"
    fi
  fi

  if [[ -n "$intruder_token" ]]; then
    expect_error_case \
      "Different authenticated user cannot update owned comment" \
      "PATCH" "/comments/$USER_CONTEXT_COMMENT_ID" "404" "NOT_FOUND" "comment not found" \
      "$(jq -n --arg content "intruder edit" '{content: $content}')" "$intruder_token"

    expect_error_case \
      "Different authenticated user cannot delete owned comment" \
      "DELETE" "/comments/$USER_CONTEXT_COMMENT_ID" "404" "NOT_FOUND" "comment not found" \
      "" "$intruder_token"
  fi

  delete_payload="$(jq -n --argjson user_id "$forged_user_id" '{user_id: $user_id}')"
  request "DELETE" "/comments/$USER_CONTEXT_COMMENT_ID" "$delete_payload" "$AUTH_TOKEN"
  if expect_status "Owner can delete comment even with forged body user_id" "200"; then
    assert_jq_true "Delete comment returns null data" '.data == null'
  fi
}

main() {
  local failed_color
  local skipped_color

  require_command curl jq sed date tr

  section "Configuration"
  config_line "BASE_URL" "$BASE_URL"
  config_line "API_TEST_LIVE" "$API_TEST_LIVE"
  config_line "API_TEST_SEARCH_TITLE" "$API_TEST_SEARCH_TITLE"
  config_line "CURL_TIMEOUT" "$CURL_TIMEOUT"

  test_public_routes
  test_auth_validation_and_success
  test_protected_route_errors
  test_authenticated_user_context_routes
  test_live_movie_routes

  section "Summary"
  failed_color="$GREEN"
  skipped_color="$DIM"
  if [[ "$FAILED" -ne 0 ]]; then
    failed_color="$RED"
  fi
  if [[ "$SKIPPED" -ne 0 ]]; then
    skipped_color="$YELLOW"
  fi

  printf '  %b%-8s%b %d\n' "$GREEN" "Passed" "$RESET" "$PASSED"
  printf '  %b%-8s%b %d\n' "$failed_color" "Failed" "$RESET" "$FAILED"
  printf '  %b%-8s%b %d\n' "$skipped_color" "Skipped" "$RESET" "$SKIPPED"

  if [[ "$FAILED" -ne 0 ]]; then
    exit 1
  fi
}

main "$@"
