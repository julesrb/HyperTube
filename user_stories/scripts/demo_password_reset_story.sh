#!/usr/bin/env bash
#
# CLI demo walkthrough for the Hypertube password reset user story.
#
# User story:
#   As a user who forgot my password, I want to request a reset link by email
#   and set a new password with the reset token, so I can regain access without
#   exposing whether an email address exists in Hypertube.
#
# Usage:
#   ./user_stories/scripts/demo_password_reset_story.sh
#
# Optional environment variables:
#   BASE_URL       API origin. Default: http://localhost:8080
#   DEMO_EMAIL     Demo user's email. Default: unique email per run.
#   DEMO_USERNAME  Demo user's username. Default: unique username per run.
#   OLD_PASSWORD   Demo user's original password. Default: OldPass123!
#   NEW_PASSWORD   Demo user's new password. Default: NewPass123!
#   RESET_LOCALE   Locale sent to the reset request. Default: en.
#   RESET_TOKEN    Token from the emailed reset link. If omitted, the token
#                 consumption steps are reported as skipped.
#
# Examples:
#   BASE_URL=http://localhost:8080 ./user_stories/scripts/demo_password_reset_story.sh
#   RESET_TOKEN=token-from-email ./user_stories/scripts/demo_password_reset_story.sh

set -o pipefail

TOTAL_STEPS=8
RUN_ID="$(date +%Y%m%d%H%M%S)-$$"

BASE_URL="${BASE_URL:-http://localhost:8080}"
BASE_URL="${BASE_URL%/}"
API_URL="$BASE_URL/api/v1"
DEMO_EMAIL="${DEMO_EMAIL:-password-reset-${RUN_ID}@example.test}"
DEMO_USERNAME="${DEMO_USERNAME:-reset_${RUN_ID//[^A-Za-z0-9_]/_}}"
OLD_PASSWORD="${OLD_PASSWORD:-OldPass123!}"
NEW_PASSWORD="${NEW_PASSWORD:-NewPass123!}"
RESET_LOCALE="${RESET_LOCALE:-en}"
RESET_TOKEN="${RESET_TOKEN:-}"

TOKEN=""
USER_ID=""
MAILER_AVAILABLE="unknown"

TMP_FILES=()
SUMMARY_NAMES=()
SUMMARY_STATUS=()
SUMMARY_DETAILS=()

REQUEST_STATUS=""
REQUEST_BODY=""
REQUEST_ERROR=""

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
  sed -n '2,25p' "$0" | sed 's/^# \{0,1\}//'
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

  heading 8 "Summary"
  explain "The CLI reports which parts of the password reset walkthrough succeeded, failed, or were skipped."

  printf '\n%sResult:%s\n' "$BOLD" "$RESET"
  for i in "${!SUMMARY_NAMES[@]}"; do
    printf '  %-5s %s - %s\n' "$(status_label "${SUMMARY_STATUS[$i]}")" "${SUMMARY_NAMES[$i]}" "${SUMMARY_DETAILS[$i]}"
  done

  printf '\n%sConfiguration:%s\n' "$BOLD" "$RESET"
  printf '  BASE_URL: %s\n' "$BASE_URL"
  printf '  DEMO_EMAIL: %s\n' "$DEMO_EMAIL"
  printf '  DEMO_USERNAME: %s\n' "$DEMO_USERNAME"
  printf '  RESET_LOCALE: %s\n' "$RESET_LOCALE"
  if [[ -n "$RESET_TOKEN" ]]; then
    printf '  RESET_TOKEN: <provided>\n'
  else
    printf '  RESET_TOKEN: <not provided>\n'
  fi
  printf '  MAILER_AVAILABLE: %s\n' "$MAILER_AVAILABLE"
}

has_failures() {
  local status
  for status in "${SUMMARY_STATUS[@]}"; do
    if [[ "$status" == "FAIL" ]]; then
      return 0
    fi
  done
  return 1
}

finish() {
  print_summary
  if has_failures; then
    exit 1
  fi
}

abort_critical() {
  local message="$1"
  printf '\n%sCritical failure:%s %s\n' "$RED" "$RESET" "$message" >&2
  record_result "Critical failure" "FAIL" "$message"
  finish
}

is_success_status() {
  [[ "$1" =~ ^2[0-9][0-9]$ ]]
}

