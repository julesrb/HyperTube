```md
# Web — Hypertube

**Résumé:** A web app for the 21th century  
**Version:** 6.3

---

# Table of Contents

- [I Introduction](#i-introduction)
- [II General Instructions](#ii-general-instructions)
- [III Mandatory Part](#iii-mandatory-part)
  - [III.1 User Interface](#iii1-user-interface)
  - [III.2 Library Part](#iii2-library-part)
    - [III.2.1 Search](#iii21-search)
    - [III.2.2 Thumbnails](#iii22-thumbnails)
  - [III.3 Video Part](#iii3-video-part)
  - [III.4 API](#iii4-api)
- [IV Bonus Part](#iv-bonus-part)
- [V Submission and Peer-evaluation](#v-submission-and-peer-evaluation)
  - [V.1 Eliminatory Rules](#v1-eliminatory-rules)

---

# I Introduction

This project aims to create a web application that enables user’s to search for and watch videos.

The player will integrated directly into the site, and the videos will be downloaded using the BitTorrent protocol.

To enhance search capabilities, the research engine will query at least two external sources of your choice.

Once a selection is made, the video will be downloaded from the server and streamed on the web player simultaneously.

This means that the player will not only display the video after the download is complete but also stream the video feed directly.

---

# II General Instructions

- For this project you are free to use any programming language of your choice.

- All frameworks, micro-frameworks, libraries, etc. are allowed, except for those that are used to create a video stream from a torrent. This restriction is to ensure that the educational purpose of the project is not compromised. For example, libraries such as `webtorrent`, `pulsar` and `peerflix` are not permitted.

- You are free to use any web server of your choice, such as Apache, Nginx or even a built-in web server.

- Your entire application must be compatible with the latest versions of Firefox and Chrome.

- Your website must have a decent layout including at least a header, a main section and a footer.

- Your website must be usable on a mobile phone and maintain an acceptable layout on small resolutions.

- All your forms must have correct validations and the entire website must be secure. This part is mandatory and will be extensively checked during the defense.

To give you an idea, here are a few elements that are not considered secure:

- Storing a “plain text” password in your database.
- Allowing injection of HTML or “user” Javascript code in unprotected variables.
- Allowing the upload of unwanted content.
- Allowing alteration of an SQL request.

> ⚠️ For obvious security reasons, any credentials, API keys, environment variables, etc. must be saved locally in a `.env` file and excluded from git. Storing credentials publicly will result in automatic failure of the project.

---

# III Mandatory Part

You will need to create a web application with the following features:

---

# III.1 User Interface

- The app must allow a user to register asking for at least their:
  - email address
  - username
  - last name
  - first name
  - password (protected)

- The user must be able to register and log in via Omniauth.

You must implement at least 2 strategies:
- the 42 strategy
- another one of your choice

- The user must be able to log in with their username and password.

- They must be able to receive an email allowing them to reset their password should they forget it.

- The user must be able to log out with one click from any pages on the site.

- The user must be able to select a preferred language that will default to English.

A user must also be able to:

- Modify their email address, profile picture and information.
- View the profile of any other user, including their profile picture and information.

However, the email address will remain private.

---

# III.2 Library Part

> ⚠️ This section can only be accessed by authenticated users.

This section must have at a minimum:

- A search field.
- A list of video thumbnails.

---

# III.2.1 Search

The search engine will query at least two external sources (of your choice) that exclusively provide video content, and display the results in the form of thumbnails.

---

# III.2.2 Thumbnails

- If search has been done, the results will be displayed as thumbnails, sorted by names.

- If no research was done, the app will display the most popular video from the external sources, sorted by the criteria of your choice:
  - downloads
  - peers
  - seeders
  - etc.

- Each thumbnail must display:
  - the name of the video
  - its production year (if available)
  - its IMDb rating (OMDb or TMDb for free API)
  - a cover image

- Watched and unwatched videos should be differentiated in the thumbnails.

- The list will be paginated, with the next page being loaded asynchronously as the user scrolls down.

There should be no link to load the next page.

- The page will be sortable and filterable according to criteria such as:
  - name
  - genre
  - IMDb grade
  - production year
  - etc.

---

# III.3 Video Part

> ⚠️ This section can only be accessed by authenticated users.

- This section will present the details of a video, including:
  - a video player
  - summary (if available)
  - casting
    - producer
    - director
    - main cast
  - production year
  - length
  - IMDb rating
  - cover image
  - anything else relevant

- Users will have the option of leaving a comment on the video, and the list of prior comments will be shown.

- To launch the video on the server:
  - if the file was not downloaded prior,
  - the associated torrent on the server will be launched,
  - and the video stream will be initiated as soon as enough data has been downloaded to ensure a seamless watching experience.

Any treatment must be done in the background in a non-blocking manner.

- Once the movie is entirely downloaded, it will be saved on the server to avoid the need to re-download it in the future.

However, if a movie is unwatched for a month, it will be erased.

- If English subtitles are available for the video, they will be downloaded and made available for the video player.

Additionally:
- if the language of the video does not match the preferred language of the user
- and subtitles are available

then the subtitles will be downloaded and selectable.

- If the video is not natively readable for the browser (i.e. not in mp4 or webm format), it will be converted on the fly into an acceptable format.

At minimum, `mkv` support is required.

---

# III.4 API

Develop a RESTful API with an OAuth2 authentication that can be used to obtain basic information about this project.

- Authenticated users are allowed to retrieve or update any profiles.

- Any user can access the website’s “front page”, which displays basic information about the top movies.

- A `GET` request on a movie should return all the relevant information that has been previously collected.

- Authenticated users can access user comments via:
  - `/comments/:id`
  - `/movie/:id/comments`

They can also post a comment using an appropriate payload.

- Any other API call should not be usable.

Return the appropriate HTTP code.

---

## Basic Documentation

### POST `/oauth/token`

Expects:
- client
- secret

Returns:
- auth token

---

### GET `/users`

Returns:
- list of users
- id
- username

---

### GET `/users/:id`

Returns:
- username
- email address
- profile picture URL

---

### PATCH `/users/:id`

Expected data:
- username
- email
- password
- profile picture URL

---

### GET `/movies`

Returns:
- list of movies available on the frontpage
- id
- name

---

### GET `/movies/:id`

Returns:
- movie name
- id
- IMDb mark
- production year
- length
- available subtitles
- number of comments

---

### GET `/comments`

Returns:
- latest comments
- author username
- date
- content
- id

---

### GET `/comments/:id`

Returns:
- comment
- author username
- comment id
- date posted

---

### PATCH `/comments/:id`

Expected data:
- comment
- username

---

### DELETE `/comments/:id`

Deletes a comment.

---

### POST `/comments`
OR

### POST `/movies/:movie_id/comments`

Expected data:
- comment
- movie_id

The rest is filled by the server.

> ℹ️ During the evaluation, you will be asked to provide evidence that your API is truly RESTful.

---

# IV Bonus Part

If the mandatory part is completed perfectly, you can now add any bonus features you wish.

They will be evaluated at the discretion of your evaluators but you must still adhere to the basic constraints.

For instance, downloading a torrent must occur on the server side in the background.

If you are needing some inspiration, here are a few ideas:

- Some additional Omniauth strategies.
- Manage various video resolutions.
- Stream the video via the MediaStream API.
- More API routes to add, delete movies, etc.

> ⚠️ The bonus part will only be assessed if the mandatory part is PERFECT.

By perfect, we mean that:
- all mandatory requirements have been fully implemented
- everything functions without malfunctioning

If any mandatory requirement has not been met, your bonus part will not be evaluated.

---

# V Submission and Peer-evaluation

Turn in your assignment in your Git repository as usual.

Only the work inside your repository will be evaluated during the defense.

Don’t hesitate to double check the names of your folders and files to ensure they are correct.

The following instructions will be part of your defense.

Be cautious when you apply them as they will be graded with a non-negotiable 0.

---

# V.1 Eliminatory Rules

- Your code cannot produce any:
  - errors
  - warnings
  - notices

either from the server or the client side in the web console.

- Anything not specifically authorized is forbidden.

- The slightest security breach will give you 0.

You must at least:
- NOT have plain text passwords stored in your database
- be protected against SQL injections
- validate all forms and uploads
```

