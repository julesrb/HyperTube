#!/usr/bin/env bash

set -uo pipefail

# API smoke test for the GitHub OAuth route surface.
#
# This test is safe to run without real GitHub credentials:
# - If /auth/github/login reports OAUTH_NOT_CONFIGURED, that path is accepted.
# - Callback CSRF and provider-denial behavior is still validated with a
#   simulated state cookie.
#
# Usage:
#   tests/api/github_auth_api_test.sh
#
# Configuration:
#   BASE_URL=http://localhost:8080/api/v1
#   CURL_TIMEOUT=20
#   NO_COLOR=1

BASE_URL="${BASE_URL:-http://localhost:8080/api/v1}"
BASE_URL="${BASE_URL%/}"
CURL_TIMEOUT="${CURL_TIMEOUT:-20}"

STATE_COOKIE_NAME="hypertube_oauth_github_state"
TMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/hypertube-github-auth-test.XXXXXX")"
LAST_BODY_FILE="$TMP_DIR/last_body"
LAST_HEADERS_FILE="$TMP_DIR/last_headers"
LAST_CURL_ERR_FILE="$TMP_DIR/last_curl_err"
LAST_REQUEST_METHOD=""
LAST_REQUEST_URL=""
LAST_REQUEST_COOKIE=""
LAST_STATUS=""
LAST_CURL_EXIT=0

PASSED=0
FAILED=0
SKIPPED=0
OAUTH_LOGIN_STATE=""
OAUTH_LOGIN_COOKIE=""

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