error_code() {
  jq -r '.error.code // empty' <<<"$REQUEST_BODY" 2>/dev/null
}

data_token() {
  jq -r '.data.access_token // empty' <<<"$REQUEST_BODY" 2>/dev/null
}

data_user_id() {
  jq -r '.data.user.id // empty' <<<"$REQUEST_BODY" 2>/dev/null
}

redact_payload() {
  local payload="$1"

  if [[ -z "$payload" ]]; then
    return 0
  fi

  if jq -e . >/dev/null 2>&1 <<<"$payload"; then
    jq 'if type == "object" then
          (if has("password") then .password = "<redacted>" else . end)
          | (if has("token") then .token = "<redacted>" else . end)
        else . end' <<<"$payload"
  else
    printf '%s\n' "$payload"
  fi
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

  printf '\n%sRequest:%s\n' "$BOLD" "$RESET"
  printf '  %s %s\n' "$method" "$url"

  if [[ -n "$payload" ]]; then
    printf '\n%sPayload:%s\n' "$BOLD" "$RESET"
    redact_payload "$payload" | sed 's/^/  /'
  fi
}

print_response() {
  printf '\n%sResponse:%s\n' "$BOLD" "$RESET"
  printf '  HTTP %s\n' "${REQUEST_STATUS:-<none>}"

  if [[ -n "$REQUEST_ERROR" ]]; then
    printf '  curl error: %s\n' "$REQUEST_ERROR"
  fi

  pretty_json_or_raw "$REQUEST_BODY"
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
  local body_file
  local error_file
  local curl_args

  body_file="$(mktemp "${TMPDIR:-/tmp}/hypertube-password-reset-body.XXXXXX")"
  error_file="$(mktemp "${TMPDIR:-/tmp}/hypertube-password-reset-curl.XXXXXX")"
  TMP_FILES+=("$body_file" "$error_file")

  curl_args=(
    --silent
    --show-error
    --max-time 45
    -X "$method"
    "$url"
    -H "Accept: application/json"
    -o "$body_file"
    -w "%{http_code}"
  )

  if [[ -n "$payload" ]]; then
    curl_args+=(-H "Content-Type: application/json" --data "$payload")
  fi

  REQUEST_STATUS="$(curl "${curl_args[@]}" 2>"$error_file")"
  REQUEST_BODY="$(<"$body_file")"
  REQUEST_ERROR="$(<"$error_file")"
}

register_payload() {
  jq -n \
    --arg email "$DEMO_EMAIL" \
    --arg username "$DEMO_USERNAME" \
    --arg first_name "Password" \
    --arg last_name "Reset" \
    --arg password "$OLD_PASSWORD" \
    '{email:$email, username:$username, first_name:$first_name, last_name:$last_name, password:$password}'
}

login_payload() {
  local password="$1"

  jq -n --arg email "$DEMO_EMAIL" --arg password "$password" '{email:$email, password:$password}'
}

password_reset_payload() {
  local email="$1"

  jq -n --arg email "$email" --arg locale "$RESET_LOCALE" '{email:$email, locale:$locale}'
}

reset_password_payload() {
  local password="$1"

  jq -n --arg token "$RESET_TOKEN" --arg password "$password" '{token:$token, password:$password}'
}

step_register_user() {
  local payload

  heading 1 "Create a demo account"
  explain "A user exists before the reset flow starts, with a known old password."

  payload="$(register_payload)"
  print_request "POST" "$API_URL/auth/register" "$payload"
  request "POST" "$API_URL/auth/register" "$payload"
  print_response

  if [[ "$REQUEST_STATUS" == "201" ]]; then
    TOKEN="$(data_token)"
    USER_ID="$(data_user_id)"
    record_result "Create demo account" "OK" "registered user id ${USER_ID:-<unknown>}"
    print_result "OK" "Registered the demo user."
    return 0
  fi

  if [[ "$REQUEST_STATUS" == "409" ]]; then
    record_result "Create demo account" "WARN" "account already exists; continuing with login"
    print_result "WARN" "The account already exists, so the walkthrough continues by logging in."
    return 0
  fi

  abort_critical "expected register to return 201 or 409, got ${REQUEST_STATUS:-<none>}"
}