## Implementation Status (reviewed May 19, 2026)

This section tracks the current repository state against the Hypertube subject. It is based on the code currently present in this repository, not on full runtime verification of every external service.

### Already Done

- Project structure and local runtime:
  - Go API service under `services/api`.
  - Go torrent-stream service under `services/torrent-stream`.
  - Next.js frontend under `frontend`.
  - PostgreSQL schema and seed migrations under `db`.
  - Docker Compose and Makefile targets for the API, stream service, frontend, and database.

- Database:
  - Tables exist for movies, torrents, featured movies, users, OAuth accounts, watch history, direct-stream movies, comments, cached movie searches, and password-reset tokens.
  - Credentials are expected through `.env`; no API secrets were found hardcoded in the inspected source.

- Backend authentication:
  - Registration stores email, username, first name, last name, and a bcrypt password hash.
  - Email/password login returns a JWT bearer token.
  - JWT middleware protects most movie and comment routes.
  - The 42 OAuth backend flow exists with state-cookie protection, provider token exchange, profile fetch, local user creation/linking, and frontend callback redirect support.
  - OAuth2-style `POST /oauth/token` password grant exists and accepts JSON or form-encoded request bodies.
  - Password-reset request and reset endpoints exist with hashed reset tokens, expiry handling, and optional Brevo email delivery.

