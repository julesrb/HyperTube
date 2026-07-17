# api

Main HTTP gateway. Handles authentication, movie search, comments, and stream coordination.
Delegates HLS transcoding to the `torrent-stream` service.

**Base path:** `/api/v1`
**Default port:** `8080` (override with `PORT`)

---

## Public endpoints

### GET /health

Health check.

**Response:** `200 OK` — no body.

---

## Auth endpoints

All authentication endpoints are public. The paths below are relative to the
base path `/api/v1`.

The API uses the same success and error envelopes for most auth endpoints:

```json
{
  "data": {}
}
```

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "human readable message"
  }
}
```

Validation errors use field-based messages instead of a top-level `message`:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "fields": {
      "email": {
        "message": "Invalid email"
      }
    }
  }
}
```

Successful registration responses use this base auth payload. Password login
and successful browser OAuth callbacks add a `refresh_token`, while
registration does not:

```json
{
  "data": {
    "access_token": "<jwt>",
    "token_type": "Bearer",
    "expires_in": 900,
    "user": {
      "id": 1,
      "email": "ada@example.com",
      "username": "ada_lovelace",
      "first_name": "Ada",
      "last_name": "Lovelace"
    }
  }
}
```

`access_token` is a JWT signed with `HS256`. It contains `user_id` and
`token_use: "access"` claims and expires after 15 minutes. Protected routes
expect:

```http
Authorization: Bearer <access_token>
```

### POST /auth/register

Registers a new password user and returns a bearer token.

#### Request body

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `email` | string | yes | Valid email address. Trimmed and lowercased before storage. |
| `username` | string | yes | 3-32 characters. Letters, digits, and underscores only. |
| `first_name` | string | yes | 1-100 characters after trimming. |
| `last_name` | string | yes | 1-100 characters after trimming. |
| `password` | string | yes | 8-72 bytes. Stored as a bcrypt hash. |

Unknown JSON fields, malformed JSON, multiple JSON documents, and request bodies
larger than 1 MiB are rejected.

#### Example request

```http
POST /api/v1/auth/register
Content-Type: application/json
```

```json
{
  "email": "Ada@example.com",
  "username": "ada_lovelace",
  "first_name": "Ada",
  "last_name": "Lovelace",
  "password": "correct-horse-battery"
}
```

#### Response

`201 Created`

```json
{
  "data": {
    "access_token": "<jwt>",
    "token_type": "Bearer",
    "expires_in": 900,
    "user": {
      "id": 1,
      "email": "ada@example.com",
      "username": "ada_lovelace",
      "first_name": "Ada",
      "last_name": "Lovelace"
    }
  }
}
```

#### Error responses

| Status | Code | Message |
|--------|------|---------|
| 400 | `BAD_REQUEST` | `Invalid JSON body` |
| 400 | `VALIDATION_ERROR` | `Email is required` or `Invalid email` |
| 400 | `VALIDATION_ERROR` | `Username is required`, `Username is too short`, `Username is too long`, or `Username has invalid characters` |
| 400 | `VALIDATION_ERROR` | `First name is required` or `First name is too long` |
| 400 | `VALIDATION_ERROR` | `Last name is required` or `Last name is too long` |
| 400 | `VALIDATION_ERROR` | `Password is too short` or `Password is too long` |
| 409 | `ALREADY_EXIST_ERROR` | Field errors for `email`, `username`, or both |
| 500 | `INTERNAL_ERROR` | `Failed to create user` or `Failed to create token` |

Example:

```json
{
  "error": {
    "code": "ALREADY_EXIST_ERROR",
    "fields": {
      "email": {
        "message": "Email is already in use"
      },
      "username": {
        "message": "Username is already in use"
      }
    }
  }
}
```

### POST /auth/login

Logs in an existing password user by email or username and returns an access
token plus a refresh token.

#### Request body

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `login` | string | yes | Accepts an email address or username. Trimmed before lookup. Email addresses are lowercased. |
| `password` | string | yes | Existing password. Must be present and no longer than 72 bytes. |

Unknown JSON fields, malformed JSON, multiple JSON documents, and request bodies
larger than 1 MiB are rejected.

#### Example request

```http
POST /api/v1/auth/login
Content-Type: application/json
```

```json
{
  "login": "ada@example.com",
  "password": "correct-horse-battery"
}
```

#### Response

`200 OK`

```json
{
  "data": {
    "access_token": "<jwt>",
    "refresh_token": "<refresh-jwt>",
    "token_type": "Bearer",
    "expires_in": 900,
    "user": {
      "id": 1,
      "email": "ada@example.com",
      "username": "ada_lovelace",
      "first_name": "Ada",
      "last_name": "Lovelace"
    }
  }
}
```

The refresh token is issued by this password-login endpoint and by successful
browser OAuth callbacks. It is valid for 7 days and contains
`token_use: "refresh"`. `expires_in` continues to describe only the 15-minute
access token. Login responses include `Cache-Control: no-store` and
`Pragma: no-cache`.

Registration and `POST /oauth/token` do not return a HyperTube refresh token.

#### Error responses

| Status | Code | Message |
|--------|------|---------|
| 400 | `BAD_REQUEST` | `Invalid JSON body` |
| 400 | `VALIDATION_ERROR` | `Email or username is required`, `Invalid email`, `Invalid email or username`, `Username is too short`, or `Username is too long` |
| 400 | `VALIDATION_ERROR` | `Password is required` or `Password is too long` |
| 401 | `INVALID_CREDENTIALS` | `Invalid email, username, or password` |
| 500 | `INTERNAL_ERROR` | `Failed to load user` or `Failed to create token` |

Example:

```json
{
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "Invalid email, username, or password"
  }
}
```

### POST /auth/refresh-token

Exchanges a valid HyperTube refresh token for a new access token. The endpoint
is public because the previous access token may already have expired.

#### Request body

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `refresh_token` | string | yes | Refresh JWT returned by `POST /auth/login` or a successful browser OAuth callback. |

Unknown JSON fields, malformed JSON, multiple JSON documents, and request bodies
larger than 1 MiB are rejected.