step_login_old_password() {
  local payload

  heading 2 "Login with the old password"
  explain "The old password works before the reset token is consumed."

  payload="$(login_payload "$OLD_PASSWORD")"
  print_request "POST" "$API_URL/auth/login" "$payload"
  request "POST" "$API_URL/auth/login" "$payload"
  print_response

  if [[ "$REQUEST_STATUS" == "200" ]]; then
    TOKEN="$(data_token)"
    USER_ID="$(data_user_id)"
    record_result "Old password login" "OK" "old password accepted before reset"
    print_result "OK" "The old password currently works."
    return 0
  fi

  abort_critical "old password login failed with HTTP ${REQUEST_STATUS:-<none>}"
}

step_validate_reset_email() {
  local payload

  heading 3 "Validate reset email input"
  explain "Invalid reset requests are rejected before any email or token work happens."

  payload="$(password_reset_payload "not-an-email")"
  print_request "POST" "$API_URL/auth/password-reset" "$payload"
  request "POST" "$API_URL/auth/password-reset" "$payload"
  print_response

  if [[ "$REQUEST_STATUS" == "400" && "$(error_code)" == "VALIDATION_ERROR" ]]; then
    record_result "Invalid reset email" "OK" "invalid email returned VALIDATION_ERROR"
    print_result "OK" "The API rejected the invalid email."
    return 0
  fi

  record_result "Invalid reset email" "FAIL" "expected VALIDATION_ERROR, got HTTP ${REQUEST_STATUS:-<none>} $(error_code)"
  print_result "FAIL" "The invalid email contract did not match."
}

step_request_reset_link() {
  local payload

  heading 4 "Request a reset link"
  explain "A valid email request should return the same accepted response shape used to avoid account enumeration."

  payload="$(password_reset_payload "$DEMO_EMAIL")"
  print_request "POST" "$API_URL/auth/password-reset" "$payload"
  request "POST" "$API_URL/auth/password-reset" "$payload"
  print_response

  if [[ "$REQUEST_STATUS" == "202" ]]; then
    MAILER_AVAILABLE="yes"
    record_result "Request reset link" "OK" "reset request accepted and email sending is configured"
    print_result "OK" "The reset request was accepted."
    return 0
  fi

  if [[ "$REQUEST_STATUS" == "503" && "$(error_code)" == "EMAIL_NOT_CONFIGURED" ]]; then
    MAILER_AVAILABLE="no"
    record_result "Request reset link" "SKIP" "email provider is not configured in this environment"
    print_result "SKIP" "The API is reachable, but email sending is not configured."
    return 0
  fi

  record_result "Request reset link" "FAIL" "expected 202 or EMAIL_NOT_CONFIGURED, got HTTP ${REQUEST_STATUS:-<none>} $(error_code)"
  print_result "FAIL" "The reset request contract did not match."
}

step_unknown_email_same_response() {
  local payload
  local missing_email

  heading 5 "Request reset for an unknown email"
  explain "When mail is configured, unknown emails should still receive the accepted response and no account detail leaks."

  if [[ "$MAILER_AVAILABLE" != "yes" ]]; then
    record_result "Unknown email privacy" "SKIP" "email provider is not configured"
    print_result "SKIP" "Skipped because the request endpoint cannot reach mail sending in this environment."
    return 0
  fi

  missing_email="missing-password-reset-${RUN_ID}@example.test"
  payload="$(password_reset_payload "$missing_email")"
  print_request "POST" "$API_URL/auth/password-reset" "$payload"
  request "POST" "$API_URL/auth/password-reset" "$payload"
  print_response

  if [[ "$REQUEST_STATUS" == "202" ]]; then
    record_result "Unknown email privacy" "OK" "unknown email received the same accepted response"
    print_result "OK" "Unknown and known email requests use the same public response."
    return 0
  fi

  record_result "Unknown email privacy" "FAIL" "expected 202, got HTTP ${REQUEST_STATUS:-<none>}"
  print_result "FAIL" "The unknown email response leaked a different status."
}