dump_last_response() {
  if [[ -n "$LAST_REQUEST_URL" ]]; then
    dump_field "Request" "$LAST_REQUEST_METHOD $LAST_REQUEST_URL"
  fi
  if [[ -n "$LAST_REQUEST_COOKIE" ]]; then
    dump_field "Cookie" "$LAST_REQUEST_COOKIE"
  fi
  if [[ "$LAST_CURL_EXIT" -ne 0 && -s "$LAST_CURL_ERR_FILE" ]]; then
    printf '    %b%s%b\n' "$RED" "Curl error" "$RESET"
    sed 's/^/      /' "$LAST_CURL_ERR_FILE"
  fi

  dump_field "Status" "${LAST_STATUS:-<none>}"

  printf '    %b%s%b\n' "$DIM" "Response headers" "$RESET"
  if [[ -s "$LAST_HEADERS_FILE" ]]; then
    sed 's/\r$//' "$LAST_HEADERS_FILE" | sed 's/^/      /'
  else
    printf '      <empty>\n'
  fi

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
  local path="$2"
  local cookie="${3:-}"
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
  LAST_REQUEST_COOKIE="$cookie"
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

  if [[ -n "$cookie" ]]; then
    curl_args+=(-H "Cookie: $cookie")
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

assert_header_contains() {
  local name="$1"
  local header="$2"
  local expected="$3"

  if sed 's/\r$//' "$LAST_HEADERS_FILE" | grep -i "^$header:" | grep -F "$expected" >/dev/null 2>&1; then
    pass "$name"
    return 0
  fi

  fail "$name: expected header $header to contain '$expected'"
  return 1
}

header_value() {
  local header="$1"

  sed 's/\r$//' "$LAST_HEADERS_FILE" \
    | awk -v header="$header" 'BEGIN { wanted = tolower(header) ":" } tolower($0) ~ "^" wanted { sub(/^[^:]*:[[:space:]]*/, ""); print; exit }'
}

set_cookie_value() {
  local cookie_name="$1"

  sed 's/\r$//' "$LAST_HEADERS_FILE" \
    | grep -i '^Set-Cookie:' \
    | sed -n "s/^Set-Cookie:[[:space:]]*$cookie_name=\\([^;]*\\).*/\\1/ip" \
    | head -n 1
}

extract_state_from_location() {
  local location="$1"

  printf '%s' "$location" | sed -n 's/.*[?&]state=\([^&#]*\).*/\1/p' | head -n 1
}

assert_location_contains() {
  local name="$1"
  local expected="$2"
  local location

  location="$(header_value "Location")"
  if [[ "$location" == *"$expected"* ]]; then
    pass "$name"
    return 0
  fi

  fail "$name: expected Location to contain '$expected'"
  return 1
}

assert_error_envelope() {
  local name="$1"
  local expected_code="$2"
  local expected_message="${3:-}"

  assert_jq_true "$name: error envelope exists" '.error | type == "object"'
  assert_jq_eq "$name: error code" '.error.code' "$expected_code"
  if [[ -n "$expected_message" ]]; then
    assert_jq_eq "$name: error message" '.error.message' "$expected_message"
  fi
}

assert_redirect_or_error() {
  local name="$1"
  local json_status="$2"
  local expected_code="$3"
  local expected_message="${4:-}"
  local location
  local plus_message

  if [[ "$LAST_STATUS" =~ ^30[12378]$ ]]; then
    pass "$name: redirected to frontend callback"
    assert_location_contains "$name: redirect carries error code" "error=$expected_code"
    if [[ -n "$expected_message" ]]; then
      location="$(header_value "Location")"
      plus_message="${expected_message// /+}"
      if [[ "$location" == *"$expected_message"* || "$location" == *"$plus_message"* ]]; then
        pass "$name: redirect carries error description"
      else
        fail "$name: expected Location to contain error description '$expected_message'"
      fi
    fi
    return 0
  fi

  if expect_status "$name" "$json_status"; then
    assert_error_envelope "$name" "$expected_code" "$expected_message"
  fi
}

test_health() {
  section "Health check"

  request "GET" "/health"
  expect_status "Health check" "200"
}

test_login_start() {
  local location
  local state
  local cookie_state

  section "GitHub login start"

  request "GET" "/auth/github/login"

  if [[ "$LAST_STATUS" == "503" ]]; then
    pass "GitHub login reports missing provider configuration"
    assert_error_envelope "GitHub login not configured response" "OAUTH_NOT_CONFIGURED" "GitHub OAuth is not configured"
    skip "GitHub authorization redirect checks" "GITHUB_CLIENT_ID, GITHUB_CLIENT_SECRET, or GITHUB_REDIRECT_URL is not configured"
    return 0
  fi

  if ! expect_status "GitHub login starts with redirect" "302"; then
    return 1
  fi

  location="$(header_value "Location")"
  state="$(extract_state_from_location "$location")"
  cookie_state="$(set_cookie_value "$STATE_COOKIE_NAME")"

  if [[ "$location" == https://github.com/login/oauth/authorize* ]]; then
    pass "GitHub login redirects to the official GitHub authorization endpoint"
  else
    fail "GitHub login redirects to the official GitHub authorization endpoint"
  fi

  [[ "$location" == *"response_type=code"* ]] && pass "GitHub authorize URL requests code flow" || fail "GitHub authorize URL requests code flow"
  [[ "$location" == *"read%3Auser"* && "$location" == *"user%3Aemail"* ]] && pass "GitHub authorize URL requests identity and email scopes" || fail "GitHub authorize URL requests identity and email scopes"
  [[ "$location" == *"client_id="* ]] && pass "GitHub authorize URL includes client_id" || fail "GitHub authorize URL includes client_id"
  [[ "$location" == *"redirect_uri="* ]] && pass "GitHub authorize URL includes redirect_uri" || fail "GitHub authorize URL includes redirect_uri"

  if [[ -n "$state" ]]; then
    pass "GitHub authorize URL includes state"
  else
    fail "GitHub authorize URL includes state"
  fi

  if [[ -n "$cookie_state" && "$cookie_state" == "$state" ]]; then
    pass "GitHub login stores matching state cookie"
    OAUTH_LOGIN_STATE="$state"
    OAUTH_LOGIN_COOKIE="$STATE_COOKIE_NAME=$cookie_state"
  else
    fail "GitHub login stores matching state cookie"
  fi

  assert_header_contains "GitHub state cookie is HttpOnly" "Set-Cookie" "HttpOnly"
  assert_header_contains "GitHub state cookie uses SameSite=Lax" "Set-Cookie" "SameSite=Lax"
}

test_callback_csrf_errors() {
  section "GitHub callback CSRF validation"

  request "GET" "/auth/github/callback"
  assert_redirect_or_error "Callback rejects missing state" "400" "INVALID_OAUTH_STATE" "invalid OAuth state"

  request "GET" "/auth/github/callback?code=fake-code&state=wrong-state" "$STATE_COOKIE_NAME=expected-state"
  assert_redirect_or_error "Callback rejects mismatched state" "400" "INVALID_OAUTH_STATE" "invalid OAuth state"
}

test_callback_denial() {
  local state
  local cookie

  section "GitHub provider denial callback"

  state="${OAUTH_LOGIN_STATE:-simulated-state}"
  cookie="${OAUTH_LOGIN_COOKIE:-$STATE_COOKIE_NAME=$state}"

  request "GET" "/auth/github/callback?error=access_denied&state=$state" "$cookie"
  assert_redirect_or_error "Callback handles denied GitHub consent" "401" "OAUTH_DENIED" "access_denied"
  assert_header_contains "Callback clears state cookie" "Set-Cookie" "$STATE_COOKIE_NAME="
}

main() {
  local failed_color
  local skipped_color

  require_command curl jq sed awk grep mktemp

  section "Configuration"
  config_line "BASE_URL" "$BASE_URL"
  config_line "CURL_TIMEOUT" "$CURL_TIMEOUT"

  test_health
  test_login_start
  test_callback_csrf_errors
  test_callback_denial

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