```http
POST /api/v1/auth/refresh-token
Content-Type: application/json
```

```json
{
  "refresh_token": "<refresh-jwt>"
}
```

#### Response

`200 OK`

```json
{
  "data": {
    "access_token": "<new-access-jwt>",
    "token_type": "Bearer",
    "expires_in": 900
  }
}
```

The response does not include a user or a new refresh token. The same refresh
token may be reused until it expires; this minimal version has no rotation,
server-side revocation, session management, or logout invalidation. All success
and error responses include `Cache-Control: no-store` and `Pragma: no-cache`.

#### Error responses

| Status | Code | Message |
|--------|------|---------|
| 400 | `BAD_REQUEST` | `Invalid JSON body` |
| 400 | `VALIDATION_ERROR` | Field error: `Refresh token is required` |
| 401 | `INVALID_REFRESH_TOKEN` | `Refresh token is invalid or expired` |
| 500 | `INTERNAL_ERROR` | `Authentication service is unavailable` or `Failed to create access token` |

### POST /auth/password-reset/send-email

Creates a single-use password-reset token and sends a reset email when email is
configured. Unknown emails intentionally receive the same accepted response as
known emails.

Password reset email sending is active only when `BREVO_API_KEY` is present and
`MAIL_FROM_EMAIL` is a valid sender address. `MAIL_FROM_NAME` is optional and
defaults to `Hypertube`.

#### Request body

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `email` | string | yes | Valid email address. Trimmed and lowercased before lookup. |
| `locale` | string | no | Optional locale override for the reset email and `PASSWORD_RESET_URL`. Defaults to `Accept-Language`, then `en`. |

Unknown JSON fields, malformed JSON, multiple JSON documents, and request bodies
larger than 1 MiB are rejected.

#### Example request

```http
POST /api/v1/auth/password-reset/send-email
Content-Type: application/json
Accept-Language: de
```

```json
{
  "email": "ada@example.com"
}
```

#### Response

`202 Accepted`

```json
{
  "data": {
    "message": "If the email exists, a password reset link has been sent"
  }
}
```

#### Error responses

| Status | Code | Message |
|--------|------|---------|
| 400 | `BAD_REQUEST` | `Invalid JSON body` |
| 400 | `VALIDATION_ERROR` | `Email is required` or `Invalid email` |
| 503 | `EMAIL_NOT_CONFIGURED` | `Password reset email is not configured` |
| 500 | `INTERNAL_ERROR` | `Authentication service is unavailable` |
| 500 | `INTERNAL_ERROR` | `Failed to load user` |
| 500 | `INTERNAL_ERROR` | `Failed to create password reset token` |
| 500 | `INTERNAL_ERROR` | `Failed to store password reset token` |
| 500 | `INTERNAL_ERROR` | `Password reset URL is not configured` |
| 500 | `INTERNAL_ERROR` | `Failed to send password reset email` |

Example:

```json
{
  "error": {
    "code": "EMAIL_NOT_CONFIGURED",
    "message": "Password reset email is not configured"
  }
}
```

### POST /auth/password-reset/set-new-password

Consumes a password-reset token and replaces the user's password.

#### Request body

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `token` | string | yes | Token from the reset email. Must be 32-256 characters after trimming. |
| `password` | string | yes | New password. Must be 8-72 bytes. |

Unknown JSON fields, malformed JSON, multiple JSON documents, and request bodies
larger than 1 MiB are rejected.

#### Example request

```http
POST /api/v1/auth/password-reset/set-new-password
Content-Type: application/json
```

```json
{
  "token": "valid-reset-token-from-email-1234567890",
  "password": "new-correct-horse"
}
```

#### Response

`200 OK`

```json
{
  "data": {
    "message": "Password has been reset"
  }
}
```

#### Error responses

| Status | Code | Message |
|--------|------|---------|
| 400 | `BAD_REQUEST` | `Invalid JSON body` |
| 400 | `INVALID_RESET_TOKEN` | `Password reset link is invalid or expired` |
| 400 | `VALIDATION_ERROR` | `Password is too short` or `Password is too long` |
| 400 | `VALIDATION_ERROR` | `Invalid password` |
| 500 | `INTERNAL_ERROR` | `Authentication service is unavailable` |
| 500 | `INTERNAL_ERROR` | `Failed to reset password` |

Example:

```json
{
  "error": {
    "code": "INVALID_RESET_TOKEN",
    "message": "Password reset link is invalid or expired"
  }
}
```

### GET /auth/42/login → GET /auth/42/callback

Starts and completes the 42 OAuth authorization-code flow.

#### GET /auth/42/login

No request body is used.

The endpoint generates a random state value, stores it in an HttpOnly cookie
named `hypertube_oauth_42_state`, and redirects the browser to 42.

##### Example request

```http
GET /api/v1/auth/42/login
```

##### Response

`302 Found`

```http
Location: https://api.intra.42.fr/oauth/authorize?client_id=<client_id>&redirect_uri=<redirect_uri>&response_type=code&scope=public&state=<state>
Set-Cookie: hypertube_oauth_42_state=<state>; Path=/; Max-Age=600; HttpOnly; SameSite=Lax
```

##### Error responses

| Status | Code | Message |
|--------|------|---------|
| 503 | `OAUTH_NOT_CONFIGURED` | `OAuth provider 42 is not configured` |
| 500 | `INTERNAL_ERROR` | `Failed to create OAuth state` |
| 500 | `INTERNAL_ERROR` | `Failed to start 42 OAuth` |

Example:

```json
{
  "error": {
    "code": "OAUTH_NOT_CONFIGURED",
    "message": "OAuth provider 42 is not configured"
  }
}
```

#### GET /auth/42/callback

No request body is used.

Required callback inputs:

| Parameter | Location | Required | Description |
|-----------|----------|----------|-------------|
| `code` | query | yes, unless provider returned `error` | Authorization code from 42. |
| `state` | query | yes | Must match the `hypertube_oauth_42_state` cookie. |
| `hypertube_oauth_42_state` | cookie | yes | State cookie set by `/auth/42/login`. |
| `error` | query | no | Provider denial or provider-side error. |

##### Example request