step_validate_new_password() {
  local payload

  heading 6 "Validate the new password"
  explain "A weak new password is rejected before a valid reset token would be consumed."

  if [[ -z "$RESET_TOKEN" ]]; then
    record_result "New password validation" "SKIP" "RESET_TOKEN was not provided"
    print_result "SKIP" "Provide RESET_TOKEN from the email link to run this step."
    return 0
  fi

  payload="$(reset_password_payload "short")"
  print_request "POST" "$API_URL/auth/reset-password" "$payload"
  request "POST" "$API_URL/auth/reset-password" "$payload"
  print_response

  if [[ "$REQUEST_STATUS" == "400" && "$(error_code)" == "VALIDATION_ERROR" ]]; then
    record_result "New password validation" "OK" "short password returned VALIDATION_ERROR"
    print_result "OK" "The API rejected the weak replacement password."
    return 0
  fi

  record_result "New password validation" "FAIL" "expected VALIDATION_ERROR, got HTTP ${REQUEST_STATUS:-<none>} $(error_code)"
  print_result "FAIL" "The new password validation contract did not match."
}

step_consume_reset_token() {
  local payload

  heading 7 "Set the new password"
  explain "The reset token is exchanged for the new password and should become single-use."

  if [[ -z "$RESET_TOKEN" ]]; then
    record_result "Consume reset token" "SKIP" "RESET_TOKEN was not provided"
    print_result "SKIP" "The reset-token exchange requires the token from the email link."
    return 0
  fi

  payload="$(reset_password_payload "$NEW_PASSWORD")"
  print_request "POST" "$API_URL/auth/reset-password" "$payload"
  request "POST" "$API_URL/auth/reset-password" "$payload"
  print_response

  if [[ "$REQUEST_STATUS" == "200" ]]; then
    record_result "Consume reset token" "OK" "new password accepted"
    print_result "OK" "The password was reset."
    return 0
  fi

  record_result "Consume reset token" "FAIL" "expected 200, got HTTP ${REQUEST_STATUS:-<none>} $(error_code)"
  print_result "FAIL" "The reset token could not be consumed."
}

step_verify_login_and_replay() {
  local payload

  heading 8 "Verify login and replay behavior"
  explain "After a successful reset, the old password should fail, the new password should work, and the token should not be reusable."

  if [[ -z "$RESET_TOKEN" ]]; then
    record_result "Login after reset" "SKIP" "RESET_TOKEN was not provided"
    record_result "Token replay" "SKIP" "RESET_TOKEN was not provided"
    print_result "SKIP" "Skipped because no reset token was provided."
    return 0
  fi

  payload="$(login_payload "$OLD_PASSWORD")"
  print_request "POST" "$API_URL/auth/login" "$payload"
  request "POST" "$API_URL/auth/login" "$payload"
  print_response

  if [[ "$REQUEST_STATUS" == "401" ]]; then
    record_result "Old password after reset" "OK" "old password rejected"
    print_result "OK" "The old password no longer works."
  else
    record_result "Old password after reset" "FAIL" "expected 401, got HTTP ${REQUEST_STATUS:-<none>}"
    print_result "FAIL" "The old password was not rejected."
  fi

  payload="$(login_payload "$NEW_PASSWORD")"
  print_request "POST" "$API_URL/auth/login" "$payload"
  request "POST" "$API_URL/auth/login" "$payload"
  print_response

  if [[ "$REQUEST_STATUS" == "200" ]]; then
    record_result "New password login" "OK" "new password accepted"
    print_result "OK" "The new password works."
  else
    record_result "New password login" "FAIL" "expected 200, got HTTP ${REQUEST_STATUS:-<none>}"
    print_result "FAIL" "The new password login failed."
  fi

  payload="$(reset_password_payload "${NEW_PASSWORD}-again")"
  print_request "POST" "$API_URL/auth/reset-password" "$payload"
  request "POST" "$API_URL/auth/reset-password" "$payload"
  print_response

  if [[ "$REQUEST_STATUS" == "400" && "$(error_code)" == "INVALID_RESET_TOKEN" ]]; then
    record_result "Token replay" "OK" "used token rejected"
    print_result "OK" "The used token cannot be replayed."
  else
    record_result "Token replay" "FAIL" "expected INVALID_RESET_TOKEN, got HTTP ${REQUEST_STATUS:-<none>} $(error_code)"
    print_result "FAIL" "The token replay contract did not match."
  fi
}

main() {
  step_register_user
  step_login_old_password
  step_validate_reset_email
  step_request_reset_link
  step_unknown_email_same_response
  step_validate_new_password
  step_consume_reset_token
  step_verify_login_and_replay
  finish
}

main "$@"
