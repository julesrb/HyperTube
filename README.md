# HyperTube

A web application to search for and stream movies downloaded via Torrent, playback begins before the download completes.

Built as a 42 Berlin project by Jules Bernard, Omio Ohoro and Florian Guiramand.

---

## Concept

A user searches for a movie. Results are pulled from two torrent sources (archive.org and C411) and enriched with metadata from TMDb (poster, rating, cast, year, synopsis). The user clicks play. The torrent service starts downloading with the BitTorrent client, transcodes the file to HLS on the fly with ffmpeg, and the browser begins playing within seconds via `hls.js`. All torrent traffic is routed through a VPN. Watch progress is tracked per user, and downloaded videos are deleted automatically after 30 days.

---

## Stack

| Layer | Technology |
|---|---|
| Frontend | Next.js 16 / React 19 (App Router, `next-intl`, TanStack Query, Tailwind 4, `hls.js`) |
| API | Go (chi router) |
| Torrent + transcode | Go (`anacrolix/torrent` + `ffmpeg`) |
| Database | PostgreSQL 16 |
| VPN | Gluetun + OpenVPN (Private Internet Access) |
| Local dev | Docker Compose |

---

## Architecture

```
browser (Next.js in srcs/requirements/frontend + hls.js)
  │  REST  /api/v1/*
  ▼
api service (Go, srcs/requirements/api, :8080)
  ├── auth — email/password, 42 / GitHub / GitLab OAuth, JWT + refresh tokens
  ├── users, comments, watch history & progress
  ├── movie search + metadata (archive.org, C411, TMDb)
  └── stream — serves the HLS playlist + segments from the shared /data volume
        │  HTTP (internal)  →  torrent-transcode (:8081, behind VPN)
        ▼
torrent-transcode service (Go, srcs/requirements/torrent-transcode)
  ├── anacrolix/torrent — sequential piece download
  ├── ffmpeg — transcode to HLS (H.264 / AAC, 5s segments)
  └── cleanup — removes unfinished torrents on restart, old videos after 30 days

ALL torrent traffic shares the VPN container's network namespace (Gluetun / PIA)
```

The `api` and `torrent-transcode` services share the `srcs/data` volume:
`torrent-transcode` downloads and transcodes into `srcs/data/videos/{id}/`, and the `api`
stream handler serves those HLS files to the browser.

---

## The Streaming Pipeline

```
magnet / .torrent
  → anacrolix/torrent (sequential pieces)
  → ffmpeg -i <source> → HLS  (stream.m3u8 + stream0.ts, stream1.ts, …)
  → api /stream/{id}/index  and  /stream/{id}/{segment}
  → browser <video> via hls.js
```

Transcoding to HLS (rather than fragmented MP4) lets the browser play arbitrary input
codecs — H.264, H.265/HEVC and AV1 sources are all transcoded to a broadly playable
H.264/AAC HLS stream. Subtitles are extracted/served separately, with an optional
synchronized-track mode on the frontend.

---

## VPN

The `torrent-transcode` service shares the VPN container's network namespace, so all of
its traffic goes through the tunnel.

```yaml
vpn:
  image: qmcgaw/gluetun
  cap_add: [NET_ADMIN]
  environment:
    VPN_SERVICE_PROVIDER: private internet access
    VPN_TYPE: openvpn
    OPENVPN_USER: ${PIA_USER}
    OPENVPN_PASSWORD: ${PIA_PASSWORD}
    SERVER_REGIONS: ${PIA_REGION:-france}

torrent-transcode:
  network_mode: "service:vpn"
  depends_on:
    vpn:
      condition: service_healthy
```


---

## Running it

Generate the environment file and fill in any provider-specific secrets:

```bash
make env
# srcs/.env is generated from .env.exemple
```

Then use the Makefile:

```bash
make             # build & start the stack
make up          # compose up
make build       # build images
make down        # stop containers
make logs        # follow all logs
make logs SERVICE=api   # follow one service
make exec SERVICE=api   # open a shell in one service
```

The frontend container is not started by Compose; run it locally:

```bash
cd srcs/requirements/frontend && npm run dev  # http://localhost:4200
```

PostgreSQL initializes its schema from `srcs/requirements/db/` only when its volume is
first created.
After a schema change, recreate the database with `make vclean` or `make re`.

| Service | Port |
|---|---|
| `postgres` | 5432 |
| `api` | 8080 |
| `torrent-transcode` | 8081 (reached via the `vpn` container) |
| `frontend` (run locally) | 4200 |