```http
GET /api/v1/auth/42/callback?code=provider-code&state=<state>
Cookie: hypertube_oauth_42_state=<state>
```

##### Response

When `FRONTEND_AUTH_CALLBACK_URL` is configured, which defaults to
`http://localhost:4200/en/auth/callback`, the API redirects to the frontend with
auth data in the URL fragment:

`303 See Other`

```http
Location: http://localhost:4200/en/auth/callback#access_token=<jwt>&refresh_token=<refresh-jwt>&token_type=Bearer&expires_in=900&user=%7B...%7D
Cache-Control: no-store
Pragma: no-cache
Set-Cookie: hypertube_oauth_42_state=; Path=/; Max-Age=0; HttpOnly; SameSite=Lax
```

The `user` fragment value is URL-encoded JSON:

```json
{
  "id": 1,
  "email": "ft.user@example.com",
  "username": "ft_user",
  "first_name": "Forty",
  "last_name": "Two"
}
```

If the handler is configured without a frontend callback URL, the success
response is `200 OK` with the standard auth payload plus `refresh_token`.
Successful browser OAuth callback responses include `Cache-Control: no-store`
and `Pragma: no-cache`. In both response forms, `expires_in` describes only the
access token.

##### Error responses

With `FRONTEND_AUTH_CALLBACK_URL`, callback errors redirect to the frontend:

```http
HTTP/1.1 303 See Other
Location: http://localhost:4200/en/auth/callback?error=INVALID_OAUTH_STATE&error_description=Invalid+OAuth+state
```

Without a frontend callback URL, errors use the standard JSON error envelope.

| Status | Code | Message |
|--------|------|---------|
| 400 | `INVALID_OAUTH_STATE` | `Invalid OAuth state` |
| 401 | `OAUTH_DENIED` | Provider error value, for example `access_denied`. |
| 502 | `OAUTH_EXCHANGE_FAILED` | `Failed to exchange 42 authorization code` |
| 503 | `OAUTH_NOT_CONFIGURED` | `OAuth provider 42 is not configured` |
| 500 | `INTERNAL_ERROR` | `Failed to create OAuth user`, `Failed to create token`, or invalid frontend callback configuration |

### GET /auth/github/login → GET /auth/github/callback

Starts and completes the GitHub OAuth authorization-code flow.

#### GET /auth/github/login

No request body is used.

The endpoint generates a random state value, stores it in an HttpOnly cookie
named `hypertube_oauth_github_state`, and redirects the browser to GitHub.

##### Example request

```http
GET /api/v1/auth/github/login
```

##### Response

`302 Found`

```http
Location: https://github.com/login/oauth/authorize?client_id=<client_id>&redirect_uri=<redirect_uri>&response_type=code&scope=read%3Auser+user%3Aemail&state=<state>
Set-Cookie: hypertube_oauth_github_state=<state>; Path=/; Max-Age=600; HttpOnly; SameSite=Lax
```

##### Error responses

| Status | Code | Message |
|--------|------|---------|
| 503 | `OAUTH_NOT_CONFIGURED` | `OAuth provider GitHub is not configured` |
| 500 | `INTERNAL_ERROR` | `Failed to create OAuth state` |
| 500 | `INTERNAL_ERROR` | `Failed to start GitHub OAuth` |

Example:

```json
{
  "error": {
    "code": "OAUTH_NOT_CONFIGURED",
    "message": "OAuth provider GitHub is not configured"
  }
}
```

#### GET /auth/github/callback

No request body is used.

Required callback inputs:

| Parameter | Location | Required | Description |
|-----------|----------|----------|-------------|
| `code` | query | yes, unless provider returned `error` | Authorization code from GitHub. |
| `state` | query | yes | Must match the `hypertube_oauth_github_state` cookie. |
| `hypertube_oauth_github_state` | cookie | yes | State cookie set by `/auth/github/login`. |
| `error` | query | no | Provider denial or provider-side error. |

##### Example request

```http
GET /api/v1/auth/github/callback?code=provider-code&state=<state>
Cookie: hypertube_oauth_github_state=<state>
```

##### Response

When `FRONTEND_AUTH_CALLBACK_URL` is configured, which defaults to
`http://localhost:4200/en/auth/callback`, the API redirects to the frontend with
auth data in the URL fragment:

`303 See Other`

```http
Location: http://localhost:4200/en/auth/callback#access_token=<jwt>&refresh_token=<refresh-jwt>&token_type=Bearer&expires_in=900&user=%7B...%7D
Cache-Control: no-store
Pragma: no-cache
Set-Cookie: hypertube_oauth_github_state=; Path=/; Max-Age=0; HttpOnly; SameSite=Lax
```

The `user` fragment value is URL-encoded JSON:

```json
{
  "id": 1,
  "email": "gh.user@example.com",
  "username": "gh_user",
  "first_name": "Git",
  "last_name": "Hub"
}
```

If the handler is configured without a frontend callback URL, the success
response is `200 OK` with the standard auth payload plus `refresh_token`.
`expires_in` continues to describe only the access token.

##### Error responses

With `FRONTEND_AUTH_CALLBACK_URL`, callback errors redirect to the frontend:

```http
HTTP/1.1 303 See Other
Location: http://localhost:4200/en/auth/callback?error=OAUTH_EXCHANGE_FAILED&error_description=failed+to+exchange+GitHub+authorization+code
```

Without a frontend callback URL, errors use the standard JSON error envelope.

| Status | Code | Message |
|--------|------|---------|
| 400 | `INVALID_OAUTH_STATE` | `Invalid OAuth state` |
| 401 | `OAUTH_DENIED` | Provider error value, for example `access_denied`. |
| 502 | `OAUTH_EXCHANGE_FAILED` | `Failed to exchange GitHub authorization code` |
| 503 | `OAUTH_NOT_CONFIGURED` | `OAuth provider GitHub is not configured` |
| 500 | `INTERNAL_ERROR` | `Failed to create OAuth user`, `Failed to create token`, or invalid frontend callback configuration |

### GET /auth/gitlab/login -> GET /auth/gitlab/callback

Starts and completes the GitLab OAuth authorization-code flow.

