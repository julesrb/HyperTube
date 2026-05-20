# Frontend and Backend Integration Challenges

This document outlines the main integration risks between the Next.js frontend,
the Go API, the torrent-stream service, and PostgreSQL. It is intentionally
planning-focused: the goal is to define the safest integration path before
changing application behavior.

## Current Integration State

The repository already has the major pieces in place, but they are not fully
connected yet.

- The frontend is a Next.js app on port `4200`.
- The API is a Go service under `/api/v1` on port `8080`.
- The streaming service is a separate Go service on port `8081`.
- PostgreSQL stores movies, torrents, users, OAuth accounts, comments, watch
  history, direct-stream movies, cached searches, and password-reset tokens.
- Frontend API service files exist but are currently empty.
- Frontend auth, movie data, user data, comments, and watch history are still
  mostly local/static.
- The API already exposes real auth, movie, torrent, comment, password-reset,
  OAuth, and OAuth2 password-grant routes.
- User profile API routes are documented in the project requirements but are
  not registered yet; `services/api/internal/users/handler.go` still contains
  empty handlers.
- The streaming service currently serves a test HLS pipeline and is not yet
  connected to authenticated movie playback or torrent selection.

## Recommended Integration Strategy

The safest approach is contract-first integration, not page-by-page wiring.
Before replacing local frontend state, define a small typed client layer and
normalize backend responses into frontend view models.

Recommended order:

1. Establish shared API conventions.
2. Add a frontend API client with auth headers, error handling, and response
   envelope parsing.
3. Integrate email/password auth and logout.
4. Add OAuth callback handling.
5. Replace static movie list/detail data with API data.
6. Replace static comments with API comments.
7. Implement or defer missing user profile API routes deliberately.
8. Connect watch history and watched/progression UI.
9. Connect streaming last, after auth and movie IDs are stable.
10. Add end-to-end verification scripts for the main user journeys.

This order keeps the security boundary stable before the streaming work starts.

## Main Challenges

### 1. Different Data Shapes

The frontend currently uses local types that do not match the API response
shape.

Frontend movie model:

```ts
type tMovie = {
  id: number;
  title: string;
  src: string;
  year: string;
  backdrops: string[];
  synopsis: string;
  genres: string[];
  directors: string[];
  stars: string[];
  length: string;
  rate: number;
};
```

API movie card response:

```json
{
  "imdb_id": "string",
  "title": "string",
  "year": "string",
  "poster_url": "string",
  "backdrop_url": "string",
  "note": 8.1,
  "genres": [878, 12, 18]
}
```

Key mismatches:

- Frontend routes use numeric local movie IDs; the backend uses IMDb IDs.
- Frontend uses local image filenames; the backend returns poster/backdrop URLs.
- Frontend genres are localized strings; the backend returns TMDB genre IDs.
- Frontend uses `rate`; backend uses `note`.
- Frontend uses `synopsis`, `directors`, `stars`, and `length`; backend detail
  responses use `summary`, `director`, `cast`, and `runtime_minutes`.
- Frontend comments store author display fields directly; backend comments
  currently return `user_id`, `movie_id`, `content`, and `updated_at`.

Recommendation:

- Do not force backend DTOs directly into React components.
- Add frontend API DTO types that mirror the backend exactly.
- Add mapper functions from API DTOs to frontend view models.
- Migrate frontend routes from `/movies/{number}` to `/movies/{imdb_id}` or
  introduce a clear route translation layer.

### 2. Auth Boundary and Token Storage

The backend already uses JWT bearer tokens for protected routes. The frontend
currently stores `token` and `user` in `localStorage`, but login/register still
use mock data.

Integration requirements:

- `POST /api/v1/auth/login` expects email and password.
- The current frontend sign-in modal asks for username and password.
- `POST /api/v1/auth/register` expects `first_name` and `last_name`, while the
  frontend local type uses `firstname` and `lastname`.
- Protected API routes require `Authorization: Bearer <access_token>`.
- Access tokens expire after 15 minutes.

Recommendation:

- Decide whether sign-in should use email or whether the backend should also
  support username login.
- Store token metadata, not only the raw token, so expiration can be handled.
- Keep route guards as UX helpers only; backend middleware remains the real
  security boundary.
- Add a single `getAccessToken()` helper so token handling is not duplicated
  across components.

Open decision:

- Keep `localStorage` for now because it matches the existing frontend, or move
  to HTTP-only cookies before streaming/auth hardening. Cookies improve browser
  playback integration but require CSRF considerations.

### 3. API Response Envelopes

The API consistently returns envelopes:

```json
{ "data": "...", "meta": { "total": 1, "page": 0, "per_page": 1 } }
```

Errors are shaped as:

```json
{ "error": { "code": "VALIDATION_ERROR", "message": "..." } }
```

Recommendation:

- The frontend API client should parse envelopes centrally.
- Components should not manually inspect `data`, `meta`, and `error`.
- Validation, unauthorized, not found, and server errors should map to explicit
  UI states.

