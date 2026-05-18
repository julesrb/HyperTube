#!/usr/bin/env bash
#
# Interactive CLI walkthrough for the real 42 OAuth login flow.
#
# User story:
#   As a developer without a connected frontend, I want to start 42 OAuth from
#   the command line, finish consent in the browser, paste the callback URL back
#   into the terminal, and use the returned JWT against a protected API route.
#
# Usage:
#   ./user_stories/scripts/demo_42_cli_login_story.sh
#
# Recommended CLI-only 42 app redirect URI:
#   http://localhost:8080/api/v1/auth/42/manual-copy
#
# Set the same value in .env:
#   FORTYTWO_REDIRECT_URL=http://localhost:8080/api/v1/auth/42/manual-copy
#
# The manual-copy path does not need to exist. It is only used so the browser
# stops on a URL that contains ?code=...&state=... for you to paste here. This
# script then calls the real backend callback route with the saved state cookie.
#
# Optional environment variables:
#   BASE_URL             API origin. Default: http://localhost:8080
#   CURL_TIMEOUT         Curl timeout in seconds. Default: 20
#   SEARCH_QUERY         Protected movie search query. Default: dune
#   AUTO_ADVANCE=1       Do not pause between steps.
#   OAUTH_CALLBACK_URL   Callback URL pasted non-interactively.
#
# Examples:
#   ./user_stories/scripts/demo_42_cli_login_story.sh
#   SEARCH_QUERY=matrix ./user_stories/scripts/demo_42_cli_login_story.sh
#   AUTO_ADVANCE=1 OAUTH_CALLBACK_URL='http://localhost:8080/...?...' ./user_stories/scripts/demo_42_cli_login_story.sh

set -o pipefail

TOTAL_STEPS=8

BASE_URL="${BASE_URL:-http://localhost:8080}"
BASE_URL="${BASE_URL%/}"
CURL_TIMEOUT="${CURL_TIMEOUT:-20}"
SEARCH_QUERY="${SEARCH_QUERY:-dune}"
AUTO_ADVANCE="${AUTO_ADVANCE:-0}"
OAUTH_CALLBACK_URL="${OAUTH_CALLBACK_URL:-}"

STATE_COOKIE_NAME="hypertube_oauth_42_state"
LOGIN_URL="$BASE_URL/api/v1/auth/42/login"
REAL_CALLBACK_URL="$BASE_URL/api/v1/auth/42/callback"
HEALTH_URL="$BASE_URL/api/v1/health"

AUTHORIZATION_URL=""
AUTHORIZATION_STATE=""
AUTHORIZATION_REDIRECT_URI=""
CALLBACK_CODE=""
CALLBACK_STATE=""
TOKEN=""
TOKEN_USER_JSON=""

HTTP_STATUS=""
HTTP_BODY=""
HTTP_HEADERS=""
CURL_ERROR=""

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
  sed -n '2,33p' "$0" | sed 's/^# \{0,1\}//'
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

COOKIE_JAR="$(mktemp)"
TMP_FILES+=("$COOKIE_JAR")

heading() {
  local number="$1"
  local title="$2"
  printf '\n%s[%s/%s] %s%s\n' "$BOLD$CYAN" "$number" "$TOTAL_STEPS" "$title" "$RESET"
}

explain() {
  printf '%sWhat happens:%s\n' "$BOLD" "$RESET"
  printf '  %s\n' "$1"
}