- Backend movie/library API:
  - Public `GET /api/v1/movies` returns featured movies.
  - Authenticated routes exist for movie search, movie details, movie torrents, watched movies, direct-stream movies, and movie comments.
  - Search queries C411 and archive.org, enriches results through TMDB, stores movies/torrents, and caches search results with pagination.
  - Startup seeding loads top C411 movies and stores featured movies.

- Backend comments:
  - Movie comment listing and creation are implemented.
  - Generic comment list/get/update/delete endpoints exist.
  - Update and delete operations are scoped to the authenticated user.

- Frontend:
  - Responsive Next.js UI exists with navigation, movie home, movie search/list, movie details, comments, profile, auth modals, notifications, and pagination components.
  - Internationalization is set up for English, French, and German with locale routing.
  - Local UI flows exist for register/sign-in, forgot password, profile editing, avatar color/image selection, comment create/edit/delete, movie filtering/sorting, watch history display, and language selection.
  - Static image and font assets are present and wired into the UI.

- Torrent/streaming:
  - Separate Go stream service exists with health, stream, HLS index, and segment endpoints.
  - HLS segmentation through ffmpeg is implemented as a proof of concept against a bundled test MP4.
  - A transcode unit test file exists for the HLS helper.

- Documentation, tests, and scripts:
  - API and auth documentation exists.
  - API shell tests and user-story demo scripts exist for auth, OAuth token, password reset, comments, featured movies, search fallback, and 42 auth.
  - Go test files exist for auth, comments, movies, JWT/middleware, password validation/reset, OAuth, and transcoding.

### Still Missing / Incomplete

- Frontend-to-backend integration:
  - `frontend/src/services/auth.ts` and `frontend/src/services/movies.ts` are empty.
  - Frontend sign-in/register flows use static fixtures and a local `coucou` token instead of the Go API.
  - Frontend movies, comments, profile, watch history, filters, and password-reset UI are backed by fixture/local state, not by API calls.
  - No frontend page was found for the OAuth callback or for a reset-password form that consumes a reset token.
  - No visible frontend 42-login or second OAuth-provider button is wired.

- User/profile API:
  - `services/api/internal/users/handler.go` contains empty handlers.
  - User routes are not registered in `services/api/main.go`.
  - Profile retrieval/update, email privacy, profile-picture upload/storage, and public user profile pages are therefore not complete on the backend.

- Required authentication coverage:
  - The subject requires at least two OAuth/Omniauth strategies. Only 42 OAuth is implemented in the backend.
  - `/auth/login` currently validates email/password, while the subject explicitly requires username/password login. The OAuth password grant can look up by login, but the normal login endpoint and frontend do not yet cover this requirement cleanly.

- Movie library behavior:
  - Infinite scroll is not implemented; the frontend uses a manual pagination component.
  - Watched/unwatched differentiation is only represented in local fixture logic, not persisted end-to-end from playback.
  - Sorting, filtering, and searching are frontend-only over static data. Backend search has pagination but does not expose all required sort/filter criteria.
  - Featured/top movies exist, but production behavior depends on external APIs and seeded database state.

- Video playback and torrent pipeline:
  - The torrent client is still a stub returning `nil, nil`.
  - Stream endpoints currently transcode a local test MP4 and do not use the requested movie ID or torrent URL.
  - No server-side BitTorrent download, sequential piece prioritization, completed-movie caching, or one-month unwatched cleanup is implemented.
  - Browser playback is not wired from the movie detail page; the play action currently logs to the console.
  - Subtitle download and selectable subtitle tracks are not implemented.
  - On-the-fly conversion for non-browser formats, especially MKV from torrent input, is not implemented end-to-end. The current ffmpeg path copies a local MP4 into HLS segments.
  - VPN/Gluetun routing is documented but commented out in `docker-compose.yml`.

- REST API completeness:
  - The subject routes for `/users`, `/users/:id`, profile update, and `POST /comments` are missing or not registered.
  - `POST /movies/:movie_id/comments` exists as `POST /api/v1/movies/{id}/comments`.
  - Movie details are implemented, but runtime fields such as watched/progress are currently hardcoded in the response DTO.
  - Some documented API behavior does not match the current router exactly and should be reconciled before defense.

- Security and validation gaps to finish:
  - Backend has parameterized SQL and password hashing, but many content inputs need stricter validation and length checks.
  - Frontend forms do not yet rely on backend validation responses.
  - Profile-picture upload validation/storage is not implemented.
  - Comment content is not trimmed or length-validated consistently server-side.

- Verification:
  - Full local verification could not be completed in this shell on May 19, 2026 because `go` is not installed.
  - `npm run lint` could not run with this shell's Node.js `v12.22.9`; the installed frontend tooling requires a newer Node version.
  - Before defense, run Go tests with a working Go toolchain, run frontend lint/build with a supported Node version, and run the API shell/user-story scripts against Docker Compose.