#### GET /auth/gitlab/login

No request body is used.

The endpoint generates a random state value, stores it in an HttpOnly cookie
named `hypertube_oauth_gitlab_state`, and redirects the browser to GitLab.

##### Example request

```http
GET /api/v1/auth/gitlab/login
```

##### Response

`302 Found`

```http
Location: https://gitlab.com/oauth/authorize?client_id=<client_id>&redirect_uri=<redirect_uri>&response_type=code&scope=read_user&state=<state>
Set-Cookie: hypertube_oauth_gitlab_state=<state>; Path=/; Max-Age=600; HttpOnly; SameSite=Lax
```

##### Error responses

| Status | Code | Message |
|--------|------|---------|
| 503 | `OAUTH_NOT_CONFIGURED` | `OAuth provider GitLab is not configured` |
| 500 | `INTERNAL_ERROR` | `Failed to create OAuth state` |
| 500 | `INTERNAL_ERROR` | `Failed to start GitLab OAuth` |

#### GET /auth/gitlab/callback

No request body is used.

Required callback inputs:

| Parameter | Location | Required | Description |
|-----------|----------|----------|-------------|
| `code` | query | yes, unless provider returned `error` | Authorization code from GitLab. |
| `state` | query | yes | Must match the `hypertube_oauth_gitlab_state` cookie. |
| `hypertube_oauth_gitlab_state` | cookie | yes | State cookie set by `/auth/gitlab/login`. |
| `error` | query | no | Provider denial or provider-side error. |

##### Example request

```http
GET /api/v1/auth/gitlab/callback?code=provider-code&state=<state>
Cookie: hypertube_oauth_gitlab_state=<state>
```

##### Response

When `FRONTEND_AUTH_CALLBACK_URL` is configured, which defaults to
`http://localhost:4200/en/auth/callback`, the API redirects to the frontend with
`access_token`, `refresh_token`, `token_type`, `expires_in`, and `user` in the
URL fragment, matching the 42 and GitHub callback shape. Without a frontend
callback URL, the success response is `200 OK` with the standard auth payload
plus `refresh_token`. `expires_in` continues to describe only the access token.

##### Error responses

| Status | Code | Message |
|--------|------|---------|
| 400 | `INVALID_OAUTH_STATE` | `Invalid OAuth state` |
| 401 | `OAUTH_DENIED` | Provider error value, for example `access_denied`. |
| 502 | `OAUTH_EXCHANGE_FAILED` | `Failed to exchange GitLab authorization code` |
| 503 | `OAUTH_NOT_CONFIGURED` | `OAuth provider GitLab is not configured` |
| 500 | `INTERNAL_ERROR` | `Failed to create OAuth user`, `Failed to create token`, or invalid frontend callback configuration |

### OAuth applications

OAuth applications are managed by authenticated users. Each application belongs
to its owner user. Successful `client_credentials` token requests issue the same
JWT access-token type used by normal login, with the owner user ID as the token
subject.

These routes require:

```http
Authorization: Bearer <access_token>
```

#### Create application

```http
POST /api/v1/oauth/applications
Content-Type: application/json
Authorization: Bearer <access_token>
```

```json
{
  "name": "My App",
  "redirect_uri": "http://localhost:4200/auth/callback"
}
```

`name` and `redirect_uri` are required and trimmed.

`redirect_uri` must be an absolute `http` or `https` URL with a host. Query
strings are allowed; URL userinfo and fragments are rejected.

`201 Created`

```json
{
  "data": {
    "id": 1,
    "name": "My App",
    "redirect_uri": "http://localhost:4200/auth/callback",
    "client_id": "htc_...",
    "client_secret": "hts_...",
    "created_at": "2026-07-07T12:00:00Z",
    "updated_at": "2026-07-07T12:00:00Z"
  }
}
```

`client_secret` is returned only once, in the create response. The API stores
only a bcrypt hash of the secret and never returns `client_secret_hash`.

#### List applications

```http
GET /api/v1/oauth/applications?page=0
Authorization: Bearer <access_token>
```

`page` is zero-based. Missing, invalid, empty, or negative values are treated as
`0`. Results are paginated with 12 applications per page and sorted by
`created_at DESC, id DESC`.

`200 OK`

```json
{
  "data": [
    {
      "id": 1,
      "name": "My App",
      "redirect_uri": "http://localhost:4200/auth/callback",
      "client_id": "htc_...",
      "created_at": "2026-07-07T12:00:00Z",
      "updated_at": "2026-07-07T12:00:00Z"
    }
  ],
  "meta": {
    "total": 1,
    "page": 0,
    "per_page": 12
  }
}
```

An empty list returns `"data": []`.

#### Update application

```http
PATCH /api/v1/oauth/applications/1
Content-Type: application/json
Authorization: Bearer <access_token>
```

```json
{
  "name": "My Renamed App",
  "redirect_uri": "https://example.com/oauth/callback"
}
```

At least one of `name` or `redirect_uri` is required. `redirect_uri` uses the
same validation rules as create. Only applications owned by the authenticated
user can be updated.

`200 OK`

```json
{
  "data": {
    "id": 1,
    "name": "My Renamed App",
    "redirect_uri": "https://example.com/oauth/callback",
    "client_id": "htc_...",
    "created_at": "2026-07-07T12:00:00Z",
    "updated_at": "2026-07-07T12:05:00Z"
  }
}
```

#### Delete application

```http
DELETE /api/v1/oauth/applications/1
Authorization: Bearer <access_token>
```

`200 OK`

```json
{
  "data": null
}
```

Deleting an OAuth application prevents future token issuance for that
application. Already-issued stateless JWT access tokens can remain valid until
their normal expiry, currently 15 minutes.

#### Application error responses

| Status | Code | Description |
|--------|------|-------------|
| 400 | `BAD_REQUEST` | Malformed JSON, unknown fields, or multiple JSON documents. |
| 400 | `VALIDATION_ERROR` | Missing name or redirect URI, invalid redirect URI, overly long name or redirect URI, or empty patch body. |
| 401 | `UNAUTHORIZED` | Missing or invalid bearer token. |
| 404 | `NOT_FOUND` | Application ID is invalid, missing, or not owned by the authenticated user. |
| 500 | `INTERNAL_ERROR` | Application storage or credential generation failed. |