pause_step() {
  if [[ "$AUTO_ADVANCE" == "1" || ! -t 0 ]]; then
    return 0
  fi
  printf '\n%sPress Enter to continue.%s' "$DIM" "$RESET"
  read -r _
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

  heading 8 "Summary"
  explain "The CLI reports which parts of the 42 login walkthrough succeeded, failed, or were skipped."

  printf '\n%sResult:%s\n' "$BOLD" "$RESET"
  for i in "${!SUMMARY_NAMES[@]}"; do
    printf '  %-5s %s - %s\n' "$(status_label "${SUMMARY_STATUS[$i]}")" "${SUMMARY_NAMES[$i]}" "${SUMMARY_DETAILS[$i]}"
  done

  printf '\n%sConfiguration:%s\n' "$BOLD" "$RESET"
  printf '  BASE_URL: %s\n' "$BASE_URL"
  printf '  SEARCH_QUERY: %s\n' "$SEARCH_QUERY"
  if [[ -n "$AUTHORIZATION_REDIRECT_URI" ]]; then
    printf '  OAuth redirect_uri: %s\n' "$AUTHORIZATION_REDIRECT_URI"
  fi

  if [[ -n "$TOKEN" ]]; then
    printf '\n%sReusable token:%s\n' "$BOLD" "$RESET"
    printf "  export HYPERTUBE_TOKEN='%s'\n" "$TOKEN"
    printf "  curl -H \"Authorization: Bearer \$HYPERTUBE_TOKEN\" '%s/api/v1/movies/search?title=%s'\n" "$BASE_URL" "$(urlencode "$SEARCH_QUERY")"
  fi
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

urldecode() {
  local value="${1//+/ }"
  printf '%b' "${value//%/\\x}"
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

header_value_from_text() {
  local headers="$1"
  local header="$2"

  awk -v header="$header" 'BEGIN { wanted = tolower(header) ":" } tolower($0) ~ "^" wanted { sub(/^[^:]*:[[:space:]]*/, ""); print; exit }' <<<"$headers"
}

url_query_param() {
  local url="$1"
  local param="$2"
  local query

  query="${url#*\?}"
  query="${query%%#*}"
  if [[ "$query" == "$url" ]]; then
    return 0
  fi

  awk -v key="$param" '
    BEGIN { RS = "&" }
    {
      split($0, parts, "=")
      if (parts[1] == key) {
        sub(/^[^=]*=/, "")
        print
        exit
      }
    }
  ' <<<"$query"
}

url_fragment_param() {
  local url="$1"
  local param="$2"
  local fragment

  fragment="${url#*#}"
  if [[ "$fragment" == "$url" ]]; then
    return 0
  fi

  awk -v key="$param" '
    BEGIN { RS = "&" }
    {
      split($0, parts, "=")
      if (parts[1] == key) {
        sub(/^[^=]*=/, "")
        print
        exit
      }
    }
  ' <<<"$fragment"
}

print_request() {
  local method="$1"
  local url="$2"
  local authenticated="${3:-false}"

  printf '\n%sRequest:%s\n' "$BOLD" "$RESET"
  printf '  %s %s\n' "$method" "$url"
  if [[ "$authenticated" == "true" ]]; then
    printf '  Authorization: Bearer %s\n' "${TOKEN:+<stored JWT>}"
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
  local cookie_mode="${3:-none}"
  local authenticated="${4:-false}"
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

  case "$cookie_mode" in
    save) curl_args+=(-c "$COOKIE_JAR") ;;
    send) curl_args+=(-b "$COOKIE_JAR") ;;
    none) ;;
    *) curl_args+=(-H "Cookie: $cookie_mode") ;;
  esac

  if [[ "$authenticated" == "true" ]]; then
    curl_args+=(-H "Authorization: Bearer $TOKEN")
  fi

  if ! HTTP_STATUS="$(curl "${curl_args[@]}" "$url" 2>"$error_file")"; then
    CURL_ERROR="$(<"$error_file")"
    HTTP_STATUS="${HTTP_STATUS:-000}"
  fi

  HTTP_BODY="$(<"$body_file")"
  HTTP_HEADERS="$(sed 's/\r$//' "$header_file")"
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

extract_token_from_callback_response() {
  local location
  local encoded_token
  local encoded_user

  TOKEN=""
  TOKEN_USER_JSON=""

  if jq -e '.data.access_token' <<<"$HTTP_BODY" >/dev/null 2>&1; then
    TOKEN="$(jq -r '.data.access_token' <<<"$HTTP_BODY")"
    TOKEN_USER_JSON="$(jq -c '.data.user // empty' <<<"$HTTP_BODY")"
    return 0
  fi

  location="$(header_value_from_text "$HTTP_HEADERS" "Location")"
  encoded_token="$(url_fragment_param "$location" "access_token")"
  encoded_user="$(url_fragment_param "$location" "user")"

  if [[ -n "$encoded_token" ]]; then
    TOKEN="$(urldecode "$encoded_token")"
  fi
  if [[ -n "$encoded_user" ]]; then
    TOKEN_USER_JSON="$(urldecode "$encoded_user")"
  fi
}

heading 1 "Setup"
explain "This flow keeps the backend OAuth state cookie in curl, while you complete the 42 consent screen in a browser."
printf '\n%sBefore running the live part, use a CLI-only redirect URI in your 42 app and .env:%s\n' "$BOLD" "$RESET"
printf '  FORTYTWO_REDIRECT_URL=%s/api/v1/auth/42/manual-copy\n' "$BASE_URL"
printf '\n%sWhy:%s the browser will stop on that URL with code and state, and this script will call the real callback with the curl cookie jar.\n' "$BOLD" "$RESET"
pause_step
record_result "Setup instructions" "OK" "CLI redirect instructions were shown."

