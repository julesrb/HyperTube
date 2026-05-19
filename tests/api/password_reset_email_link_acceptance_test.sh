#!/usr/bin/env bash

set -uo pipefail

# Interactive acceptance test for the real password reset email link flow.
#
# This test is intentionally not part of tests/start_me or Run All because it
# depends on real email delivery and a human pasting the received link.
#
# Usage:
#   tests/api/password_reset_email_link_acceptance_test.sh
#
# Configuration:
#   BASE_URL=http://localhost:8080/api/v1
#   RESET_EMAIL=user@example.com       # optional; otherwise prompted
#   RESET_LINK=https://...token=...    # optional; otherwise prompted after email
#   NEW_PASSWORD=NewPass123!           # optional; otherwise prompted
#   RESET_LOCALE=en
#   CURL_TIMEOUT=45
#   NO_COLOR=1
#   FORCE_COLOR=1

BASE_URL="${BASE_URL:-http://localhost:8080/api/v1}"
BASE_URL="${BASE_URL%/}"
RESET_LOCALE="${RESET_LOCALE:-en}"
CURL_TIMEOUT="${CURL_TIMEOUT:-45}"
RUN_ID="$(date +%Y%m%d%H%M%S)-$$-$RANDOM"

RESET_EMAIL="${RESET_EMAIL:-}"
RESET_LINK="${RESET_LINK:-}"
NEW_PASSWORD="${NEW_PASSWORD:-}"
OLD_PASSWORD="${OLD_PASSWORD:-OldResetPass123!-$RUN_ID}"
RAW_USERNAME="reset_${RUN_ID//[^A-Za-z0-9_]/_}"
TEST_USERNAME="${RESET_USERNAME:-${RAW_USERNAME:0:32}}"
CREATED_USER=0
STOP_AFTER_SKIP=0

TMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/hypertube-password-reset-email-link.XXXXXX")"
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

usage() {
  sed -n '5,22p' "$0" | sed 's/^# \{0,1\}//'
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

prompt_password() {
  local value

  if [[ -n "$NEW_PASSWORD" ]]; then
    return 0
  fi

  if [[ ! -t 0 ]]; then
    printf 'NEW_PASSWORD is required in non-interactive mode.\n' >&2
    exit 64
  fi

  while true; do
    printf '%bNew password to set:%b ' "$BOLD" "$RESET"
    IFS= read -rs value
    printf '\n'
    value="$(trim "$value")"
    if [[ "${#value}" -ge 8 ]]; then
      NEW_PASSWORD="$value"
      return 0
    fi
    printf '%bPassword must be at least 8 characters.%b\n' "$YELLOW" "$RESET"
  done
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

error_code() {
  jq -r '.error.code // empty' "$LAST_BODY_FILE" 2>/dev/null
}

register_payload() {
  jq -n \
    --arg email "$RESET_EMAIL" \
    --arg username "$TEST_USERNAME" \
    --arg password "$OLD_PASSWORD" \
    '{email:$email, username:$username, first_name:"Reset", last_name:"Acceptance", password:$password}'
}

login_payload() {
  local password="$1"

  jq -n --arg email "$RESET_EMAIL" --arg password "$password" '{email:$email, password:$password}'
}

request_reset_payload() {
  jq -n --arg email "$RESET_EMAIL" --arg locale "$RESET_LOCALE" '{email:$email, locale:$locale}'
}

reset_password_payload() {
  local token="$1"
  local password="$2"

  jq -n --arg token "$token" --arg password "$password" '{token:$token, password:$password}'
}

extract_token_from_link() {
	local link="$1"
	local tail
	local token

  link="$(trim "$link")"
  link="${link#<}"
  link="${link%>}"
  if [[ "$link" != *token=* ]]; then
    return 1
  fi

  tail="${link#*token=}"
  token="${tail%%[&#]*}"
  token="$(trim "$token")"
  if [[ -z "$token" ]]; then
    return 1
  fi

	printf '%s' "$token"
}

resolve_effective_url() {
	local link="$1"
	local error_file
	local effective_url
	local curl_exit

	error_file="$TMP_DIR/resolve_link_error"
	: >"$error_file"

	effective_url="$(curl \
		--silent \
		--show-error \
		--location \
		--max-redirs 10 \
		--max-time "$CURL_TIMEOUT" \
		--output /dev/null \
		--write-out '%{url_effective}' \
		"$link" 2>"$error_file")"
	curl_exit=$?

	if [[ -n "$effective_url" ]]; then
		printf '%s' "$effective_url"
		return 0
	fi

	if [[ "$curl_exit" -ne 0 ]]; then
		if [[ -s "$error_file" ]]; then
			sed 's/^/    curl: /' "$error_file" >&2
		fi
		return 1
	fi

	printf '%s' "$effective_url"
}

create_or_reuse_user() {
  local payload

  section "1. Create or reuse the user"
  payload="$(register_payload)"
  request "POST" "/auth/register" "$payload"

  if [[ "$LAST_CURL_EXIT" -ne 0 ]]; then
    fail "Register user: curl failed"
    return 1
  fi

  case "$LAST_STATUS" in
    201)
      CREATED_USER=1
      pass "Created demo user for $RESET_EMAIL"
      ;;
    409)
      CREATED_USER=0
      pass "User already exists; no duplicate user was created"
      ;;
    *)
      fail "Register user: expected HTTP 201 or 409, got ${LAST_STATUS:-<none>}"
      return 1
      ;;
  esac
}