### 4. Protected vs Public Movie Routes

Current API route behavior:

- `GET /api/v1/movies` is public and returns featured/front-page movies.
- `GET /api/v1/movies/search` is protected.
- `GET /api/v1/movies/{id}` is protected.
- `GET /api/v1/movies/{id}/torrents` is protected.
- `GET /api/v1/movies/{id}/comments` is protected.
- `POST /api/v1/movies/{id}/comments` is protected.

The frontend currently lets users browse local movie details without backend
auth. The subject requires the library and video sections to be authenticated,
but the front page can be public.

Recommendation:

- Keep the homepage on public `GET /movies`.
- Require auth before calling search, details, torrents, comments, and stream
  setup.
- Redirect or open the sign-in modal before entering protected movie details.

### 5. Missing User Profile API

The subject requires users to:

- update email address, profile picture, and profile information;
- view profiles of other users;
- keep email private when viewing other users.

Current gap:

- `/users`, `/users/{id}`, and `PATCH /users/{id}` are not registered.
- The frontend profile pages rely on local `AuthContext` user data.
- Profile pictures and user colors exist only in the frontend data model.

Recommendation:

- Decide whether profile integration is part of the first integration pass or a
  follow-up.
- If included, implement backend user DTOs before wiring the frontend.
- Separate "current user private profile" from "public user profile" responses
  so email privacy is explicit.

### 6. Comments Need Author Display Data

Backend comments currently return:

```json
{
  "id": 1,
  "user_id": 2,
  "movie_id": "tt1234567",
  "content": "string",
  "updated_at": "2026-05-06T12:00:00Z"
}
```

Frontend comments expect author names, profile picture, color, edited status,
and Unix timestamps.

Recommendation:

- Either extend comment responses with embedded author display fields, or fetch
  users separately and join in the frontend.
- Prefer embedding minimal author display fields in comment responses. It keeps
  comment rendering simple and avoids N+1 frontend calls.
- Add `created_at` separately if the UI must distinguish creation time from
  edit time.
- Add backend validation for empty or oversized comment content before wiring
  the UI.

### 7. Search, Sorting, Filtering, and Pagination

The frontend currently performs local search, sort, filter, and pagination over
static arrays. The backend has search pagination but does not expose every
frontend sort/filter control as query parameters yet.

Current backend search:

```text
GET /api/v1/movies/search?title=&page=
```

Recommendation:

- First wire search by title and backend pagination.
- Keep frontend-only sorting/filtering only for the loaded page if needed, but
  clearly mark it as temporary.
- Later move sort/filter parameters to the backend if the UX depends on global
  result ordering.
- Replace click-based pagination with infinite loading only after backend
  pagination is reliable.

### 8. Genre Localization

The frontend displays translated genre names. The backend returns TMDB genre
IDs.

Recommendation:

- Add a frontend genre map from TMDB IDs to translation keys.
- Avoid storing localized genre labels in backend data.
- Keep API responses language-neutral for genre IDs, and localize labels in the
  frontend.

### 9. Movie Detail Localization

`GET /api/v1/movies/{id}` supports a `lang` query parameter for TMDB details.
The frontend uses locales `en`, `fr`, and `de`.

Recommendation:

- Map frontend locales to TMDB language tags:
  - `en` -> `en-US`
  - `fr` -> `fr-FR`
  - `de` -> `de-DE`
- Pass the mapped language when loading movie details.
- Define fallback behavior if TMDB does not return localized text.

### 10. Streaming Is a Separate Security Problem

The stream service exposes:

- `GET /stream/{id}`
- `GET /stream/{id}/index`
- `GET /stream/{id}/{segment}`

These routes are separate from the API and currently do not share the API auth
middleware. HLS playback also makes auth harder because the player fetches the
playlist and segments after the initial play request.

Recommendation:

- Do not wire the play button directly to port `8081` as the final design.
- Prefer an authenticated API proxy for the first secure version:
  - browser calls `/api/v1/stream/...`;
  - API validates JWT;
  - API forwards internally to `torrent-stream`;
  - `torrent-stream` is no longer publicly exposed in production.
- Alternatively, use signed stream tickets, but that is more complex and should
  be considered only after basic playback works.

Streaming should be integrated after auth, movie IDs, and torrent selection are
stable.

### 11. CORS, Docker, and Environment Configuration

Current local ports:

- frontend: `4200`
- API: `8080`
- torrent-stream: `8081`
- Postgres: `5432`

The frontend service is commented out in `docker-compose.yml`, so local
frontend development likely runs with `npm run dev`.

Recommendation:

- Add `NEXT_PUBLIC_API_URL`, for example
  `http://localhost:8080/api/v1`.
- Add a separate internal stream URL for the API if an API proxy is used.
- Define allowed frontend origins in the API instead of using broad CORS.
- Keep torrent-stream internal-only in production-like setups.
- Document one canonical local startup flow.

