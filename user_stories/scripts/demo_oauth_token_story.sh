#!/usr/bin/env bash
#
# CLI demo walkthrough for the Hypertube OAuth2 token endpoint.
#
# Usage:
#   ./user_stories/scripts/demo_oauth_token_story.sh
#
# Optional environment variables:
#   BASE_URL        API origin. Default: http://localhost:8080
#   DEMO_EMAIL      Demo user's email. Default: unique email per run.
#   DEMO_USERNAME   Demo user's username. Default: unique username per run.
#   DEMO_PASSWORD   Demo user's password. Default: DemoOAuth123!
#
# Examples:
#   BASE_URL=http://localhost:8080 ./user_stories/scripts/demo_oauth_token_story.sh
#   DEMO_USERNAME=oauth_demo DEMO_PASSWORD=DemoOAuth123! ./user_stories/scripts/demo_oauth_token_story.sh

set -o pipefail

TOTAL_STEPS=8
RUN_ID="$(date +%Y%m%d%H%M%S)-$$"

BASE_URL="${BASE_URL:-http://localhost:8080}"
BASE_URL="${BASE_URL%/}"
DEMO_EMAIL="${DEMO_EMAIL:-oauth-demo-${RUN_ID}@example.test}"
DEMO_USERNAME="${DEMO_USERNAME:-oauth_demo_${RUN_ID//[^A-Za-z0-9_]/_}}"
DEMO_USERNAME="${DEMO_USERNAME:0:32}"
DEMO_PASSWORD="${DEMO_PASSWORD:-DemoOAuth123!}"

TOKEN=""
JSON_TOKEN=""

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
  sed -n '2,15p' "$0" | sed 's/^# \{0,1\}//'
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
    *) printf '%s' "$1" ;;
  esac
}

print_summary() {
  local i

  heading 8 "Summary"
  explain "The CLI reports whether the OAuth2 token flow produced a usable bearer token and OAuth2-style errors."

  printf '\n%sResult:%s\n' "$BOLD" "$RESET"
  for i in "${!SUMMARY_NAMES[@]}"; do
    printf '  %-5s %s - %s\n' "$(status_label "${SUMMARY_STATUS[$i]}")" "${SUMMARY_NAMES[$i]}" "${SUMMARY_DETAILS[$i]}"
  done

  printf '\n%sConfiguration:%s\n' "$BOLD" "$RESET"
  printf '  BASE_URL: %s\n' "$BASE_URL"
  printf '  DEMO_EMAIL: %s\n' "$DEMO_EMAIL"
  printf '  DEMO_USERNAME: %s\n' "$DEMO_USERNAME"
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

redact_payload() {
  local payload="$1"
  local content_type="$2"

  if [[ "$content_type" == "application/json" ]] && jq . >/dev/null 2>&1 <<<"$payload"; then
    jq 'if type == "object" and has("password") then .password = "<redacted>" else . end' <<<"$payload"
  else
    printf '%s\n' "$payload" | sed -E 's/(password=)[^&]*/\1<redacted>/g'
  fi
}

print_request() {
  local method="$1"
  local url="$2"
  local payload="${3:-}"
  local content_type="${4:-application/json}"
  local authenticated="${5:-false}"

  printf '\n%sRequest:%s\n' "$BOLD" "$RESET"
  printf '  %s %s\n' "$method" "$url"

  if [[ "$authenticated" == "true" ]]; then
    printf '  Authorization: Bearer %s\n' "${TOKEN:+<stored OAuth2 token>}"
  fi

  if [[ -n "$payload" ]]; then
    printf '  Content-Type: %s\n' "$content_type"
    printf '\n%sPayload:%s\n' "$BOLD" "$RESET"
    redact_payload "$payload" "$content_type" | sed 's/^/  /'
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
  local content_type="${4:-application/json}"
  local authenticated="${5:-false}"
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
    curl_args+=(-H "Content-Type: $content_type" --data "$payload")
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

extract_access_token() {
  local body="$1"

  jq -r '.access_token // .data.access_token // empty' <<<"$body" 2>/dev/null
}

REGISTER_PAYLOAD="$(
  jq -n \
    --arg email "$DEMO_EMAIL" \
    --arg username "$DEMO_USERNAME" \
    --arg password "$DEMO_PASSWORD" \
    '{email:$email, username:$username, first_name:"OAuth", last_name:"Demo", password:$password}'
)"

FORM_TOKEN_PAYLOAD="grant_type=password&username=$(urlencode "$DEMO_USERNAME")&password=$(urlencode "$DEMO_PASSWORD")"
JSON_TOKEN_PAYLOAD="$(
  jq -n \
    --arg username "$(printf '%s' "$DEMO_EMAIL" | tr '[:lower:]' '[:upper:]')" \
    --arg password "$DEMO_PASSWORD" \
    '{grant_type:"password", username:$username, password:$password}'
)"
BAD_TOKEN_PAYLOAD="grant_type=password&username=$(urlencode "$DEMO_USERNAME")&password=$(urlencode "wrong-password")"