### POST /oauth/token

OAuth2-compatible token endpoint for registered application
`client_credentials` grants. It returns a JWT bearer access token for API routes.
User password login belongs to `POST /api/v1/auth/login`.

Unlike the other auth endpoints, this endpoint uses OAuth2 token response shapes
directly. Success responses are not wrapped in `data`, and errors are not
wrapped in `error`.

#### Request body

Supported content types:

- `application/x-www-form-urlencoded`
- `application/json`
- empty `Content-Type`, parsed as form data

#### Client credentials grant

Registered OAuth applications use `grant_type=client_credentials`. When the
application authenticates successfully, the API issues a normal access token for
the application owner user ID. This flow does not require or use
`redirect_uri` and does not support OAuth scopes.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `grant_type` | string | usually | Must be `client_credentials` when present. It may be omitted when `client_id` and `client_secret` are present and password-login fields are absent. |
| `client_id` | string | yes | Client ID returned by application creation. Trimmed before lookup. |
| `client_secret` | string | yes | One-time secret returned by application creation. It is not trimmed before verification. |

##### Example request

```http
POST /api/v1/oauth/token
Content-Type: application/x-www-form-urlencoded
```

```text
grant_type=client_credentials&client_id=htc_...&client_secret=hts_...
```

JSON is also accepted, including without `grant_type`:

```json
{
  "client_id": "htc_...",
  "client_secret": "hts_..."
}
```

The endpoint issues a normal owner-user bearer token. It does not accept,
return, or enforce OAuth scopes.

#### Response

`200 OK`

```json
{
  "access_token": "<jwt>",
  "token_type": "Bearer",
  "expires_in": 900
}
```

Response headers include:

```http
Cache-Control: no-store
Pragma: no-cache
```

#### Error responses

| Status | Error | Description |
|--------|-------|-------------|
| 400 | `invalid_request` | Invalid `Content-Type`, invalid body, missing `grant_type`, missing `client_id` or `client_secret`, or omitted `grant_type` with password-login fields present. |
| 400 | `unsupported_grant_type` | `grant_type` is not `client_credentials`. |
| 401 | `invalid_client` | `client_id` or `client_secret` is incorrect. |
| 415 | `invalid_request` | Body must be form encoded or JSON. |
| 500 | `server_error` | Auth service unavailable, token creation failed, or application lookup failed. |

Example:

```json
{
  "error": "unsupported_grant_type",
  "error_description": "Only client_credentials grant type is supported"
}
```

---

## User endpoints

All user endpoints require `Authorization: Bearer <access_token>`.

## GET /users/{id}

Returns the public profile of the requested user. The path parameter `id` must
be a positive integer. First and last names are reduced to initials; email,
password data, and `updated_at` are never returned.

### Response

```json
{
  "data": {
    "id": 7,
    "username": "alice",
    "first_name": "A",
    "last_name": "L",
    "profile_picture": null,
    "color": "green",
    "created_at": "2026-05-06T12:00:00Z"
  }
}
```

A syntactically valid ID for an unknown user returns `404 NOT_FOUND`.

### Error responses

| Status | Code | Description |
|--------|------|-------------|
| 401 | `UNAUTHORIZED` | Bearer token is missing or invalid. |
| 401 | `TOKEN_EXPIRED` | Bearer token has expired. |
| 404 | `NOT_FOUND` | Path user ID is invalid or the user does not exist. |
| 500 | `INTERNAL_ERROR` | Loading the user failed. |

---

## GET /users/{id}/comments

Returns comments posted by the requested user. Any authenticated user may read
another user's comments.

### Path and query parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `id` | integer | yes | | Positive ID of the user whose comments should be displayed. |
| `page` | integer | no | `0` | Zero-based page index. Invalid or negative values use page `0`. |

Results contain 12 comments per page and are ordered by `updated_at DESC`, then
by `id DESC` when timestamps are equal.

### Response

```json
{
  "data": [
    {
      "id": 17,
      "user_id": 42,
      "movie_id": "tt1234567",
      "movie": {
        "imdb_id": "tt1234567",
        "title": "Example Movie",
        "year": "2025",
        "backdrop_url": "https://example.test/backdrop.jpg"
      },
      "content": "A very good movie.",
      "edited": false,
      "updated_at": "2026-06-20T12:00:00Z"
    }
  ],
  "meta": {
    "total": 1,
    "page": 0,
    "per_page": 12
  }
}
```

The response contains the comment fields, a small `movie` object, and pagination
metadata. `movie_id` remains present alongside `movie` for frontend
compatibility. An empty collection is returned as `"data": []`, never `null`.
A syntactically valid ID for an unknown user returns an empty collection.

### Error responses

| Status | Code | Description |
|--------|------|-------------|
| 401 | `UNAUTHORIZED` | Bearer token is missing or invalid. |
| 401 | `TOKEN_EXPIRED` | Bearer token has expired. |
| 404 | `NOT_FOUND` | Path user ID is not a positive integer. |
| 500 | `INTERNAL_ERROR` | Counting or loading the user's comments failed. |

---

## GET /users/{id}/movie-history

Returns all stored watch-history entries for the requested user, ordered by
`watched_at DESC`. Any authenticated user may read another user's history.
There is no query pagination; every existing entry is returned. History has at
most one row per user/movie.

### Path parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `id` | integer | yes | Positive ID of the user whose history should be displayed. |

### Response

```json
{
  "data": [
    {
      "imdb_id": "tt1234567",
      "title": "Example Movie",
      "year": "2025",
      "poster_url": "https://example.test/poster.jpg",
      "backdrop_url": "https://example.test/backdrop.jpg",
      "note": 8.1,
      "genres": [12, 18],
      "progress": 1804,
      "complete": false
    }
  ],
  "meta": {
    "total": 1,
    "page": 0,
    "per_page": 1
  }
}
```

A syntactically valid ID for an unknown user returns `"data": []` with zeroed
metadata.

### Error responses

