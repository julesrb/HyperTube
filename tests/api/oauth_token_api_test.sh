#!/usr/bin/env bash

set -uo pipefail

# API smoke test for the OAuth2 token endpoint.
#
# Usage:
#   tests/api/oauth_token_api_test.sh
#
# Configuration:
#   BASE_URL=http://localhost:8080/api/v1
#   API_ORIGIN=http://localhost:8080    # optional; used for the /oauth/token alias
#   CURL_TIMEOUT=20
#   NO_COLOR=1
#   FORCE_COLOR=1

BASE_URL="${BASE_URL:-http://localhost:8080/api/v1}"
BASE_URL="${BASE_URL%/}"
if [[ -n "${API_ORIGIN:-}" ]]; then
  API_ORIGIN="${API_ORIGIN%/}"
elif [[ "$BASE_URL" == */api/v1 ]]; then
  API_ORIGIN="${BASE_URL%/api/v1}"
else
  API_ORIGIN=""
fi
CURL_TIMEOUT="${CURL_TIMEOUT:-20}"

RUN_ID="${API_TEST_RUN_ID:-$(date +%s)-$$-$RANDOM}"
RAW_USERNAME="oauth_${RUN_ID//[^A-Za-z0-9]/_}"
TEST_USERNAME="${API_TEST_USERNAME:-${RAW_USERNAME:0:32}}"
TEST_EMAIL="${API_TEST_EMAIL:-OAuthToken+${RUN_ID}@Example.COM}"
NORMALIZED_EMAIL="$(printf '%s' "$TEST_EMAIL" | tr '[:upper:]' '[:lower:]')"
UPPERCASE_EMAIL="$(printf '%s' "$NORMALIZED_EMAIL" | tr '[:lower:]' '[:upper:]')"
TEST_PASSWORD="${API_TEST_PASSWORD:-OAuthPass123!-$RUN_ID}"

TMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/hypertube-oauth-token-test.XXXXXX")"
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
OAUTH_TOKEN=""

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

urlencode() {
  jq -rn --arg value "$1" '$value | @uri'
}