request_password_reset_email() {
  local payload

  section "2. Request the password reset email"
  payload="$(request_reset_payload)"
  request "POST" "/auth/password-reset" "$payload"

  if [[ "$LAST_CURL_EXIT" -ne 0 ]]; then
    fail "Request reset email: curl failed"
    return 1
  fi

  if [[ "$LAST_STATUS" == "202" ]]; then
    pass "Password reset request accepted"
    return 0
  fi

  if [[ "$LAST_STATUS" == "503" && "$(error_code)" == "EMAIL_NOT_CONFIGURED" ]]; then
    skip "Password reset email delivery" "EMAIL_NOT_CONFIGURED; configure BREVO_API_KEY and MAIL_FROM_EMAIL, then rerun"
    STOP_AFTER_SKIP=1
    return 0
  fi

  fail "Request reset email: expected HTTP 202, got ${LAST_STATUS:-<none>}"
  return 1
}

read_reset_link_and_extract_token() {
	local effective_url
	local token

	section "3. Paste the email link"
	printf 'Open the password reset email for %b%s%b and paste the full link here.\n' "$BOLD" "$RESET_EMAIL" "$RESET"

	prompt_required RESET_LINK "Reset link:"
	if token="$(extract_token_from_link "$RESET_LINK")"; then
		RESET_TOKEN="$token"
		pass "Extracted reset token from pasted link"
		return 0
	fi

	printf '  %bINFO%b Pasted link has no token parameter; following redirects in case this is an email tracking link.\n' "$CYAN" "$RESET"
	if ! effective_url="$(resolve_effective_url "$RESET_LINK")"; then
		result_line "FAIL" "$RED" "Could not resolve the pasted link"
		FAILED=$((FAILED + 1))
		return 1
	fi

	if ! token="$(extract_token_from_link "$effective_url")"; then
		result_line "FAIL" "$RED" "Could not extract token=... from the pasted link"
		FAILED=$((FAILED + 1))
		return 1
	fi

	RESET_TOKEN="$token"
	pass "Resolved tracking link and extracted reset token"
}

set_new_password() {
  local payload

  section "4. Set the new password"
  prompt_password

  payload="$(reset_password_payload "$RESET_TOKEN" "$NEW_PASSWORD")"
  request "POST" "/auth/reset-password" "$payload"

  if [[ "$LAST_CURL_EXIT" -ne 0 ]]; then
    fail "Reset password: curl failed"
    return 1
  fi

  if [[ "$LAST_STATUS" == "200" ]]; then
    pass "Password reset token accepted and password changed"
    return 0
  fi

  fail "Reset password: expected HTTP 200, got ${LAST_STATUS:-<none>}"
  return 1
}