| Status | Code | Description |
|--------|------|-------------|
| 401 | `UNAUTHORIZED` | Bearer token is missing or invalid. |
| 401 | `TOKEN_EXPIRED` | Bearer token has expired. |
| 404 | `NOT_FOUND` | Path user ID is not a positive integer. |
| 500 | `INTERNAL_ERROR` | Loading the user's film history failed. |

---

## PATCH /users/new-password

Changes the authenticated password user's password. This route requires a valid
access-token bearer header; the user ID is read only from that token.

### Request body

Only the following underscore-named JSON fields are accepted:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `current_password` | string | yes | Existing password, 1-72 bytes. |
| `new_password` | string | yes | New password, 8-72 bytes and not a common password. Must differ from the current password. |
| `new_password_confirm` | string | no | Frontend compatibility field. When present, it must exactly equal `new_password`. |

Passwords are compared exactly as sent and are never trimmed or normalized.
Unknown fields, hyphenated request aliases, malformed JSON, multiple JSON
documents, and bodies larger than 1 MiB are rejected.

Response validation fields deliberately use the names expected by the existing
form: `current-password`, `new-password`, and `confirm-new-password`.

```http
PATCH /api/v1/users/new-password
Authorization: Bearer <access_token>
Content-Type: application/json
Accept-Language: en
```

```json
{
  "current_password": "old-correct-horse",
  "new_password": "new-correct-horse",
  "new_password_confirm": "new-correct-horse"
}
```

### Response

```json
{
  "data": {
    "message": "Password has been changed"
  }
}
```

No access or refresh tokens are issued, rotated, or revoked by this endpoint.

### Status codes

| Status | Code | Description |
|--------|------|-------------|
| 200 | — | Password changed. |
| 400 | `BAD_REQUEST` | Invalid JSON structure, unknown field, or oversized body. |
| 400 | `VALIDATION_ERROR` | Invalid field, confirmation mismatch, common password, or OAuth user. |
| 401 | `UNAUTHORIZED` | Bearer token is missing or invalid. |
| 401 | `TOKEN_EXPIRED` | Bearer token has expired. |
| 401 | `INVALID_CURRENT_PASSWORD` | Current password is incorrect or changed concurrently. |
| 404 | `NOT_FOUND` | The authenticated user no longer exists. |
| 409 | `PASSWORD_UNCHANGED` | New password equals the current password. |
| 500 | `INTERNAL_ERROR` | Loading or updating the user failed. |

```json
{
  "error": {
    "code": "INVALID_CURRENT_PASSWORD",
    "fields": {
      "current-password": { "message": "Current password is invalid" }
    }
  }
}
```

```json
{
  "error": {
    "code": "PASSWORD_UNCHANGED",
    "fields": {
      "new-password": { "message": "New password must differ from current password" }
    }
  }
}
```

---

## PATCH /users/{id}

Updates the authenticated user's own profile. The `{id}` path value must match
the user ID in the bearer token. Send only the fields that should change; all
other profile values stay untouched.

### Request body

At least one field is required. Unknown JSON fields, malformed JSON, multiple
JSON documents, and request bodies larger than 1 MiB are rejected.

| Parameter | Type | Description |
|-----------|------|-------------|
| `email` | string | Valid email address. Password users only. |
| `username` | string | 3-32 characters. Letters, digits, and underscores only. Password users only. |
| `first_name` | string | 1-100 characters after trimming. Password users only. |
| `last_name` | string | 1-100 characters after trimming. Password users only. |
| `profile_picture` | string or null | Protected field. Send `null` or an empty string to remove it. Non-empty strings are rejected. |
| `color` | string | One of `yellow`, `pink`, `green`, `purple`, `blue`, or `red`. |

Password users can update the documented identity and appearance fields, but
`profile_picture` can only be removed. OAuth users can update only `color` and
remove `profile_picture` through this endpoint; `email`, `username`,
`first_name`, and `last_name` are managed by the OAuth provider and are rejected.
A user is considered an OAuth user when they have at least one linked
`oauth_accounts` row.

### Example request

```http
PATCH /api/v1/users/1
Authorization: Bearer <access_token>
Content-Type: application/json
```

```json
{
  "username": "ada_lovelace",
  "first_name": "Ada",
  "profile_picture": null,
  "color": "purple"
}
```

Use `profile_picture: null` to remove the stored profile picture.
`profile_picture: ""` also removes it, matching the existing frontend behavior.
Non-empty `profile_picture` strings are rejected.
For password users, include identity fields only when they should change. Use
`PATCH /users/new-password` for password changes.

### Response

When no profile picture is stored, responses include `"profile_picture": null`.

```json
{
  "data": {
    "id": 1,
    "email": "ada@example.com",
    "username": "ada_lovelace",
    "first_name": "Ada",
    "last_name": "Lovelace",
    "profile_picture": null,
    "color": "purple",
    "created_at": "2026-05-06T12:00:00Z",
    "updated_at": "2026-05-06T12:00:00Z"
  }
}
```

### Error responses

```json
{ "error": { "code": "FORBIDDEN", "message": "Cannot update another user's profile" } }
```
```json
{ "error": { "code": "VALIDATION_ERROR", "fields": { "email": { "message": "OAuth users cannot change their email" } } } }
```
```json
{ "error": { "code": "VALIDATION_ERROR", "fields": { "profile_picture": { "message": "Profile picture can only be removed" } } } }
```
```json
{ "error": { "code": "ALREADY_EXIST_ERROR", "fields": { "email": { "message": "Email is already in use" } } } }
```

---

## Stream endpoints *(temporarily public — will move behind auth)*

### GET /stream/{id}

Initialise the HLS stream for a movie. Calls `POST /transcode/{id}` on the `torrent-stream` service.
If `/data/videos/{id}/` already exists the call returns immediately without re-transcoding.

| Parameter | Description          |
|-----------|----------------------|
| `id`      | IMDb ID of the movie |

| Code | Meaning                               |
|------|---------------------------------------|
| 200  | Stream is ready or has been started   |
| 500  | Transcode service unreachable/failed  |

### GET /stream/{id}/index