redact_payload() {
  local payload="$1"
  local content_type="$2"

  if [[ "$content_type" == "application/json" ]] && jq . >/dev/null 2>&1 <<<"$payload"; then
    jq 'if type == "object" and has("password") then .password = "<redacted>" else . end' <<<"$payload"
    return
  fi

  printf '%s\n' "$payload" | sed -E 's/(password=)[^&]*/\1<redacted>/g'
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
    redact_payload "$LAST_REQUEST_BODY" "$LAST_REQUEST_CONTENT_TYPE" | sed 's/^/      /'
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

request() {
  local method="$1"
  local path_or_url="$2"
  local body="${3:-}"
  local content_type="${4:-application/json}"
  local token="${5:-}"
  local url
  local curl_args

  if [[ "$path_or_url" =~ ^https?:// ]]; then
    url="$path_or_url"
  else
    url="$BASE_URL$path_or_url"
  fi

  : >"$LAST_BODY_FILE"
  : >"$LAST_HEADERS_FILE"
  : >"$LAST_CURL_ERR_FILE"

  LAST_REQUEST_METHOD="$method"
  LAST_REQUEST_URL="$url"
  LAST_REQUEST_BODY="$body"
  LAST_REQUEST_CONTENT_TYPE="$content_type"
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
    curl_args+=(-H "Content-Type: $content_type" --data "$body")
  fi

  if [[ -n "$token" ]]; then
    curl_args+=(-H "Authorization: Bearer $token")
    LAST_REQUEST_AUTH="Bearer <redacted>"
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

assert_oauth_error() {
  local name="$1"
  local expected="$2"

  assert_jq_eq "$name: OAuth error code" '.error' "$expected"
  assert_jq_true "$name: no access token on error" 'has("access_token") | not'
}

assert_error_envelope() {
  local name="$1"
  local expected_code="$2"
  local expected_message="$3"

  assert_jq_true "$name: error envelope exists" '.error | type == "object"'
  assert_jq_eq "$name: error code" '.error.code' "$expected_code"
  assert_jq_eq "$name: error message" '.error.message' "$expected_message"
}

assert_header_contains() {
  local name="$1"
  local header="$2"
  local pattern="$3"

  if sed 's/\r$//' "$LAST_HEADERS_FILE" | grep -iE "^${header}:.*${pattern}" >/dev/null; then
    pass "$name"
    return 0
  fi

  fail "$name: response header $header did not match $pattern"
  return 1
}

register_payload() {
  jq -n \
    --arg email "$TEST_EMAIL" \
    --arg username "$TEST_USERNAME" \
    --arg password "$TEST_PASSWORD" \
    '{email:$email, username:$username, first_name:"OAuth", last_name:"Tester", password:$password}'
}

token_form_payload() {
  local username="$1"
  local password="$2"
  local grant_type="${3:-password}"

  printf 'grant_type=%s&username=%s&password=%s' \
    "$(urlencode "$grant_type")" \
    "$(urlencode "$username")" \
    "$(urlencode "$password")"
}

require_command curl jq sed date tr grep

section "Configuration"
config_line "BASE_URL" "$BASE_URL"
config_line "API_ORIGIN" "${API_ORIGIN:-<not derived>}"
config_line "TEST_EMAIL" "$TEST_EMAIL"
config_line "TEST_USERNAME" "$TEST_USERNAME"
config_line "CURL_TIMEOUT" "$CURL_TIMEOUT"

section "Setup: register a password user"
request "POST" "/auth/register" "$(register_payload)" "application/json"
if expect_status "Register OAuth2 test user" "201"; then
  assert_jq_eq "Register normalizes email" '.data.user.email' "$NORMALIZED_EMAIL"
  assert_jq_eq "Register returns username" '.data.user.username' "$TEST_USERNAME"
  assert_jq_true "Register returns legacy auth token too" '.data.access_token | type == "string" and length > 20'
fi

section "Password grant with form encoding"
request "POST" "/oauth/token" "$(token_form_payload "$TEST_USERNAME" "$TEST_PASSWORD")" "application/x-www-form-urlencoded"
if expect_status "OAuth token endpoint accepts username password grant" "200"; then
  assert_jq_true "OAuth token response is top-level JSON" 'has("data") | not'
  assert_jq_true "OAuth token response includes access_token" '.access_token | type == "string" and length > 20'
  assert_jq_eq "OAuth token response uses Bearer" '.token_type' "Bearer"
  assert_jq_true "OAuth token response has positive expires_in" '.expires_in | type == "number" and . > 0'
  assert_header_contains "OAuth token response disables cache" "Cache-Control" "no-store"
  assert_header_contains "OAuth token response disables pragma cache" "Pragma" "no-cache"
  OAUTH_TOKEN="$(jq -r '.access_token // empty' "$LAST_BODY_FILE")"
fi

section "Bearer token reaches protected API handler"
if [[ -n "$OAUTH_TOKEN" ]]; then
  request "GET" "/movies/search" "" "application/json" "$OAUTH_TOKEN"
  if expect_status "OAuth token is accepted by protected route middleware" "400"; then
    assert_error_envelope "Protected route validation proves token reached handler" "VALIDATION_ERROR" "title query parameter is required"
  fi
else
  skip "Protected route token validation" "OAuth token was not available"
fi

section "Password grant with JSON and email login"
request "POST" "/oauth/token" \
  "$(jq -n --arg username "$UPPERCASE_EMAIL" --arg password "$TEST_PASSWORD" '{grant_type:"password", username:$username, password:$password}')" \
  "application/json"
if expect_status "OAuth token endpoint accepts email login in JSON body" "200"; then
  assert_jq_true "JSON grant returns access_token" '.access_token | type == "string" and length > 20'
  assert_jq_eq "JSON grant returns Bearer" '.token_type' "Bearer"
fi

section "OAuth2 error responses"
request "POST" "/oauth/token" "grant_type=client_credentials" "application/x-www-form-urlencoded"
if expect_status "OAuth token rejects unsupported grant type" "400"; then
  assert_oauth_error "Unsupported grant response" "unsupported_grant_type"
fi

request "POST" "/oauth/token" "grant_type=password" "application/x-www-form-urlencoded"
if expect_status "OAuth token rejects missing credentials" "400"; then
  assert_oauth_error "Missing credentials response" "invalid_request"
fi

request "POST" "/oauth/token" "$(token_form_payload "$TEST_USERNAME" "wrong-password")" "application/x-www-form-urlencoded"
if expect_status "OAuth token rejects wrong password" "400"; then
  assert_oauth_error "Wrong password response" "invalid_grant"
fi

section "Subject route alias"
if [[ -n "$API_ORIGIN" ]]; then
  request "POST" "$API_ORIGIN/oauth/token" "grant_type=client_credentials" "application/x-www-form-urlencoded"
  if expect_status "Root /oauth/token alias is registered" "400"; then
    assert_oauth_error "Root alias unsupported grant response" "unsupported_grant_type"
  fi
else
  skip "Root /oauth/token alias" "API_ORIGIN could not be derived from BASE_URL"
fi

section "Summary"
config_line "Passed" "$PASSED"
config_line "Failed" "$FAILED"
config_line "Skipped" "$SKIPPED"

if [[ "$FAILED" -ne 0 ]]; then
  printf '\n%bOAuth2 token API test failed.%b\n' "$RED" "$RESET" >&2
  exit 1
fi

printf '\n%bOAuth2 token API test passed.%b\n' "$GREEN" "$RESET"
