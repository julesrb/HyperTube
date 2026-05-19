#!/usr/bin/env bash

set -uo pipefail

# Interactive acceptance test for the real GitHub OAuth login flow.
#
# This test is intentionally not part of tests/start_me or Run All because it
# depends on real GitHub OAuth app credentials and a human completing consent in
# a browser.
#
# Usage:
#   tests/api/github_oauth_acceptance_test.sh
#
# Configuration:
#   BASE_URL=http://localhost:8080/api/v1
#   AUTH_RESULT_URL='http://localhost:4200/auth/callback#access_token=...' # optional; otherwise prompted
#   CURL_TIMEOUT=45
#   NO_COLOR=1
#   FORCE_COLOR=1

BASE_URL="${BASE_URL:-http://localhost:8080/api/v1}"
BASE_URL="${BASE_URL%/}"
CURL_TIMEOUT="${CURL_TIMEOUT:-45}"
AUTH_RESULT_URL="${AUTH_RESULT_URL:-}"

STATE_COOKIE_NAME="hypertube_oauth_github_state"
ACCESS_TOKEN=""
STOP_AFTER_SKIP=0

TMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/hypertube-github-oauth-acceptance.XXXXXX")"
LAST_BODY_FILE="$TMP_DIR/last_body"
LAST_HEADERS_FILE="$TMP_DIR/last_headers"
LAST_CURL_ERR_FILE="$TMP_DIR/last_curl_err"
LAST_REQUEST_METHOD=""
LAST_REQUEST_URL=""
LAST_REQUEST_AUTH=""
LAST_STATUS=""
LAST_CURL_EXIT=0

PASSED=0
FAILED=0
SKIPPED=0

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