heading 1 "Health Check"
explain "The CLI checks whether the backend is reachable before demonstrating OAuth2 token exchange."
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

heading 2 "Register Password User"
explain "The CLI creates a normal email/password user that can later exchange credentials for an OAuth2 bearer token."
REGISTER_URL="$BASE_URL/api/v1/auth/register"
print_request "POST" "$REGISTER_URL" "$REGISTER_PAYLOAD" "application/json"
request "POST" "$REGISTER_URL" "$REGISTER_PAYLOAD" "application/json"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$CURL_ERROR"

if is_success_status "$HTTP_STATUS"; then
  record_result "Register user" "OK" "Demo user was registered."
  print_result "OK" "Demo user was registered."
elif is_conflict_status "$HTTP_STATUS"; then
  record_result "Register user" "WARN" "Demo user already exists; continuing to token exchange."
  print_result "WARN" "Demo user already exists; continuing to token exchange."
else
  record_result "Register user" "WARN" "Registration failed; token exchange will verify whether the credentials are usable."
  print_result "WARN" "Registration failed; token exchange will verify whether the credentials are usable."
fi

heading 3 "Request OAuth2 Token With Username"
explain "The CLI calls POST /oauth/token with grant_type=password as form data, which is the route expected by the subject API section."
TOKEN_URL="$BASE_URL/api/v1/oauth/token"
print_request "POST" "$TOKEN_URL" "$FORM_TOKEN_PAYLOAD" "application/x-www-form-urlencoded"
request "POST" "$TOKEN_URL" "$FORM_TOKEN_PAYLOAD" "application/x-www-form-urlencoded"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$CURL_ERROR"

if is_success_status "$HTTP_STATUS"; then
  TOKEN="$(extract_access_token "$HTTP_BODY")"
  if [[ -n "$TOKEN" ]]; then
    record_result "Form token exchange" "OK" "OAuth2 access token was returned."
    print_result "OK" "OAuth2 access token was returned."
  else
    record_result "Form token exchange" "FAIL" "Token response did not include access_token."
    print_result "FAIL" "Token response did not include access_token."
    abort_critical "No OAuth2 access token was available for the protected-route check."
  fi
else
  record_result "Form token exchange" "FAIL" "OAuth2 token request failed."
  print_result "FAIL" "OAuth2 token request failed."
  abort_critical "Password grant did not return a token."
fi

heading 4 "Use Bearer Token On Protected Route"
explain "The CLI sends the token to a protected movie search route. A VALIDATION_ERROR response still proves the token passed authentication because the request reached handler validation."
PROTECTED_URL="$BASE_URL/api/v1/movies/search"
print_request "GET" "$PROTECTED_URL" "" "application/json" "true"
request "GET" "$PROTECTED_URL" "" "application/json" "true"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$CURL_ERROR"

