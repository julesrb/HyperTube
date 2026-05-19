#!/usr/bin/env bash

set -uo pipefail

# API smoke/e2e contract test for password reset routes.
#
# Usage:
#   tests/api/password_reset_api_test.sh
#
# Configuration:
#   BASE_URL=http://localhost:8080/api/v1
#   CURL_TIMEOUT=45
#   NO_COLOR=1
#   FORCE_COLOR=1

BASE_URL="${BASE_URL:-http://localhost:8080/api/v1}"
BASE_URL="${BASE_URL%/}"
CURL_TIMEOUT="${CURL_TIMEOUT:-45}"

TMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/hypertube-password-reset-test.XXXXXX")"
LAST_BODY_FILE="$TMP_DIR/last_body"
LAST_HEADERS_FILE="$TMP_DIR/last_headers"
LAST_CURL_ERR_FILE="$TMP_DIR/last_curl_err"
LAST_REQUEST_METHOD=""
LAST_REQUEST_URL=""
LAST_REQUEST_BODY=""
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

redact_json_payload() {
  local payload="$1"

  if printf '%s' "$payload" | jq . >/dev/null 2>&1; then
    printf '%s' "$payload" | jq 'if type == "object" then
      (if has("password") then .password = "<redacted>" else . end)
      | (if has("token") then .token = "<redacted>" else . end)
    else . end'
  else
    printf '%s\n' "$payload"
  fi
}

dump_last_response() {
  if [[ -n "$LAST_REQUEST_URL" ]]; then
    printf '    %b%-14s%b %s %s\n' "$DIM" "Request" "$RESET" "$LAST_REQUEST_METHOD" "$LAST_REQUEST_URL"
  fi
  if [[ -n "$LAST_REQUEST_BODY" ]]; then
    printf '    %b%s%b\n' "$DIM" "Payload" "$RESET"
    redact_json_payload "$LAST_REQUEST_BODY" | sed 's/^/      /'
  fi
  if [[ "$LAST_CURL_EXIT" -ne 0 && -s "$LAST_CURL_ERR_FILE" ]]; then
    printf '    %b%s%b\n' "$RED" "Curl error" "$RESET"
    sed 's/^/      /' "$LAST_CURL_ERR_FILE"
  fi

  printf '    %b%-14s%b %s\n' "$DIM" "Status" "$RESET" "${LAST_STATUS:-<none>}"
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
  local body="${3:-}"
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

  request "$method" "$path" "$body"
  if expect_status "$name" "$expected_status"; then
    assert_error "$name" "$expected_code" "$expected_message"
  fi
}

password_reset_payload() {
  local email="$1"
  local locale="${2:-en}"

  jq -n --arg email "$email" --arg locale "$locale" '{email:$email, locale:$locale}'
}

reset_password_payload() {
  local token="$1"
  local password="$2"

  jq -n --arg token "$token" --arg password "$password" '{token:$token, password:$password}'
}

test_request_password_reset_route() {
  local payload

  section "Request password reset"

  expect_error_case \
    "Password reset request rejects missing body" \
    "POST" "/auth/password-reset" "400" "BAD_REQUEST" "invalid JSON body"

  expect_error_case \
    "Password reset request rejects invalid JSON" \
    "POST" "/auth/password-reset" "400" "BAD_REQUEST" "invalid JSON body" \
    '{'

  expect_error_case \
    "Password reset request rejects unknown JSON field" \
    "POST" "/auth/password-reset" "400" "BAD_REQUEST" "invalid JSON body" \
    '{"email":"alice@example.com","role":"admin"}'

  payload="$(password_reset_payload "not-an-email")"
  expect_error_case \
    "Password reset request rejects invalid email" \
    "POST" "/auth/password-reset" "400" "VALIDATION_ERROR" "valid email is required" \
    "$payload"

  payload="$(password_reset_payload "missing-password-reset@example.test" "de")"
  request "POST" "/auth/password-reset" "$payload"
  case "$LAST_STATUS" in
    202)
      pass "Password reset request accepts valid email without disclosing account existence: HTTP 202"
      assert_jq_eq "Password reset request returns accepted message" '.data.message' "if the email exists, a password reset link has been sent"
      ;;
    503)
      pass "Password reset request reports missing mailer when email is not configured: HTTP 503"
      assert_error "Password reset request missing mailer" "EMAIL_NOT_CONFIGURED" "password reset email is not configured"
      ;;
    *)
      fail "Password reset request: expected HTTP 202 or 503, got ${LAST_STATUS:-<none>}"
      ;;
  esac
}

test_reset_password_route() {
  local valid_format_token
  local payload

  section "Reset password with token"

  valid_format_token="unknown-reset-token-with-enough-length-123"

  expect_error_case \
    "Reset password rejects missing body" \
    "POST" "/auth/reset-password" "400" "BAD_REQUEST" "invalid JSON body"

  expect_error_case \
    "Reset password rejects invalid JSON" \
    "POST" "/auth/reset-password" "400" "BAD_REQUEST" "invalid JSON body" \
    '{'

  expect_error_case \
    "Reset password rejects unknown JSON field" \
    "POST" "/auth/reset-password" "400" "BAD_REQUEST" "invalid JSON body" \
    '{"token":"unknown-reset-token-with-enough-length-123","password":"new-password","admin":true}'

  payload="$(reset_password_payload "short" "new-password")"
  expect_error_case \
    "Reset password rejects malformed token" \
    "POST" "/auth/reset-password" "400" "INVALID_RESET_TOKEN" "password reset link is invalid or expired" \
    "$payload"

  payload="$(reset_password_payload "$valid_format_token" "short")"
  expect_error_case \
    "Reset password rejects short password before consuming a token" \
    "POST" "/auth/reset-password" "400" "VALIDATION_ERROR" "password must be between 8 and 72 bytes" \
    "$payload"

  payload="$(reset_password_payload "$valid_format_token" "new-password")"
  expect_error_case \
    "Reset password rejects unknown or expired token" \
    "POST" "/auth/reset-password" "400" "INVALID_RESET_TOKEN" "password reset link is invalid or expired" \
    "$payload"
}

print_summary() {
  echo
  printf '%b%s%b\n' "$DIM" "================================================================" "$RESET"
  printf '%bPassword reset API test summary%b\n' "${BOLD}${CYAN}" "$RESET"
  printf '%b%s%b\n' "$DIM" "================================================================" "$RESET"
  printf '  %bPassed%b  %d\n' "$GREEN" "$RESET" "$PASSED"
  printf '  %bFailed%b  %d\n' "$RED" "$RESET" "$FAILED"
  printf '  %bSkipped%b %d\n' "$YELLOW" "$RESET" "$SKIPPED"
  printf '  %bBASE_URL%b %s\n' "$DIM" "$RESET" "$BASE_URL"
}

main() {
  require_command curl jq sed

  section "Configuration"
  printf '  %b%-16s%b %s\n' "$DIM" "BASE_URL" "$RESET" "$BASE_URL"
  printf '  %b%-16s%b %s\n' "$DIM" "CURL_TIMEOUT" "$RESET" "$CURL_TIMEOUT"

  test_request_password_reset_route
  test_reset_password_route
  print_summary

  if [[ "$FAILED" -ne 0 ]]; then
    exit 1
  fi
}

main "$@"