heading 2 "Health Check"
explain "The CLI checks whether the backend is reachable."
print_request "GET" "$HEALTH_URL"
request "GET" "$HEALTH_URL"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$HTTP_HEADERS" "$CURL_ERROR"

if is_success_status "$HTTP_STATUS"; then
  record_result "Health check" "OK" "Backend is reachable."
  print_result "OK" "Backend is reachable."
else
  record_result "Health check" "FAIL" "Backend did not return a successful health response."
  print_result "FAIL" "Backend did not return a successful health response."
  abort_critical "Start the API before running the 42 login story."
fi
pause_step

heading 3 "Start 42 Sign-In"
explain "The CLI calls the backend login route, stores the OAuth state cookie, and captures the 42 authorization URL."
print_request "GET" "$LOGIN_URL"
request "GET" "$LOGIN_URL" "save"
print_response "$HTTP_STATUS" "$HTTP_BODY" "$HTTP_HEADERS" "$CURL_ERROR"

if [[ "$HTTP_STATUS" == "503" && "$(jq -r '.error.code // empty' <<<"$HTTP_BODY" 2>/dev/null)" == "OAUTH_NOT_CONFIGURED" ]]; then
  record_result "Start 42 sign-in" "FAIL" "42 OAuth credentials are missing from the backend environment."
  print_result "FAIL" "Set FORTYTWO_CLIENT_ID, FORTYTWO_CLIENT_SECRET, and FORTYTWO_REDIRECT_URL in .env, then restart the API."
  abort_critical "42 OAuth is not configured."
fi

if [[ "$HTTP_STATUS" != "302" ]]; then
  record_result "Start 42 sign-in" "FAIL" "Backend did not return the expected 302 authorization redirect."
  print_result "FAIL" "Backend did not return the expected 302 authorization redirect."
  abort_critical "The OAuth start route did not behave as expected."
fi

AUTHORIZATION_URL="$(header_value_from_text "$HTTP_HEADERS" "Location")"
AUTHORIZATION_STATE="$(urldecode "$(url_query_param "$AUTHORIZATION_URL" "state")")"
AUTHORIZATION_REDIRECT_URI="$(urldecode "$(url_query_param "$AUTHORIZATION_URL" "redirect_uri")")"

if [[ -z "$AUTHORIZATION_URL" || -z "$AUTHORIZATION_STATE" ]]; then
  record_result "Start 42 sign-in" "FAIL" "Authorization redirect was missing Location or state."
  print_result "FAIL" "Authorization redirect was missing Location or state."
  abort_critical "The OAuth start response was incomplete."
fi

record_result "Start 42 sign-in" "OK" "Backend issued a 42 authorization redirect and saved the state cookie."
print_result "OK" "Backend issued a 42 authorization redirect and saved the state cookie."
pause_step

heading 4 "Browser Consent"
explain "Open the authorization URL in your browser, log in at 42, then paste the final manual-copy callback URL here."
printf '\n%sAuthorization URL:%s\n' "$BOLD" "$RESET"
printf '  %s\n' "$AUTHORIZATION_URL"
printf '\n%sExpected redirect URI:%s\n' "$BOLD" "$RESET"
printf '  %s\n' "${AUTHORIZATION_REDIRECT_URI:-unknown}"

if [[ "$AUTHORIZATION_REDIRECT_URI" == "$REAL_CALLBACK_URL" ]]; then
  printf '\n%sWarning:%s Your redirect_uri is the real backend callback. For CLI use, set it to %s/api/v1/auth/42/manual-copy and restart the API.\n' "$YELLOW" "$RESET" "$BASE_URL"
fi

if [[ -z "$OAUTH_CALLBACK_URL" ]]; then
  if [[ -t 0 ]]; then
    printf '\nPaste callback URL containing code and state, or press Enter to skip the live exchange:\n> '
    read -r OAUTH_CALLBACK_URL
  fi
fi

if [[ -z "$OAUTH_CALLBACK_URL" ]]; then
  record_result "Browser consent" "SKIP" "No callback URL was pasted; live token exchange was skipped."
  print_result "SKIP" "No callback URL was pasted; live token exchange was skipped."