Returns the HLS playlist (`stream.m3u8`).

**Response:** `200 OK` — `Content-Type: application/vnd.apple.mpegurl`

### GET /stream/{id}/{segment}

Serves a single `.ts` segment (e.g. `stream0.ts`). The player fetches these automatically from the playlist.

**Response:** `200 OK` — `Content-Type: video/mp2t`

---

## Movie and comment endpoints

`GET /movies` and `GET /movies/featured` are public. The remaining movie and comment endpoints in this
section require `Authorization: Bearer <jwt>`.

## GET /movies

Returns a list of tracker-wide popular movies.

### Response

```json
{
  "data": [
    {
      "imdb_id": "string",
      "title": "string",
      "year": "string",
      "poster_url": "string",
      "backdrop_url": "string",
      "note": 8.1,
      "genres": [878, 12, 18]
    }
  ],
  "meta": { "total": 12, "page": 0, "per_page": 12 }
}
```

> Only the card fields are returned. Full details are available via `GET /movies/{id}`.

### Error responses

```json
{ "error": { "code": "INTERNAL_ERROR", "message": "Failed to load movies" } }
```

---

## GET /movies/featured

Returns a curated selection of movies.

### Response

```json
{
  "data": [
    {
      "imdb_id": "string",
      "title": "string",
      "year": "string",
      "poster_url": "string",
      "backdrop_url": "string",
      "note": 8.1,
      "genres": [878, 12, 18]
    }
  ],
  "meta": { "total": 12, "page": 0, "per_page": 12 }
}
```

> Only the card fields are returned. Full details are available via `GET /movies/{id}`.

### Error responses

```json
{ "error": { "code": "INTERNAL_ERROR", "message": "Failed to load movies" } }
```

---

## GET /movies/directstream

Returns movies available for direct streaming.

### Response

```json
{
  "data": [
    {
      "imdb_id": "string",
      "title": "string",
      "year": "string",
      "poster_url": "string",
      "backdrop_url": "string",
      "note": 8.1,
      "genres": [878, 12, 18]
    }
  ],
  "meta": { "total": 2, "page": 0, "per_page": 2 }
}
```

### Error responses

```json
{ "error": { "code": "INTERNAL_ERROR", "message": "Failed to load movies" } }
```

---

## GET /movies/search?title=&page=

Searches for movies by title. On first request, fetches from tracker sources, resolves metadata via TMDB, and persists results to the database. Subsequent requests for the same title are served directly from the database. Results are paginated at 12 per page.

### Query parameters

| Parameter | Type    | Required | Default | Description                  |
|-----------|---------|----------|---------|------------------------------|
| `title`   | string  | yes      |         | Title to search for          |
| `page`    | integer | no       | `0`     | Page index (0 is first page) |

### Response

```json
{
  "data": [
    {
      "imdb_id": "string",
      "title": "string",
      "year": "string",
      "poster_url": "string",
      "backdrop_url": "string",
      "note": 8.1,
      "genres": [878, 12, 18]
    }
  ],
  "meta": { "total": 87, "page": 0, "per_page": 12 }
}
```

### Error responses

```json
{ "error": { "code": "VALIDATION_ERROR", "fields": { "title": { "message": "Title query parameter is required" } } } }
```
```json
{ "error": { "code": "INTERNAL_ERROR", "message": "Failed to search movies" } }
```

---

## GET /movies/watched

Returns the watch history for the authenticated user, ordered by most recently
watched. Requires `Authorization: Bearer <access_token>`.

No request body is used. The user is taken from the JWT.

### Response

```json
{
  "data": [
    {
      "imdb_id": "string",
      "title": "string",
      "year": "string",
      "poster_url": "string",
      "backdrop_url": "string",
      "note": 8.1,
      "genres": [878, 12, 18]
    }
  ],
  "meta": { "total": 3, "page": 0, "per_page": 3 }
}
```

### Error responses

```json
{ "error": { "code": "VALIDATION_ERROR", "fields": { "body": { "message": "Invalid request body" } } } }
```
```json
{ "error": { "code": "INTERNAL_ERROR", "message": "Failed to load movies" } }
```

---

## PATCH /movies/{imdbId}/progress

Saves playback progress for the authenticated user and movie. Requires
`Authorization: Bearer <access_token>`. The user is always taken from the JWT.

### Path parameters

| Parameter | Type   | Description          |
|-----------|--------|----------------------|
| `imdbId`  | string | IMDb ID of the movie |

### Request body

Only `progress`, `pourcent` and `complete` are accepted. `progress` is the playback position
in seconds and must be a non-negative integer. `complete` marks whether the
movie was fully watched. And `pourcent` is the percentage of the movie watched, an integer between 0 and 100.

```json
{
  "progress": 1804,
  "pourcent": 54,
  "complete": false
}
```

### Response

`200 OK`

```json
{
  "data": {
    "progress": 1804,
    "pourcent": 54,
    "complete": false
  }
}
```

### Error responses

| Status | Code | Description |
|--------|------|-------------|
| 400 | `BAD_REQUEST` | Malformed JSON, unknown fields, or multiple JSON documents. |
| 400 | `VALIDATION_ERROR` | Missing, null, wrong-type, or negative fields. |
| 401 | `UNAUTHORIZED` | Bearer token is missing or invalid. |
| 401 | `TOKEN_EXPIRED` | Bearer token has expired. |
| 404 | `NOT_FOUND` | Movie ID is empty or unknown. |
| 500 | `INTERNAL_ERROR` | Saving progress failed. |

---

## GET /movies/{id}

Returns full metadata for a single movie. Summary, director, and cast are fetched live from TMDB.

### Path parameters

| Parameter | Type   | Description          |
|-----------|--------|----------------------|
| `id`      | string | IMDb ID of the movie |

### Query parameters

| Parameter | Type   | Required | Default  | Description                        |
|-----------|--------|----------|----------|------------------------------------|
| `lang`    | string | no       | `en`     | App locale for the details (`en`, `fr`, or `de`) |

### Response