### Useful dev commands

```bash
make ps           # list containers
make detach       # compose up in detached mode
make clean        # remove containers and local images
make vclean       # remove containers, volumes, env, and data
make fclean       # remove everything, including images
make re           # reset env/data and start again
```

---

## Repository Structure

```
/
├── README.md
├── Makefile
├── .env.exemple
├── launch.d/
│   └── 01generatePasswordsAndKeys.sh
└── srcs/
    ├── docker-compose.yml
    ├── .env
    ├── data/                 # shared volume: torrents/ and videos/ (git-ignored)
    └── requirements/
        ├── api/              # Go REST API (chi)
        │   ├── main.go
        │   └── internal/
        │       ├── auth/     # password + 42/GitHub/GitLab OAuth, JWT, refresh, reset
        │       ├── comments/
        │       ├── downloader/
        │       ├── email/
        │       ├── i18n/
        │       ├── models/
        │       ├── movies/
        │       ├── requestjson/
        │       ├── respond/
        │       ├── stream/
        │       ├── userinput/
        │       └── users/
        ├── db/              # SQL schema + seeds (auto-loaded by postgres)
        │   ├── 001_schema.sql
        │   ├── 002_seed_dev.sql
        │   └── 007_seed_curated_movies.sql
        ├── frontend/        # Next.js app
        │   ├── messages/{de,en,fr}.json
        │   ├── public/
        │   └── src/
        │       ├── app/[locale]/
        │       ├── components/{features,layout,ui}
        │       ├── contexts/
        │       ├── hooks/
        │       ├── i18n/
        │       ├── services/
        │       ├── types/
        │       └── utils/
        └── torrent-transcode/
            └── internal/
                ├── cleanup/
                ├── torrent/
                └── transcode/
```

---

## API

Base path: `/api/v1`. Most content routes are JWT-protected. Public routes: health,
movie listing/search, registration, login, refresh, password-reset, OAuth start/callback,
and the OAuth2 token endpoint. Detailed auth request/response examples live in
`srcs/requirements/api/README.md`.

```
GET    /health

POST   /auth/register
POST   /auth/login
POST   /auth/refresh-token
POST   /auth/password-reset/send-email
POST   /auth/password-reset/set-new-password
GET    /auth/42/login        GET /auth/42/callback
GET    /auth/github/login    GET /auth/github/callback
GET    /auth/gitlab/login    GET /auth/gitlab/callback
POST   /oauth/token
POST   /oauth/applications   GET /oauth/applications
PATCH  /oauth/applications/{id}
DELETE /oauth/applications/{id}

GET    /movies                       # default/curated list (public)
GET    /movies/featured              # public
GET    /movies/search                # public
GET    /movies/watched               # auth
GET    /movies/directstream          # auth
GET    /movies/{id}                  # auth
PATCH  /movies/{id}/progress         # auth — watch progress (percentage)
GET    /movies/{id}/torrents         # auth
GET    /movies/{id}/comments         # auth
POST   /movies/{id}/comments         # auth

GET    /comments        GET /comments/{id}
PATCH  /comments/{id}   DELETE /comments/{id}

GET    /users           GET /users/{id}
GET    /users/{id}/comments
GET    /users/{id}/movie-history
PATCH  /users/new-password
PATCH  /users/{id}

GET    /stream/{id}                  # init download + transcode
GET    /stream/{id}/index            # HLS playlist (stream.m3u8)
GET    /stream/{id}/{segment}        # HLS segment (.ts)
```

---

## Auth

- Registration: email, username, first name, last name, password (bcrypt)
- 42, GitHub and GitLab OAuth (when provider client credentials are configured);
  browser provider flows request fixed profile scopes: 42 `public`, GitHub
  `read:user user:email`, GitLab `read_user`
- JWT issued on login / OAuth callback, validated on every protected request, with refresh tokens
- User-managed OAuth applications at `/api/v1/oauth/applications` and the
  OAuth2 token endpoint at `POST /api/v1/oauth/token` support registered
  `client_credentials` grants. Client credentials issue a normal owner-user
  bearer token and do not support application or token scopes.
- Password reset by email via Brevo (when configured)

---

## Security

Eliminatory per the 42 subject — any breach scores zero:

- Passwords hashed (bcrypt)
- Parameterized queries (no SQL injection)
- Escaped output (no XSS)
- Form and input validation
- Credentials in `.env`, never committed