else
  CALLBACK_CODE="$(urldecode "$(url_query_param "$OAUTH_CALLBACK_URL" "code")")"
  CALLBACK_STATE="$(urldecode "$(url_query_param "$OAUTH_CALLBACK_URL" "state")")"
  if [[ -z "$CALLBACK_CODE" || -z "$CALLBACK_STATE" ]]; then
    record_result "Browser consent" "FAIL" "Pasted URL did not contain code and state."
    print_result "FAIL" "Pasted URL did not contain code and state."
    abort_critical "Paste the redirect URL that contains both code and state."
  fi
  if [[ "$CALLBACK_STATE" != "$AUTHORIZATION_STATE" ]]; then
    record_result "Browser consent" "FAIL" "Pasted state does not match the state from the login start response."
    print_result "FAIL" "Pasted state does not match the state from the login start response."
    abort_critical "The pasted callback URL belongs to a different OAuth attempt."
  fi
  record_result "Browser consent" "OK" "Callback code and matching state were captured."
  print_result "OK" "Callback code and matching state were captured."
fi
pause_step

heading 5 "Exchange Code"
explain "The CLI calls the real backend callback route with the saved curl state cookie."

if [[ -z "$CALLBACK_CODE" ]]; then
  record_result "Exchange code" "SKIP" "Skipped because no callback code was provided."
  print_result "SKIP" "Skipped because no callback code was provided."
else
  EXCHANGE_URL="$REAL_CALLBACK_URL?code=$(urlencode "$CALLBACK_CODE")&state=$(urlencode "$CALLBACK_STATE")"
  print_request "GET" "$EXCHANGE_URL"
  request "GET" "$EXCHANGE_URL" "send"
  print_response "$HTTP_STATUS" "$HTTP_BODY" "$HTTP_HEADERS" "$CURL_ERROR"
  extract_token_from_callback_response

  if { is_success_status "$HTTP_STATUS" || is_redirect_status "$HTTP_STATUS"; } && [[ -n "$TOKEN" ]]; then
    record_result "Exchange code" "OK" "Backend exchanged the 42 code and returned a JWT."
    print_result "OK" "Backend exchanged the 42 code and returned a JWT."
    if [[ -n "$TOKEN_USER_JSON" ]]; then
      printf '\n%sOAuth user:%s\n' "$BOLD" "$RESET"
      if jq -e . >/dev/null 2>&1 <<<"$TOKEN_USER_JSON"; then
        jq . <<<"$TOKEN_USER_JSON" | sed 's/^/  /'
      else
        printf '  %s\n' "$TOKEN_USER_JSON"
      fi
    fi
  elif response_has_error_code "OAUTH_EXCHANGE_FAILED"; then
    record_result "Exchange code" "FAIL" "42 rejected the authorization code exchange."
    print_result "FAIL" "Check that FORTYTWO_REDIRECT_URL exactly matches the redirect URI registered in the 42 app."
    abort_critical "42 authorization code exchange failed."
  else
    record_result "Exchange code" "FAIL" "Backend did not return a JWT."
    print_result "FAIL" "Backend did not return a JWT."
    abort_critical "The OAuth callback did not complete successfully."
  fi
fi
pause_step

heading 6 "Use JWT"
explain "The CLI uses the returned JWT against a protected movie search route."

if [[ -z "$TOKEN" ]]; then
  record_result "Use JWT" "SKIP" "Skipped because no JWT was created."
  print_result "SKIP" "Skipped because no JWT was created."
else
  SEARCH_URL="$BASE_URL/api/v1/movies/search?title=$(urlencode "$SEARCH_QUERY")"
  print_request "GET" "$SEARCH_URL" "true"
  request "GET" "$SEARCH_URL" "none" "true"
  print_response "$HTTP_STATUS" "$HTTP_BODY" "$HTTP_HEADERS" "$CURL_ERROR"

  if is_success_status "$HTTP_STATUS"; then
    record_result "Use JWT" "OK" "Protected route accepted the JWT."
    print_result "OK" "Protected route accepted the JWT."
  elif [[ "$HTTP_STATUS" == "401" || "$HTTP_STATUS" == "403" ]]; then
    record_result "Use JWT" "FAIL" "Protected route rejected the JWT."
    print_result "FAIL" "Protected route rejected the JWT."
    abort_critical "The returned token was not accepted by protected API routes."
  else
    record_result "Use JWT" "WARN" "JWT reached the protected route, but the downstream movie search returned HTTP $HTTP_STATUS."
    print_result "WARN" "JWT reached the protected route, but movie search returned HTTP $HTTP_STATUS."
  fi
fi
pause_step

heading 7 "State Safety Checks"
explain "The CLI verifies that malformed callback requests are rejected."
MISSING_STATE_URL="$REAL_CALLBACK_URL"
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

MISMATCH_URL="$REAL_CALLBACK_URL?code=fake-code&state=wrong-state"
print_request "GET" "$MISMATCH_URL"
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
