#!/usr/bin/env bash
#
# CLI demo walkthrough for the GitHub authentication user story.
#
# User story:
#   As a visitor, I want to choose GitHub sign-in, be redirected to GitHub
#   authorization, and return safely to the app if authorization is cancelled
#   or malformed.
#
# Usage:
#   ./user_stories/scripts/demo_github_auth_story.sh
#
# Optional environment variables:
#   BASE_URL      API origin. Default: http://localhost:8080
#   CURL_TIMEOUT  Curl timeout in seconds. Default: 20
#
# Examples:
#   BASE_URL=http://localhost:8080 ./user_stories/scripts/demo_github_auth_story.sh
#   CURL_TIMEOUT=10 ./user_stories/scripts/demo_github_auth_story.sh

set -o pipefail

TOTAL_STEPS=7
RUN_ID="$(date +%Y%m%d%H%M%S)-$$"

BASE_URL="${BASE_URL:-http://localhost:8080}"
BASE_URL="${BASE_URL%/}"
CURL_TIMEOUT="${CURL_TIMEOUT:-20}"
STATE_COOKIE_NAME="hypertube_oauth_github_state"

OAUTH_STATE=""
STATE_COOKIE=""
AUTHORIZATION_URL=""

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
  sed -n '2,20p' "$0" | sed 's/^# \{0,1\}//'
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
require_command sed

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

  heading 7 "Summary"
  explain "The CLI reports which GitHub auth story checks succeeded, failed, or were skipped."

  printf '\n%sResult:%s\n' "$BOLD" "$RESET"
  for i in "${!SUMMARY_NAMES[@]}"; do
    printf '  %-5s %s - %s\n' "$(status_label "${SUMMARY_STATUS[$i]}")" "${SUMMARY_NAMES[$i]}" "${SUMMARY_DETAILS[$i]}"
  done

  printf '\n%sConfiguration:%s\n' "$BOLD" "$RESET"
  printf '  BASE_URL: %s\n' "$BASE_URL"
  printf '  CURL_TIMEOUT: %s\n' "$CURL_TIMEOUT"
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

is_redirect_status() {
  [[ "$1" =~ ^30[12378]$ ]]
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
  local cookie="${3:-}"

  printf '\n%sRequest:%s\n' "$BOLD" "$RESET"
  printf '  %s %s\n' "$method" "$url"

  if [[ -n "$cookie" ]]; then
    printf '  Cookie: %s\n' "$cookie"
  fi
}

print_response() {
  local status="$1"
  local body="$2"
  local headers="$3"
  local curl_error="${4:-}"
  local location
  local set_cookie

  location="$(header_value_from_text "$headers" "Location")"
  set_cookie="$(header_value_from_text "$headers" "Set-Cookie")"

  printf '\n%sResponse:%s\n' "$BOLD" "$RESET"
  printf '  HTTP %s\n' "$status"

  if [[ -n "$location" ]]; then
    printf '  Location: %s\n' "$location"
  fi
  if [[ -n "$set_cookie" ]]; then
    printf '  Set-Cookie: %s\n' "$set_cookie"
  fi
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
  local cookie="${3:-}"
  local body_file
  local header_file
  local error_file
  local curl_args

  HTTP_STATUS=""
  HTTP_BODY=""
  HTTP_HEADERS=""
  CURL_ERROR=""

  body_file="$(mktemp)"
  header_file="$(mktemp)"
  error_file="$(mktemp)"
  TMP_FILES+=("$body_file" "$header_file" "$error_file")

  curl_args=(
    -sS
    --max-time "$CURL_TIMEOUT"
    -o "$body_file"
    -D "$header_file"
    -w "%{http_code}"
    -X "$method"
    -H "Accept: application/json"
  )

  if [[ -n "$cookie" ]]; then
    curl_args+=(-H "Cookie: $cookie")
  fi

  if ! HTTP_STATUS="$(curl "${curl_args[@]}" "$url" 2>"$error_file")"; then
    CURL_ERROR="$(<"$error_file")"
    HTTP_STATUS="${HTTP_STATUS:-000}"
  fi

  HTTP_BODY="$(<"$body_file")"
  HTTP_HEADERS="$(sed 's/\r$//' "$header_file")"
}