verify_login_with_new_password() {
  local payload

  section "5. Verify login with the new password"
  payload="$(login_payload "$NEW_PASSWORD")"
  request "POST" "/auth/login" "$payload"

  if [[ "$LAST_CURL_EXIT" -ne 0 ]]; then
    fail "Login with new password: curl failed"
    return 1
  fi

  if [[ "$LAST_STATUS" == "200" ]]; then
    pass "Login succeeds with the new password"
    return 0
  fi

  fail "Login with new password: expected HTTP 200, got ${LAST_STATUS:-<none>}"
  return 1
}

verify_old_password_if_created() {
  local payload

  section "6. Verify old password behavior"
  if [[ "$CREATED_USER" -ne 1 ]]; then
    skip "Old password rejection" "the user already existed, so the old password is not known"
    return 0
  fi

  payload="$(login_payload "$OLD_PASSWORD")"
  request "POST" "/auth/login" "$payload"

  if [[ "$LAST_CURL_EXIT" -ne 0 ]]; then
    fail "Login with old password: curl failed"
    return 1
  fi

  if [[ "$LAST_STATUS" == "401" ]]; then
    pass "Old password is rejected after reset"
    return 0
  fi

  fail "Login with old password: expected HTTP 401, got ${LAST_STATUS:-<none>}"
  return 1
}

verify_token_replay_is_rejected() {
  local payload

  section "7. Verify the reset token is single-use"
  payload="$(reset_password_payload "$RESET_TOKEN" "ReplayPass123!")"
  request "POST" "/auth/reset-password" "$payload"

  if [[ "$LAST_CURL_EXIT" -ne 0 ]]; then
    fail "Replay reset token: curl failed"
    return 1
  fi

  if [[ "$LAST_STATUS" == "400" && "$(error_code)" == "INVALID_RESET_TOKEN" ]]; then
    pass "Used reset token is rejected"
    return 0
  fi

  fail "Replay reset token: expected INVALID_RESET_TOKEN, got HTTP ${LAST_STATUS:-<none>} $(error_code)"
  return 1
}

print_summary() {
  section "Summary"
  printf '  %bPassed%b  %d\n' "$GREEN" "$RESET" "$PASSED"
  printf '  %bFailed%b  %d\n' "$RED" "$RESET" "$FAILED"
  printf '  %bSkipped%b %d\n' "$YELLOW" "$RESET" "$SKIPPED"
  printf '  %bBASE_URL%b %s\n' "$DIM" "$RESET" "$BASE_URL"
  printf '  %bRESET_EMAIL%b %s\n' "$DIM" "$RESET" "$RESET_EMAIL"
  printf '  %bUSERNAME%b %s\n' "$DIM" "$RESET" "$TEST_USERNAME"
}

main() {
  require_command curl jq sed

  section "Configuration"
  printf '  %bBASE_URL%b %s\n' "$DIM" "$RESET" "$BASE_URL"
  printf '  %bRESET_LOCALE%b %s\n' "$DIM" "$RESET" "$RESET_LOCALE"
  printf '  %bMode%b interactive email-link acceptance test\n' "$DIM" "$RESET"

  prompt_required RESET_EMAIL "Email address for the test user:"
  create_or_reuse_user || true

  if [[ "$FAILED" -eq 0 ]]; then
    request_password_reset_email || true
  fi
  if [[ "$FAILED" -eq 0 && "$STOP_AFTER_SKIP" -eq 0 ]]; then
    read_reset_link_and_extract_token || true
  fi
  if [[ "$FAILED" -eq 0 && "$STOP_AFTER_SKIP" -eq 0 ]]; then
    set_new_password || true
  fi
  if [[ "$FAILED" -eq 0 && "$STOP_AFTER_SKIP" -eq 0 ]]; then
    verify_login_with_new_password || true
  fi
  if [[ "$FAILED" -eq 0 && "$STOP_AFTER_SKIP" -eq 0 ]]; then
    verify_old_password_if_created || true
  fi
  if [[ "$FAILED" -eq 0 && "$STOP_AFTER_SKIP" -eq 0 ]]; then
    verify_token_replay_is_rejected || true
  fi

  print_summary
  if [[ "$FAILED" -ne 0 ]]; then
    exit 1
  fi
}

main "$@"