usage() {
  sed -n '5,18p' "$0" | sed 's/^# \{0,1\}//'
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

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

trim() {
  local value="$1"

  value="${value#"${value%%[![:space:]]*}"}"
  value="${value%"${value##*[![:space:]]}"}"
  value="${value#<}"
  value="${value%>}"
  printf '%s' "$value"
}

prompt_required() {
  local var_name="$1"
  local label="$2"
  local value

  value="${!var_name:-}"
  if [[ -n "$value" ]]; then
    printf -v "$var_name" '%s' "$(trim "$value")"
    return 0
  fi

  if [[ ! -t 0 ]]; then
    printf '%s is required in non-interactive mode.\n' "$var_name" >&2
    exit 64
  fi

  while true; do
    printf '%b%s%b ' "$BOLD" "$label" "$RESET"
    IFS= read -r value
    value="$(trim "$value")"
    if [[ -n "$value" ]]; then
      printf -v "$var_name" '%s' "$value"
      return 0
    fi
    printf '%bPlease enter a value.%b\n' "$YELLOW" "$RESET"
  done
}

dump_last_response() {
  if [[ -n "$LAST_REQUEST_URL" ]]; then
    printf '    %b%-14s%b %s %s\n' "$DIM" "Request" "$RESET" "$LAST_REQUEST_METHOD" "$LAST_REQUEST_URL"
  fi
  if [[ -n "$LAST_REQUEST_AUTH" ]]; then
    printf '    %b%-14s%b Bearer <redacted>\n' "$DIM" "Authorization" "$RESET"
  fi
  if [[ "$LAST_CURL_EXIT" -ne 0 && -s "$LAST_CURL_ERR_FILE" ]]; then
    printf '    %b%s%b\n' "$RED" "Curl error" "$RESET"
    sed 's/^/      /' "$LAST_CURL_ERR_FILE"
  fi

  printf '    %b%-14s%b %s\n' "$DIM" "Status" "$RESET" "${LAST_STATUS:-<none>}"
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
  local token="${3:-}"
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
  LAST_REQUEST_AUTH="$token"
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

  if [[ -n "$token" ]]; then
    curl_args+=(-H "Authorization: Bearer $token")
  fi

  LAST_STATUS="$(curl "${curl_args[@]}" 2>"$LAST_CURL_ERR_FILE")"
  LAST_CURL_EXIT=$?
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

error_code() {
  jq -r '.error.code // empty' "$LAST_BODY_FILE" 2>/dev/null
}

param_from_text() {
  local text="$1"
  local name="$2"

  text="$(trim "$text")"
  printf '%s' "$text" \
    | sed 's/[?#]/\&/g' \
    | tr '&' '\n' \
    | sed -n "s/^$name=//p" \
    | head -n 1
}

looks_like_jwt() {
  [[ "$1" =~ ^[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+$ ]]
}

check_api_health() {
  section "1. Check API health"
  request "GET" "/health"

  if [[ "$LAST_CURL_EXIT" -ne 0 ]]; then
    fail "Health check: curl failed"
    return 1
  fi

  if [[ "$LAST_STATUS" == "200" ]]; then
    pass "API is reachable"
    return 0
  fi

  fail "Health check: expected HTTP 200, got ${LAST_STATUS:-<none>}"
  return 1
}

check_github_login_route() {
  local location
  local cookie_state

  section "2. Check GitHub OAuth start route"
  request "GET" "/auth/github/login"

  if [[ "$LAST_CURL_EXIT" -ne 0 ]]; then
    fail "GitHub login start: curl failed"
    return 1
  fi

  if [[ "$LAST_STATUS" == "503" && "$(error_code)" == "OAUTH_NOT_CONFIGURED" ]]; then
    skip "Real GitHub OAuth acceptance" "GITHUB_CLIENT_ID, GITHUB_CLIENT_SECRET, or GITHUB_REDIRECT_URL is not configured"
    STOP_AFTER_SKIP=1
    return 0
  fi

  if [[ "$LAST_STATUS" != "302" ]]; then
    fail "GitHub login start: expected HTTP 302 or OAUTH_NOT_CONFIGURED, got ${LAST_STATUS:-<none>}"
    return 1
  fi

  location="$(header_value "Location")"
  cookie_state="$(set_cookie_value "$STATE_COOKIE_NAME")"

  if [[ "$location" == https://github.com/login/oauth/authorize* ]]; then
    pass "GitHub login redirects to GitHub authorize endpoint"
  else
    fail "GitHub login Location does not point to GitHub authorize endpoint"
  fi

  if [[ "$location" == *"read%3Auser"* && "$location" == *"user%3Aemail"* ]]; then
    pass "GitHub login requests user identity and email scopes"
  else
    fail "GitHub login Location is missing expected scopes"
  fi

  if [[ -n "$cookie_state" ]]; then
    pass "GitHub login sets the OAuth state cookie"
  else
    fail "GitHub login did not set the OAuth state cookie"
  fi
}

read_browser_result() {
  local login_url
  local token_type
  local error
  local token

  section "3. Complete GitHub OAuth in a browser"
  login_url="$BASE_URL/auth/github/login"

  printf 'Open this URL in your browser:\n'
  printf '  %b%s%b\n\n' "$BOLD" "$login_url" "$RESET"
  printf 'After GitHub redirects back, paste the final browser URL here.\n'
  printf 'It should contain %b#access_token=...%b. If the frontend is not running, the browser can still show the localhost:4200 URL in the address bar.\n\n' "$BOLD" "$RESET"

  prompt_required AUTH_RESULT_URL "Final callback URL or raw access token:"

  error="$(param_from_text "$AUTH_RESULT_URL" "error")"
  if [[ -n "$error" ]]; then
    result_line "FAIL" "$RED" "GitHub OAuth returned error=$error"
    FAILED=$((FAILED + 1))
    return 1
  fi

  token="$(param_from_text "$AUTH_RESULT_URL" "access_token")"
  if [[ -z "$token" ]] && looks_like_jwt "$AUTH_RESULT_URL"; then
    token="$AUTH_RESULT_URL"
  fi
  token="$(trim "$token")"

  if [[ -z "$token" ]]; then
    result_line "FAIL" "$RED" "Could not extract access_token from the pasted value"
    FAILED=$((FAILED + 1))
    return 1
  fi

  if looks_like_jwt "$token"; then
    pass "Extracted app JWT from GitHub OAuth callback"
  else
    result_line "FAIL" "$RED" "Extracted access token does not look like the app JWT"
    FAILED=$((FAILED + 1))
    return 1
  fi

  token_type="$(param_from_text "$AUTH_RESULT_URL" "token_type")"
  if [[ -z "$token_type" || "$token_type" == "Bearer" ]]; then
    pass "Callback token type is Bearer"
  else
    result_line "FAIL" "$RED" "Unexpected token_type=$token_type"
    FAILED=$((FAILED + 1))
    return 1
  fi

  ACCESS_TOKEN="$token"
}

verify_protected_route_rejects_missing_token() {
  section "4. Verify protected route rejects anonymous access"
  request "GET" "/movies/search"

  if [[ "$LAST_CURL_EXIT" -ne 0 ]]; then
    fail "Anonymous protected route request: curl failed"
    return 1
  fi

  if [[ "$LAST_STATUS" == "401" && "$(error_code)" == "UNAUTHORIZED" ]]; then
    pass "Protected route rejects missing bearer token"
    return 0
  fi

  fail "Protected route without token: expected UNAUTHORIZED, got HTTP ${LAST_STATUS:-<none>} $(error_code)"
  return 1
}

verify_github_token_is_accepted() {
  section "5. Verify GitHub-issued app JWT is accepted"
  request "GET" "/movies/search" "$ACCESS_TOKEN"

  if [[ "$LAST_CURL_EXIT" -ne 0 ]]; then
    fail "Authenticated protected route request: curl failed"
    return 1
  fi

  if [[ "$LAST_STATUS" == "400" && "$(error_code)" == "VALIDATION_ERROR" ]]; then
    pass "Protected route accepted the JWT and reached handler validation"
    return 0
  fi

  fail "Protected route with GitHub JWT: expected handler validation, got HTTP ${LAST_STATUS:-<none>} $(error_code)"
  return 1
}

print_summary() {
  section "Summary"
  printf '  %bPassed%b  %d\n' "$GREEN" "$RESET" "$PASSED"
  printf '  %bFailed%b  %d\n' "$RED" "$RESET" "$FAILED"
  printf '  %bSkipped%b %d\n' "$YELLOW" "$RESET" "$SKIPPED"
  printf '  %bBASE_URL%b %s\n' "$DIM" "$RESET" "$BASE_URL"
}

main() {
  require_command curl jq sed awk grep tr

  section "Configuration"
  printf '  %bBASE_URL%b %s\n' "$DIM" "$RESET" "$BASE_URL"
  printf '  %bMode%b interactive real GitHub OAuth acceptance test\n' "$DIM" "$RESET"

  check_api_health || true
  if [[ "$FAILED" -eq 0 ]]; then
    check_github_login_route || true
  fi
  if [[ "$FAILED" -eq 0 && "$STOP_AFTER_SKIP" -eq 0 ]]; then
    read_browser_result || true
  fi
  if [[ "$FAILED" -eq 0 && "$STOP_AFTER_SKIP" -eq 0 ]]; then
    verify_protected_route_rejects_missing_token || true
  fi
  if [[ "$FAILED" -eq 0 && "$STOP_AFTER_SKIP" -eq 0 ]]; then
    verify_github_token_is_accepted || true
  fi

  print_summary
  if [[ "$FAILED" -ne 0 ]]; then
    exit 1
  fi
}

main "$@"