header_value_from_text() {
  local headers="$1"
  local header="$2"

  awk -v header="$header" 'BEGIN { wanted = tolower(header) ":" } tolower($0) ~ "^" wanted { sub(/^[^:]*:[[:space:]]*/, ""); print; exit }' <<<"$headers"
}

cookie_value_from_text() {
  local headers="$1"
  local cookie_name="$2"

  grep -i '^Set-Cookie:' <<<"$headers" \
    | sed -n "s/^Set-Cookie:[[:space:]]*$cookie_name=\\([^;]*\\).*/\\1/ip" \
    | head -n 1
}

extract_state_from_location() {
  local location="$1"

  sed -n 's/.*[?&]state=\([^&#]*\).*/\1/p' <<<"$location" | head -n 1
}

response_has_error_code() {
  local code="$1"
  local location

  location="$(header_value_from_text "$HTTP_HEADERS" "Location")"

  if [[ "$location" == *"error=$code"* ]]; then
    return 0
  fi

  jq -e --arg code "$code" '.error.code == $code' <<<"$HTTP_BODY" >/dev/null 2>&1
}

heading 1 "Health Check"
explain "The CLI checks whether the backend is reachable before starting the GitHub auth walkthrough."
HEALTH_URL="$BASE_URL/api/v1/health"
print_request "GET" "$HEALTH_URL"
request "GET" "$HEALTH_URL"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$HTTP_HEADERS" "$CURL_ERROR"

if is_success_status "$HTTP_STATUS"; then
  record_result "Health check" "OK" "Backend is reachable."
  print_result "OK" "Backend is reachable."
else
  record_result "Health check" "FAIL" "Backend did not return a successful health response."
  print_result "FAIL" "Backend did not return a successful health response."
  abort_critical "Health check failed; the API must be running before the demo can continue."
fi

heading 2 "Start GitHub Sign-In"
explain "The CLI calls the backend route that the frontend uses when a visitor clicks Sign In with GitHub."
LOGIN_URL="$BASE_URL/api/v1/auth/github/login"
print_request "GET" "$LOGIN_URL"
request "GET" "$LOGIN_URL"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$HTTP_HEADERS" "$CURL_ERROR"

if [[ "$HTTP_STATUS" == "302" ]]; then
  AUTHORIZATION_URL="$(header_value_from_text "$HTTP_HEADERS" "Location")"
  OAUTH_STATE="$(extract_state_from_location "$AUTHORIZATION_URL")"
  STATE_COOKIE="$STATE_COOKIE_NAME=$(cookie_value_from_text "$HTTP_HEADERS" "$STATE_COOKIE_NAME")"
  if [[ -n "$AUTHORIZATION_URL" && -n "$OAUTH_STATE" && "$STATE_COOKIE" != "$STATE_COOKIE_NAME=" ]]; then
    record_result "Start GitHub sign-in" "OK" "Backend issued a GitHub authorization redirect and state cookie."
    print_result "OK" "Backend issued a GitHub authorization redirect and state cookie."
  else
    record_result "Start GitHub sign-in" "FAIL" "Redirect response was missing Location, state, or state cookie."
    print_result "FAIL" "Redirect response was missing Location, state, or state cookie."
    abort_critical "The OAuth start response was incomplete."
  fi
elif [[ "$HTTP_STATUS" == "503" && "$(jq -r '.error.code // empty' <<<"$HTTP_BODY" 2>/dev/null)" == "OAUTH_NOT_CONFIGURED" ]]; then
  OAUTH_STATE="simulated-state-${RUN_ID//[^A-Za-z0-9]/}"
  STATE_COOKIE="$STATE_COOKIE_NAME=$OAUTH_STATE"
  record_result "Start GitHub sign-in" "WARN" "GitHub provider credentials are not configured; continuing with callback safety checks."
  print_result "WARN" "GitHub provider credentials are not configured; continuing with callback safety checks."
else
  record_result "Start GitHub sign-in" "FAIL" "Unexpected response from GitHub login start."
  print_result "FAIL" "Unexpected response from GitHub login start."
  abort_critical "The GitHub login start route did not behave as expected."
fi

heading 3 "Inspect Authorization Redirect"
explain "The CLI verifies that the redirect points at the official GitHub authorization endpoint with identity and email scopes."