ERROR_CODE="$(jq -r '.error.code // empty' <<<"$HTTP_BODY" 2>/dev/null)"
if [[ "$HTTP_STATUS" == "400" && "$ERROR_CODE" == "VALIDATION_ERROR" ]]; then
  record_result "Protected route with token" "OK" "Bearer token passed auth and reached route validation."
  print_result "OK" "Bearer token passed auth and reached route validation."
elif is_success_status "$HTTP_STATUS"; then
  record_result "Protected route with token" "OK" "Bearer token passed auth and route returned success."
  print_result "OK" "Bearer token passed auth and route returned success."
else
  record_result "Protected route with token" "FAIL" "Bearer token was not accepted by the protected route."
  print_result "FAIL" "Bearer token was not accepted by the protected route."
fi

heading 5 "Request OAuth2 Token With Email JSON"
explain "The same endpoint also accepts JSON body input and normalizes email login before checking the password."
print_request "POST" "$TOKEN_URL" "$JSON_TOKEN_PAYLOAD" "application/json"
request "POST" "$TOKEN_URL" "$JSON_TOKEN_PAYLOAD" "application/json"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$CURL_ERROR"

if is_success_status "$HTTP_STATUS"; then
  JSON_TOKEN="$(extract_access_token "$HTTP_BODY")"
  if [[ -n "$JSON_TOKEN" ]]; then
    record_result "JSON token exchange" "OK" "OAuth2 token was returned for email login."
    print_result "OK" "OAuth2 token was returned for email login."
  else
    record_result "JSON token exchange" "FAIL" "JSON token response did not include access_token."
    print_result "FAIL" "JSON token response did not include access_token."
  fi
else
  record_result "JSON token exchange" "FAIL" "JSON token request failed."
  print_result "FAIL" "JSON token request failed."
fi

heading 6 "Verify Invalid Grant Error"
explain "The CLI intentionally sends a wrong password to show OAuth2-style invalid_grant errors without leaking whether the username exists."
print_request "POST" "$TOKEN_URL" "$BAD_TOKEN_PAYLOAD" "application/x-www-form-urlencoded"
request "POST" "$TOKEN_URL" "$BAD_TOKEN_PAYLOAD" "application/x-www-form-urlencoded"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$CURL_ERROR"

OAUTH_ERROR="$(jq -r '.error // empty' <<<"$HTTP_BODY" 2>/dev/null)"
if [[ "$HTTP_STATUS" == "400" && "$OAUTH_ERROR" == "invalid_grant" ]]; then
  record_result "Invalid grant" "OK" "Wrong password returned invalid_grant."
  print_result "OK" "Wrong password returned invalid_grant."
else
  record_result "Invalid grant" "FAIL" "Wrong password did not return the expected OAuth2 error."
  print_result "FAIL" "Wrong password did not return the expected OAuth2 error."
fi

heading 7 "Verify Root Alias"
explain "The CLI calls the root /oauth/token alias, matching the route form shown in the subject."
ALIAS_URL="$BASE_URL/oauth/token"
print_request "POST" "$ALIAS_URL" "grant_type=client_credentials" "application/x-www-form-urlencoded"
request "POST" "$ALIAS_URL" "grant_type=client_credentials" "application/x-www-form-urlencoded"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$CURL_ERROR"

OAUTH_ERROR="$(jq -r '.error // empty' <<<"$HTTP_BODY" 2>/dev/null)"
if [[ "$HTTP_STATUS" == "400" && "$OAUTH_ERROR" == "unsupported_grant_type" ]]; then
  record_result "Root alias" "OK" "Root /oauth/token alias is registered."
  print_result "OK" "Root /oauth/token alias is registered."
else
  record_result "Root alias" "WARN" "Root alias did not return the expected unsupported_grant_type response."
  print_result "WARN" "Root alias did not return the expected unsupported_grant_type response."
fi

print_summary