```json
{
  "data": {
    "imdb_id": "string",
    "tmdb_id": "string",
    "title": "string",
    "year": "string",
    "poster_url": "string",
    "backdrop_url": "string",
    "note": 8.1,
    "genres": [878, 12, 18],
    "summary": "string",
    "director": "string",
    "cast": ["string"],
    "extra_backdrops": ["string"],
  }
}
```

### Error responses

```json
{ "error": { "code": "NOT_FOUND", "message": "Movie not found" } }
```
```json
{ "error": { "code": "INTERNAL_ERROR", "message": "Failed to load movie" } }
```
```json
{ "error": { "code": "INTERNAL_ERROR", "message": "Failed to fetch movie details" } }
```

---

## GET /movies/{id}/torrents

Returns available torrent sources for a movie.

### Path parameters

| Parameter | Type   | Description          |
|-----------|--------|----------------------|
| `id`      | string | IMDb ID of the movie |

### Response

```json
{
  "data": [
    {
      "id": 1,
      "imdb_id": "string",
      "title": "string",
      "year": 2024,
      "source": "string",
      "url": "string",
      "quality": "string",
      "size": 2.4,
      "language": "string",
      "seeds": "string"
    }
  ],
  "meta": { "total": 4, "page": 0, "per_page": 4 }
}
```

### Error responses

```json
{ "error": { "code": "NOT_FOUND", "message": "No tracker source found for this movie" } }
```
```json
{ "error": { "code": "INTERNAL_ERROR", "message": "Failed to load tracker source" } }
```

---

## GET /movies/{id}/comments

Returns comments posted on a movie, ordered by most recent first.

### Path parameters

| Parameter | Type   | Description          |
|-----------|--------|----------------------|
| `id`      | string | IMDb ID of the movie |

### Query parameters

| Parameter | Type    | Required | Default | Description                    |
|-----------|---------|----------|---------|--------------------------------|
| `page`    | integer | no       | `0`     | Page index (0 is first page)   |

### Response

```json
{
  "data": [
    {
      "id": 1,
      "user_id": 2,
      "movie_id": "string",
      "content": "string",
      "edited": false,
      "updated_at": "2026-05-06T12:00:00Z"
    }
  ],
  "meta": { "total": 8, "page": 0, "per_page": 12 }
}
```

### Error responses

```json
{ "error": { "code": "NOT_FOUND", "message": "Movie not found" } }
```
```json
{ "error": { "code": "INTERNAL_ERROR", "message": "Failed to access comments" } }
```

---

## POST /movies/{id}/comments

Posts a new comment on a movie as the authenticated user. Requires
`Authorization: Bearer <access_token>`.

### Path parameters

| Parameter | Type   | Description          |
|-----------|--------|----------------------|
| `id`      | string | IMDb ID of the movie |

### Request body

```json
{
  "content": "string"
}
```

### Response

`201 Created`

```json
{
  "data": {
    "id": 1,
    "user_id": 1,
    "movie_id": "string",
    "content": "string",
    "edited": false,
    "updated_at": "2026-05-06T12:00:00Z"
  }
}
```

### Error responses

```json
{ "error": { "code": "VALIDATION_ERROR", "fields": { "content": { "message": "Invalid request body" } } } }
```
```json
{ "error": { "code": "NOT_FOUND", "message": "Movie not found" } }
```
```json
{ "error": { "code": "INTERNAL_ERROR", "message": "Failed to create comment" } }
```

---

## GET /comments

Returns all comments across all movies.

### Query parameters

| Parameter | Type    | Required | Default | Description                    |
|-----------|---------|----------|---------|--------------------------------|
| `page`    | integer | no       | `0`     | Page index (0 is first page)   |

### Response

```json
{
  "data": [
    {
      "id": 1,
      "user_id": 2,
      "movie_id": "string",
      "content": "string",
      "edited": false,
      "updated_at": "2026-05-06T12:00:00Z"
    }
  ],
  "meta": { "total": 8, "page": 0, "per_page": 12 }
}
```

### Error responses

```json
{ "error": { "code": "INTERNAL_ERROR", "message": "Failed to load comments" } }
```

---

## GET /comments/{id}

Returns a single comment by its ID.

### Path parameters

| Parameter | Type   | Description       |
|-----------|--------|-------------------|
| `id`      | string | ID of the comment |

### Response

```json
{
  "data": {
    "id": 1,
    "user_id": 2,
    "movie_id": "string",
    "content": "string",
    "edited": false,
    "updated_at": "2026-05-06T12:00:00Z"
  }
}
```

### Error responses

```json
{ "error": { "code": "NOT_FOUND", "message": "Comment not found" } }
```
```json
{ "error": { "code": "INTERNAL_ERROR", "message": "Failed to load comment" } }
```

---

## PATCH /comments/{id}

Updates the content of an existing comment if it belongs to the authenticated
user. Requires `Authorization: Bearer <access_token>`.

### Path parameters

| Parameter | Type   | Description       |
|-----------|--------|-------------------|
| `id`      | string | ID of the comment |

### Request body

```json
{
  "content": "string"
}
```

### Response

```json
{
  "data": {
    "id": 1,
    "user_id": 2,
    "movie_id": "string",
    "content": "string",
    "edited": true,
    "updated_at": "2026-05-06T12:00:00Z"
  }
}
```

### Error responses

```json
{ "error": { "code": "VALIDATION_ERROR", "fields": { "content": { "message": "Invalid request body" } } } }
```
```json
{ "error": { "code": "NOT_FOUND", "message": "Comment not found" } }
```
```json
{ "error": { "code": "INTERNAL_ERROR", "message": "Failed to update comment" } }
```

---

## DELETE /comments/{id}

Deletes a comment if it belongs to the authenticated user. Requires
`Authorization: Bearer <access_token>`.

### Path parameters

| Parameter | Type   | Description       |
|-----------|--------|-------------------|
| `id`      | string | ID of the comment |

No request body is used. The user is taken from the JWT.

### Response

`200 OK`

```json
{ "data": null }
```

### Error responses

```json
{ "error": { "code": "NOT_FOUND", "message": "Comment not found" } }
```
```json
{ "error": { "code": "INTERNAL_ERROR", "message": "Failed to delete comment" } }
```
