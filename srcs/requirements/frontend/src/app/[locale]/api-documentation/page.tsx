import React from "react";

type Row = {name: string; type?: string; required?: string; defaultValue?: string; description: string};
type ErrorRow = {status: string; code: string; message: string};

const json = (value: string) => (
    <pre className="overflow-x-auto border p-4 text-xs sm:text-sm leading-6">
        <code className="whitespace-pre">{value}</code>
    </pre>
);

function Table({headers, rows}: {headers: string[]; rows: string[][]}) {
    return (<div className="overflow-x-auto border bg-white">
        <table className="min-w-full text-left text-sm">
            <thead className="bg-black text-white">
            <tr>
                {headers.map((header) => (
                    <th key={header} className="px-4 py-3 font-wide uppercase tracking-wide">
                        {header}
                    </th>
                ))}
            </tr>
            </thead>
            <tbody>
            {rows.map((row, rowIndex) => (
                <tr key={rowIndex} className="border-t border-gray-loading align-top">
                    {row.map((cell, cellIndex) => (
                        <td key={cellIndex} className="px-4 py-3 text-sm">
                            {cell}
                        </td>
                    ))}
                </tr>
            ))}
            </tbody>
        </table>
    </div>);
}

function Endpoint({method, path, children, badge}: {method: string; path: string; children: React.ReactNode; badge?: string}) {
    return (<section className="border bg-white p-6 custom-shadow-animation-l">
        <div className="flex flex-wrap items-start justify-between gap-3 border-b border-gray-loading pb-4">
            <div>
                <div className="flex flex-wrap items-center gap-2">
                    <span className="border bg-black px-3 py-1 text-xs font-wide uppercase tracking-widest text-white">
                        {method}
                    </span>
                    {badge && (
                        <span className="border px-3 py-1 text-xs uppercase tracking-widest text-black">
                            {badge}
                        </span>
                    )}
                </div>
                <h3 className="mt-3 break-all text-2xl sm:text-3xl">{path}</h3>
            </div>
        </div>
        <div className="mt-5 space-y-5 text-sm sm:text-base">{children}</div>
    </section>);
}

function FieldTable({rows}: {rows: Row[]}) {
    return Table({
        headers: ["Parameter", "Type", "Required", "Default", "Description"],
        rows: rows.map((row) => [row.name, row.type ?? "—", row.required ?? "no", row.defaultValue ?? "—", row.description]),
    });
}

function ErrorTable({rows}: {rows: ErrorRow[]}) {
    return Table({
        headers: ["Status", "Code", "Message"],
        rows: rows.map((row) => [row.status, row.code, row.message]),
    });
}

function Note({children, color="yellow"}: {children: React.ReactNode, color?: string}) {
    return (<p className={`border bg-${color} px-4 py-3 text-sm`}>{children}</p>);
}