### 12. Password Reset and OAuth Frontend Flows

The backend has password reset and OAuth routes. The frontend currently shows
mock password-reset success and does not yet have an OAuth callback page.

Recommendation:

- Wire password reset request first.
- Add a reset-password page matching `PASSWORD_RESET_URL`.
- Add an auth callback page that reads OAuth success or error parameters and
  stores the returned auth state.
- Confirm `FRONTEND_AUTH_CALLBACK_URL` matches the Next.js localized routing
  strategy.

### 13. Watch History and Progression

The frontend has `watch_history` with `movie_id` and `watch_percent`. The
backend has a `watch_history` table keyed by user ID and IMDb ID, and movie
detail responses currently expose `watched` and `progression` fields.

Recommendation:

- Define whether progression is percent, seconds, or a fraction.
- Store enough data to resume playback accurately.
- Add explicit API routes or payloads for updating progress.
- Use IMDb IDs consistently in watch history.

### 14. Tests and Verification

Do not rely on manual browser testing alone. Integration should be verified at
the contract level and at the user-journey level.

Recommended checks:

- API health check.
- Register, login, and logout flow.
- Invalid login and expired/missing token handling.
- Public homepage movie load.
- Protected movie search with and without token.
- Movie detail load with locale mapping.
- Comment list, create, update, and delete.
- Profile read/update once user routes exist.
- Stream init, playlist, and segment access with and without auth.
- Browser smoke test in Chrome and Firefox.
- Mobile layout smoke test after dynamic API data replaces fixed mock data.

## Suggested Integration Milestones

### Milestone 1: Contract and Client Foundation

- Create frontend API DTO types.
- Create a shared frontend API client.
- Parse `data`, `meta`, and `error` envelopes centrally.
- Add token injection for protected requests.
- Add typed mappers from backend DTOs to frontend view models.

Exit criteria:

- One frontend component can call the API through the client.
- Unauthorized responses are handled consistently.
- No component directly builds raw API URLs except the API client.

### Milestone 2: Real Auth

- Wire login to `POST /api/v1/auth/login`.
- Wire registration to `POST /api/v1/auth/register`.
- Align frontend field names with backend names.
- Decide email-login vs username-login.
- Add logout behavior that clears local auth state.

Exit criteria:

- A real backend user can register, log in, refresh the page, and remain
  recognized until token expiration.
- Protected API calls include a bearer token.

### Milestone 3: Movies

- Replace homepage static featured movies with `GET /api/v1/movies`.
- Replace search with `GET /api/v1/movies/search`.
- Replace movie detail page with `GET /api/v1/movies/{imdb_id}`.
- Map backend image URLs, genres, runtime, rating, summary, director, and cast.

Exit criteria:

- Frontend movie cards and detail pages use backend data.
- Routes work with IMDb IDs.
- Empty, loading, error, and unauthorized states are visible.

### Milestone 4: Comments

- Replace static comments with `GET /api/v1/movies/{id}/comments`.
- Wire comment creation to `POST /api/v1/movies/{id}/comments`.
- Wire edit/delete through `/comments/{id}`.
- Decide how author display data is supplied.

Exit criteria:

- Comments persist across reloads.
- Users can only edit/delete their own comments.
- Empty comments and invalid comment IDs fail cleanly.

### Milestone 5: Profiles and Watch History

- Implement and register missing user profile routes if this milestone is in
  scope.
- Replace local profile updates with backend updates.
- Replace local watch history with backend data.
- Define and implement progression updates.

Exit criteria:

- Current user profile changes persist.
- Public profiles do not expose private email addresses.
- Watched/progress indicators reflect backend data.

### Milestone 6: Streaming

- Decide API proxy vs signed stream-ticket strategy.
- Prevent direct public access to torrent-stream in production-like setups.
- Connect movie play action to authenticated stream setup.
- Serve playlist and segments through the chosen auth strategy.
- Handle token expiry during long playback.

Exit criteria:

- Logged-out users cannot access stream init, playlist, or segments.
- Logged-in users can start playback.
- HLS content types are correct.
- Playback works in current Firefox and Chrome.

## Highest-Risk Decisions To Make Early

1. Should movie URLs use IMDb IDs instead of numeric frontend IDs?
2. Should login use email, username, or both?
3. Should auth remain in `localStorage` or move to HTTP-only cookies?
4. Should streaming be protected by an API proxy or signed stream tickets?
5. Should comments include embedded author display data?
6. Should profile routes be implemented before movie/comment integration?
7. What exact format should watch progression use?

## Practical Next Step

The best first implementation step is to add a small frontend API client and
DTO layer without changing every page at once. After that, integrate auth
against the real backend. Once auth is stable, replace the movie and comment
mock data incrementally, page by page.

Streaming should remain last because it depends on stable auth, stable movie
identifiers, and a deliberate security design for playlist and segment access.