if [[ -n "$AUTHORIZATION_URL" ]]; then
  printf '\n%sAuthorization URL:%s\n' "$BOLD" "$RESET"
  printf '  %s\n' "$AUTHORIZATION_URL"

  if [[ "$AUTHORIZATION_URL" == https://github.com/login/oauth/authorize* && "$AUTHORIZATION_URL" == *"response_type=code"* && "$AUTHORIZATION_URL" == *"state=$OAUTH_STATE"* && "$AUTHORIZATION_URL" == *"read%3Auser"* && "$AUTHORIZATION_URL" == *"user%3Aemail"* ]]; then
    record_result "Inspect authorization redirect" "OK" "Redirect targets GitHub OAuth with code flow, state, and email scope."
    print_result "OK" "Redirect targets GitHub OAuth with code flow, state, and email scope."
  else
    record_result "Inspect authorization redirect" "FAIL" "Authorization redirect did not include the expected endpoint or query parameters."
    print_result "FAIL" "Authorization redirect did not include the expected endpoint or query parameters."
    abort_critical "The GitHub authorization redirect was malformed."
  fi
else
  record_result "Inspect authorization redirect" "SKIP" "Skipped because GitHub credentials are not configured."
  print_result "SKIP" "Skipped because GitHub credentials are not configured."
fi

heading 4 "Cancel GitHub Consent"
explain "The CLI simulates GitHub redirecting back with an access_denied error and the correct state cookie."
CANCEL_URL="$BASE_URL/api/v1/auth/github/callback?error=access_denied&state=$OAUTH_STATE"
print_request "GET" "$CANCEL_URL" "$STATE_COOKIE"
request "GET" "$CANCEL_URL" "$STATE_COOKIE"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$HTTP_HEADERS" "$CURL_ERROR"

if { is_redirect_status "$HTTP_STATUS" || [[ "$HTTP_STATUS" == "401" ]]; } && response_has_error_code "OAUTH_DENIED"; then
  record_result "Cancel GitHub consent" "OK" "Backend returned an OAuth denied error for the frontend to display."
  print_result "OK" "Backend returned an OAuth denied error for the frontend to display."
else
  record_result "Cancel GitHub consent" "FAIL" "Backend did not return the expected OAuth denied error."
  print_result "FAIL" "Backend did not return the expected OAuth denied error."
  abort_critical "The cancelled GitHub callback was not handled safely."
fi

heading 5 "Reject Missing State"
explain "The CLI calls the callback without code or state to prove the CSRF state check is enforced."
MISSING_STATE_URL="$BASE_URL/api/v1/auth/github/callback"
print_request "GET" "$MISSING_STATE_URL"
request "GET" "$MISSING_STATE_URL"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$HTTP_HEADERS" "$CURL_ERROR"

if { is_redirect_status "$HTTP_STATUS" || [[ "$HTTP_STATUS" == "400" ]]; } && response_has_error_code "INVALID_OAUTH_STATE"; then
  record_result "Reject missing state" "OK" "Callback without state was rejected."
  print_result "OK" "Callback without state was rejected."
else
  record_result "Reject missing state" "FAIL" "Callback without state was not rejected as expected."
  print_result "FAIL" "Callback without state was not rejected as expected."
  abort_critical "Missing OAuth state was accepted."
fi

heading 6 "Reject Mismatched State"
explain "The CLI sends a callback state that does not match the state cookie."
MISMATCH_URL="$BASE_URL/api/v1/auth/github/callback?code=fake-code&state=wrong-state"
print_request "GET" "$MISMATCH_URL" "$STATE_COOKIE_NAME=expected-state"
request "GET" "$MISMATCH_URL" "$STATE_COOKIE_NAME=expected-state"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$HTTP_HEADERS" "$CURL_ERROR"

if { is_redirect_status "$HTTP_STATUS" || [[ "$HTTP_STATUS" == "400" ]]; } && response_has_error_code "INVALID_OAUTH_STATE"; then
  record_result "Reject mismatched state" "OK" "Callback with mismatched state was rejected."
  print_result "OK" "Callback with mismatched state was rejected."
else
  record_result "Reject mismatched state" "FAIL" "Callback with mismatched state was not rejected as expected."
  print_result "FAIL" "Callback with mismatched state was not rejected as expected."
  abort_critical "Mismatched OAuth state was accepted."
fi

print_summary