export default function ApiDocumentationPage() {
    return (<main className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 lg:px-8 lg:py-12">
        <div className="p-6 sm:p-8 lg:p-10">
            <div className="flex flex-wrap items-start justify-between gap-6">
                <div className="max-w-3xl">
                    <p className="text-xs uppercase tracking-[0.3em] text-gray">Hypertube API</p>
                    <h1 className="mt-2">Documentation</h1>
                    <p className="mt-4 max-w-3xl text-base  sm:text-lg">
                        Main HTTP gateway for auth, movies, comments, users, and streaming.
                        The layout follows the site’s stark black-and-white style with bold borders and condensed headings.
                    </p>
                </div>

                <div className="grid gap-3 sm:grid-cols-2">
                    <div className="border px-4 py-3">
                        <p className="text-xs uppercase tracking-widest text-gray">Base path</p>
                        <p className="mt-1 font-wide text-lg">/api/v1</p>
                    </div>
                    <div className="border px-4 py-3">
                        <p className="text-xs uppercase tracking-widest text-gray">Default port</p>
                        <p className="mt-1 font-wide text-lg">8080</p>
                    </div>
                </div>
            </div>

            <div className="mt-8 grid gap-4 md:grid-cols-3">
                <a href="#auth-endpoints" className="block border bg-black px-5 py-4 text-white custom-shadow-animation-l" aria-label="Go to auth endpoints">
                    <p className="text-xs uppercase tracking-widest text-white">Auth</p>
                    <p className="mt-2 text-sm text-white/90">Registration, password login, refresh tokens, password reset, and OAuth.</p>
                </a>
                <a href="#content-endpoints" className="block border px-5 py-4 custom-shadow-animation-l" aria-label="Go to content endpoints">
                    <p className="text-xs uppercase tracking-widest text-gray">Content</p>
                    <p className="mt-2 text-sm">Movies, comments, watch history, and progress tracking.</p>
                </a>
                <a href="#streaming-endpoints" className="block border px-5 py-4 custom-shadow-animation-l" aria-label="Go to streaming endpoints">
                    <p className="text-xs uppercase tracking-widest text-gray">Streaming</p>
                    <p className="mt-2 text-sm">HLS orchestration through the torrent-stream service.</p>
                </a>
            </div>

            <div className="mt-10 space-y-8">
                <section className="space-y-7">
                    <h2>Public endpoints</h2>
                    <Endpoint method="GET" path="/health">
                        <p>Health check.</p>
                        <Note color="green">Response: <strong>200 OK</strong> — no body.</Note>
                    </Endpoint>
                </section>

                <section id="auth-endpoints" className="scroll-mt-6 space-y-7">
                    <h2>Auth endpoints</h2>
                    <p>
                        All authentication routes are public. Most use the standard <code>data</code> / <code>error</code> envelope.
                    </p>

                    <Endpoint method="POST" path="/auth/register">
                        <p>Registers a password user and returns an access token.</p>
                        <FieldTable rows={[
                            {name: "email", type: "string", required: "yes", description: "Valid email address, trimmed and lowercased."},
                            {name: "username", type: "string", required: "yes", description: "3-32 characters, letters, digits, and underscores only."},
                            {name: "first_name", type: "string", required: "yes", description: "1-100 characters after trimming."},
                            {name: "last_name", type: "string", required: "yes", description: "1-100 characters after trimming."},
                            {name: "password", type: "string", required: "yes", description: "8-72 bytes, stored as bcrypt."},
                        ]} />
                        {json(`{
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
}`)}
                        <ErrorTable rows={[
                            {status: "400", code: "BAD_REQUEST", message: "Invalid JSON body"},
                            {status: "409", code: "ALREADY_EXIST_ERROR", message: "Field errors for email, username, or both"},
                            {status: "500", code: "INTERNAL_ERROR", message: "Failed to create user or token"},
                        ]} />
                    </Endpoint>

                    <Endpoint method="POST" path="/auth/login">
                        <p>Logs in by email or username and returns access plus refresh tokens.</p>
                        <FieldTable rows={[
                            {name: "login", type: "string", required: "yes", description: "Email address or username, trimmed before lookup."},
                            {name: "password", type: "string", required: "yes", description: "Existing password, max 72 bytes."},
                        ]} />
                        <Note>Login responses include <strong>Cache-Control: no-store</strong> and <strong>Pragma: no-cache</strong>.</Note>
                        {json(`{
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
}`)}
                    </Endpoint>

                    <Endpoint method="POST" path="/auth/refresh-token">
                        <p>Exchanges a valid HyperTube refresh token for a new access token.</p>
                        <FieldTable rows={[
                            {name: "refresh_token", type: "string", required: "yes", description: "Refresh JWT returned by login or browser OAuth callback."},
                        ]} />
                        {json(`{
    "data": {
        "access_token": "<new-access-jwt>",
        "token_type": "Bearer",
        "expires_in": 900
    }
}`)}
                    </Endpoint>

                    <Endpoint method="POST" path="/auth/password-reset/send-email">
                        <p>Creates a single-use password-reset token and sends the email when configured.</p>
                        <FieldTable rows={[
                            {name: "email", type: "string", required: "yes", description: "Valid email address, trimmed and lowercased."},
                            {name: "locale", type: "string", required: "no", description: "Optional locale override for the reset email."},
                        ]} />
                        {json(`{
    "data": {
        "message": "If the email exists, a password reset link has been sent"
    }
}`)}
                    </Endpoint>

                    <Endpoint method="POST" path="/auth/password-reset/set-new-password">
                        <p>Consumes a password-reset token and replaces the user password.</p>
                        <FieldTable rows={[
                            {name: "token", type: "string", required: "yes", description: "Reset token from email, 32-256 characters after trimming."},
                            {name: "password", type: "string", required: "yes", description: "New password, 8-72 bytes."},
                        ]} />
                        {json(`{
    "data": {
        "message": "Password has been reset"
    }
}`)}
                    </Endpoint>

                    <Endpoint method="GET" path="/auth/42/login → /auth/42/callback" badge="OAuth">
                        <p>Starts and completes the 42 OAuth authorization-code flow.</p>
                        <p>The same pattern is used for GitHub and GitLab, with provider-specific state cookies and callback redirects.</p>
                        <Note>Success callbacks can redirect to the frontend with auth data in the URL fragment, or return JSON if no frontend callback is configured.</Note>
                    </Endpoint>

                    <Endpoint method="GET" path="/auth/github/login → /auth/github/callback" badge="OAuth">
                        <p>GitHub OAuth flow, mirroring the 42 and GitLab patterns.</p>
                    </Endpoint>

                    <Endpoint method="GET" path="/auth/gitlab/login → /auth/gitlab/callback" badge="OAuth">
                        <p>GitLab OAuth flow, mirroring the 42 and GitHub patterns.</p>
                    </Endpoint>

                    <Endpoint method="POST" path="/oauth/applications">
                        <p>Creates a new OAuth application for the authenticated user.</p>
                        <FieldTable rows={[
                            {name: "name", type: "string", required: "yes", description: "Trimmed application name."},
                            {name: "redirect_uri", type: "string", required: "yes", description: "Absolute http(s) URL with host."},
                        ]} />
                        <Note><strong>client_secret</strong> is returned only once, in the create response.</Note>
                    </Endpoint>

                    <Endpoint method="GET" path="/oauth/applications?page=0">
                        <p>Lists the authenticated user’s OAuth applications, 12 per page.</p>
                    </Endpoint>

                    <Endpoint method="PATCH" path="/oauth/applications/{id}">
                        <p>Updates name or redirect URI for an application owned by the authenticated user.</p>
                    </Endpoint>

                    <Endpoint method="DELETE" path="/oauth/applications/{id}">
                        <p>Deletes an OAuth application and blocks future token issuance for it.</p>
                    </Endpoint>

                    <Endpoint method="POST" path="/oauth/token">
                        <p>OAuth2-compatible token endpoint for client_credentials grants.</p>
                        <Note>Success responses are not wrapped in <code>data</code>; errors are not wrapped in <code>error</code>.</Note>
                    </Endpoint>
                </section>

                <section className="space-y-7">
                    <h2>User endpoints</h2>
                    <Endpoint method="GET" path="/users/{id}">
                        <p>Returns the public profile of the requested user.</p>
                        <FieldTable rows={[
                            {name: "id", type: "integer", required: "yes", description: "Positive user ID in the path."},
                        ]} />
                        {json(`{
    "data": {
        "id": 7,
        "username": "alice",
        "first_name": "A",
        "last_name": "L",
        "profile_picture": null,
        "color": "green",
        "created_at": "2026-05-06T12:00:00Z"
    }
}`)}
                        <ErrorTable rows={[
                            {status: "401", code: "UNAUTHORIZED", message: "Bearer token is missing or invalid"},
                            {status: "401", code: "TOKEN_EXPIRED", message: "Bearer token has expired"},
                            {status: "404", code: "NOT_FOUND", message: "Path user ID is invalid or the user does not exist"},
                            {status: "500", code: "INTERNAL_ERROR", message: "Loading the user failed"},
                        ]} />
                    </Endpoint>

                    <Endpoint method="GET" path="/users/{id}/comments">
                        <p>Returns comments posted by the requested user.</p>
                        <FieldTable rows={[
                            {name: "id", type: "integer", required: "yes", description: "Positive user ID in the path."},
                            {name: "page", type: "integer", defaultValue: "0", description: "Zero-based page index; invalid or negative values use page 0."},
                        ]} />
                        {json(`{
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
}`)}
                        <ErrorTable rows={[
                            {status: "401", code: "UNAUTHORIZED", message: "Bearer token is missing or invalid"},
                            {status: "401", code: "TOKEN_EXPIRED", message: "Bearer token has expired"},
                            {status: "404", code: "NOT_FOUND", message: "Path user ID is not a positive integer"},
                            {status: "500", code: "INTERNAL_ERROR", message: "Counting or loading the user's comments failed"},
                        ]} />
                    </Endpoint>

                    <Endpoint method="GET" path="/users/{id}/movie-history">
                        <p>Returns stored watch-history entries for the requested user.</p>
                        <FieldTable rows={[
                            {name: "id", type: "integer", required: "yes", description: "Positive user ID in the path."},
                        ]} />
                        {json(`{
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
}`)}
                        <ErrorTable rows={[
                            {status: "401", code: "UNAUTHORIZED", message: "Bearer token is missing or invalid"},
                            {status: "401", code: "TOKEN_EXPIRED", message: "Bearer token has expired"},
                            {status: "404", code: "NOT_FOUND", message: "Path user ID is not a positive integer"},
                            {status: "500", code: "INTERNAL_ERROR", message: "Loading the user's film history failed"},
                        ]} />
                    </Endpoint>

                    <Endpoint method="PATCH" path="/users/new-password">
                        <p>Changes the authenticated password user’s password.</p>
                        <FieldTable rows={[
                            {name: "current_password", type: "string", required: "yes", description: "Existing password, 1-72 bytes."},
                            {name: "new_password", type: "string", required: "yes", description: "New password, 8-72 bytes and not a common password."},
                            {name: "new_password_confirm", type: "string", required: "no", description: "Compatibility field that must match new_password when present."},
                        ]} />
                        {json(`{
    "data": {
        "message": "Password has been changed"
    }
}`)}
                        <ErrorTable rows={[
                            {status: "400", code: "BAD_REQUEST", message: "Invalid JSON structure, unknown field, or oversized body"},
                            {status: "401", code: "UNAUTHORIZED", message: "Bearer token is missing or invalid"},
                            {status: "401", code: "TOKEN_EXPIRED", message: "Bearer token has expired"},
                            {status: "401", code: "INVALID_CURRENT_PASSWORD", message: "Current password is incorrect or changed concurrently"},
                            {status: "409", code: "PASSWORD_UNCHANGED", message: "New password equals the current password"},
                            {status: "500", code: "INTERNAL_ERROR", message: "Loading or updating the user failed"},
                        ]} />
                    </Endpoint>

                    <Endpoint method="PATCH" path="/users/{id}">
                        <p>Updates the authenticated user profile.</p>
                        <p>Password users can update identity and appearance fields. OAuth users can only update <code>color</code> and remove <code>profile_picture</code>.</p>
                        <FieldTable rows={[
                            {name: "id", type: "integer", required: "yes", description: "Path ID must match the bearer token user ID."},
                            {name: "email", type: "string", description: "Password users only."},
                            {name: "username", type: "string", description: "Password users only."},
                            {name: "first_name", type: "string", description: "Password users only."},
                            {name: "last_name", type: "string", description: "Password users only."},
                            {name: "profile_picture", type: "string or null", description: "Send null or an empty string to remove it; non-empty strings are rejected."},
                            {name: "color", type: "string", description: "One of yellow, pink, green, purple, blue, or red."},
                        ]} />
                        {json(`{
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
}`)}
                        <ErrorTable rows={[
                            {status: "400", code: "BAD_REQUEST", message: "Invalid JSON, unknown fields, or empty body"},
                            {status: "400", code: "VALIDATION_ERROR", message: "Invalid profile field or OAuth restriction"},
                            {status: "401", code: "UNAUTHORIZED", message: "Bearer token is missing or invalid"},
                            {status: "401", code: "TOKEN_EXPIRED", message: "Bearer token has expired"},
                            {status: "403", code: "FORBIDDEN", message: "Cannot update another user's profile"},
                            {status: "404", code: "NOT_FOUND", message: "The authenticated user no longer exists"},
                            {status: "409", code: "ALREADY_EXIST_ERROR", message: "Email or username is already in use"},
                            {status: "500", code: "INTERNAL_ERROR", message: "Loading or updating the user failed"},
                        ]} />
                    </Endpoint>
                </section>

                <section id="streaming-endpoints" className="scroll-mt-6 space-y-7">
                    <h2>Stream endpoints</h2>
                    <Endpoint method="GET" path="/stream/{id}">
                        <p>Initializes HLS streaming and delegates transcoding to torrent-stream.</p>
                        <FieldTable rows={[
                            {name: "id", type: "string", required: "yes", description: "IMDb ID of the movie."},
                        ]} />
                        <Note>Returns immediately if <code>/data/videos/{"{id}"}/</code> already exists.</Note>
                        <Table headers={["Code", "Meaning"]} rows={[
                            ["200", "Stream is ready or has been started"],
                            ["500", "Transcode service unreachable or failed"],
                        ]} />
                    </Endpoint>
                    <Endpoint method="GET" path="/stream/{id}/index">
                        <p>Returns the HLS playlist.</p>
                        <FieldTable rows={[
                            {name: "id", type: "string", required: "yes", description: "IMDb ID of the movie."},
                        ]} />
                        <Note color="green">Response: <strong>200 OK</strong> — <code>Content-Type: application/vnd.apple.mpegurl</code>.</Note>
                    </Endpoint>
                    <Endpoint method="GET" path="/stream/{id}/{segment}">
                        <p>Serves a single HLS transport stream segment.</p>
                        <FieldTable rows={[
                            {name: "id", type: "string", required: "yes", description: "IMDb ID of the movie."},
                            {name: "segment", type: "string", required: "yes", description: "Segment filename such as stream0.ts."},
                        ]} />
                        <Note color="green">Response: <strong>200 OK</strong> — <code>Content-Type: video/mp2t</code>.</Note>
                    </Endpoint>
                </section>

                <section id="content-endpoints" className="scroll-mt-6 space-y-7">
                    <h2>Movie and comment endpoints</h2>
                    <Endpoint method="GET" path="/movies">
                        <p>Returns tracker-wide popular movies.</p>
                        <Note>Only card fields are returned; full details live under <code>/movies/{"{id}"}</code>.</Note>
                        {json(`{
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
}`)}
                        <ErrorTable rows={[
                            {status: "500", code: "INTERNAL_ERROR", message: "Failed to load movies"},
                        ]} />
                    </Endpoint>
                    <Endpoint method="GET" path="/movies/featured">
                        <p>Returns a curated selection of movies.</p>
                        <Note>Only card fields are returned; full details live under <code>/movies/{"{id}"}</code>.</Note>
                        {json(`{
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
}`)}
                        <ErrorTable rows={[
                            {status: "500", code: "INTERNAL_ERROR", message: "Failed to load movies"},
                        ]} />
                    </Endpoint>
                    <Endpoint method="GET" path="/movies/directstream">
                        <p>Returns movies available for direct streaming.</p>
                        {json(`{
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
}`)}
                        <ErrorTable rows={[
                            {status: "500", code: "INTERNAL_ERROR", message: "Failed to load movies"},
                        ]} />
                    </Endpoint>
                    <Endpoint method="GET" path="/movies/search?title=&page=">
                        <p>Searches for movies by title and persists results after the first lookup.</p>
                        <FieldTable rows={[
                            {name: "title", type: "string", required: "yes", description: "Title to search for."},
                            {name: "page", type: "integer", defaultValue: "0", description: "Page index; 0 is first page."},
                        ]} />
                        {json(`{
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
}`)}
                        <ErrorTable rows={[
                            {status: "400", code: "VALIDATION_ERROR", message: "Title query parameter is required"},
                            {status: "500", code: "INTERNAL_ERROR", message: "Failed to search movies"},
                        ]} />
                    </Endpoint>
                    <Endpoint method="GET" path="/movies/watched">
                        <p>Returns watch history for the authenticated user.</p>
                        <Note>No request body is used. The user comes from the bearer token.</Note>
                        {json(`{
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
}`)}
                        <ErrorTable rows={[
                            {status: "400", code: "VALIDATION_ERROR", message: "Invalid request body"},
                            {status: "401", code: "UNAUTHORIZED", message: "Bearer token is missing or invalid"},
                            {status: "401", code: "TOKEN_EXPIRED", message: "Bearer token has expired"},
                            {status: "500", code: "INTERNAL_ERROR", message: "Failed to load movies"},
                        ]} />
                    </Endpoint>
                    <Endpoint method="PATCH" path="/movies/{imdbId}/progress">
                        <p>Saves playback progress for the authenticated user and movie.</p>
                        <FieldTable rows={[
                            {name: "imdbId", type: "string", required: "yes", description: "IMDb ID of the movie."},
                            {name: "progress", type: "integer", required: "yes", description: "Playback position in seconds, non-negative."},
                            {name: "pourcent", type: "integer", required: "yes", description: "Watched percentage between 0 and 100."},
                            {name: "complete", type: "boolean", required: "yes", description: "Marks whether the movie was fully watched."},
                        ]} />
                        {json(`{
    "data": {
        "progress": 1804,
        "pourcent": 54,
        "complete": false
    }
}`)}
                        <ErrorTable rows={[
                            {status: "400", code: "BAD_REQUEST", message: "Malformed JSON, unknown fields, or multiple JSON documents"},
                            {status: "400", code: "VALIDATION_ERROR", message: "Missing, null, wrong-type, or negative fields"},
                            {status: "401", code: "UNAUTHORIZED", message: "Bearer token is missing or invalid"},
                            {status: "401", code: "TOKEN_EXPIRED", message: "Bearer token has expired"},
                            {status: "404", code: "NOT_FOUND", message: "Movie ID is empty or unknown"},
                            {status: "500", code: "INTERNAL_ERROR", message: "Saving progress failed"},
                        ]} />
                    </Endpoint>
                    <Endpoint method="GET" path="/movies/{id}">
                        <p>Returns full metadata for a single movie.</p>
                        <FieldTable rows={[
                            {name: "id", type: "string", required: "yes", description: "IMDb ID of the movie."},
                            {name: "lang", type: "string", required: "no", defaultValue: "en", description: "Locale for the details: en, fr, or de."},
                        ]} />
                        {json(`{
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
        "extra_backdrops": ["string"]
    }
}`)}
                        <ErrorTable rows={[
                            {status: "404", code: "NOT_FOUND", message: "Movie not found"},
                            {status: "500", code: "INTERNAL_ERROR", message: "Failed to load movie"},
                            {status: "500", code: "INTERNAL_ERROR", message: "Failed to fetch movie details"},
                        ]} />
                    </Endpoint>
                    <Endpoint method="GET" path="/movies/{id}/torrents">
                        <p>Returns available torrent sources for a movie.</p>
                        <FieldTable rows={[
                            {name: "id", type: "string", required: "yes", description: "IMDb ID of the movie."},
                        ]} />
                        {json(`{
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
}`)}
                        <ErrorTable rows={[
                            {status: "404", code: "NOT_FOUND", message: "No tracker source found for this movie"},
                            {status: "500", code: "INTERNAL_ERROR", message: "Failed to load tracker source"},
                        ]} />
                    </Endpoint>
                    <Endpoint method="GET" path="/movies/{id}/comments">
                        <p>Returns comments posted on a movie, ordered by most recent first.</p>
                        <FieldTable rows={[
                            {name: "id", type: "string", required: "yes", description: "IMDb ID of the movie."},
                            {name: "page", type: "integer", defaultValue: "0", description: "Page index; 0 is first page."},
                        ]} />
                        {json(`{
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
}`)}
                        <ErrorTable rows={[
                            {status: "404", code: "NOT_FOUND", message: "Movie not found"},
                            {status: "500", code: "INTERNAL_ERROR", message: "Failed to access comments"},
                        ]} />
                    </Endpoint>
                    <Endpoint method="POST" path="/movies/{id}/comments">
                        <p>Posts a new comment on a movie as the authenticated user.</p>
                        <FieldTable rows={[
                            {name: "id", type: "string", required: "yes", description: "IMDb ID of the movie."},
                            {name: "content", type: "string", required: "yes", description: "Comment body."},
                        ]} />
                        {json(`{
    "data": {
        "id": 1,
        "user_id": 1,
        "movie_id": "string",
        "content": "string",
        "edited": false,
        "updated_at": "2026-05-06T12:00:00Z"
    }
}`)}
                        <ErrorTable rows={[
                            {status: "400", code: "VALIDATION_ERROR", message: "Invalid request body"},
                            {status: "401", code: "UNAUTHORIZED", message: "Bearer token is missing or invalid"},
                            {status: "401", code: "TOKEN_EXPIRED", message: "Bearer token has expired"},
                            {status: "404", code: "NOT_FOUND", message: "Movie not found"},
                            {status: "500", code: "INTERNAL_ERROR", message: "Failed to create comment"},
                        ]} />
                    </Endpoint>
                    <Endpoint method="GET" path="/comments">
                        <p>Returns all comments across all movies.</p>
                        <FieldTable rows={[
                            {name: "page", type: "integer", defaultValue: "0", description: "Page index; 0 is first page."},
                        ]} />
                        {json(`{
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
}`)}
                        <ErrorTable rows={[
                            {status: "500", code: "INTERNAL_ERROR", message: "Failed to load comments"},
                        ]} />
                    </Endpoint>
                    <Endpoint method="GET" path="/comments/{id}">
                        <p>Returns a single comment by its ID.</p>
                        <FieldTable rows={[
                            {name: "id", type: "string", required: "yes", description: "ID of the comment."},
                        ]} />
                        {json(`{
    "data": {
        "id": 1,
        "user_id": 2,
        "movie_id": "string",
        "content": "string",
        "edited": false,
        "updated_at": "2026-05-06T12:00:00Z"
    }
}`)}
                        <ErrorTable rows={[
                            {status: "404", code: "NOT_FOUND", message: "Comment not found"},
                            {status: "500", code: "INTERNAL_ERROR", message: "Failed to load comment"},
                        ]} />
                    </Endpoint>
                    <Endpoint method="PATCH" path="/comments/{id}">
                        <p>Updates the content of an existing comment if it belongs to the authenticated user.</p>
                        <FieldTable rows={[
                            {name: "id", type: "string", required: "yes", description: "ID of the comment."},
                            {name: "content", type: "string", required: "yes", description: "New comment body."},
                        ]} />
                        {json(`{
    "data": {
        "id": 1,
        "user_id": 2,
        "movie_id": "string",
        "content": "string",
        "edited": true,
        "updated_at": "2026-05-06T12:00:00Z"
    }
}`)}
                        <ErrorTable rows={[
                            {status: "400", code: "VALIDATION_ERROR", message: "Invalid request body"},
                            {status: "401", code: "UNAUTHORIZED", message: "Bearer token is missing or invalid"},
                            {status: "401", code: "TOKEN_EXPIRED", message: "Bearer token has expired"},
                            {status: "404", code: "NOT_FOUND", message: "Comment not found"},
                            {status: "500", code: "INTERNAL_ERROR", message: "Failed to update comment"},
                        ]} />
                    </Endpoint>
                    <Endpoint method="DELETE" path="/comments/{id}">
                        <p>Deletes a comment if it belongs to the authenticated user.</p>
                        <FieldTable rows={[
                            {name: "id", type: "string", required: "yes", description: "ID of the comment."},
                        ]} />
                        {json(`{
    "data": null
}`)}
                        <ErrorTable rows={[
                            {status: "401", code: "UNAUTHORIZED", message: "Bearer token is missing or invalid"},
                            {status: "401", code: "TOKEN_EXPIRED", message: "Bearer token has expired"},
                            {status: "404", code: "NOT_FOUND", message: "Comment not found"},
                            {status: "500", code: "INTERNAL_ERROR", message: "Failed to delete comment"},
                        ]} />
                    </Endpoint>
                </section>
            </div>
        </div>
    </main>);
}