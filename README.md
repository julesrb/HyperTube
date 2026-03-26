# HyperTube

A web application to search for and stream videos downloaded via BitTorrent — streaming begins before the download completes.

Built as a 42 Berlin project.

---

## Concept

A user searches for a movie. Results are pulled from at least two legal torrent sources (archive.org, legittorrents.info) and enriched with metadata from TMDb (poster, rating, cast, year). The user clicks play. The server starts downloading the torrent with sequential piece prioritization and begins streaming to the browser immediately. Completed movies are cached on the server and erased after 30 days unwatched.

---

## Stack

| Layer | Technology |
|---|---|
| Frontend | React (Next.js) **or** Angular 17+ — TBD |
| API | Go |
| Torrent + stream | Go |
| Database | PostgreSQL |
| VPN | Gluetun + WireGuard |
| Local dev | Docker Compose |

---

## Architecture

```
browser
  └── frontend (React or Angular)
        │
        │ REST
        ▼
      API service (Go)
        ├── auth (OAuth2 — 42 + GitHub, JWT)
        ├── users, comments, watch history
        └── movie metadata (TMDb)
              │
              ▼
        Torrent stream service (Go)
              ├── anacrolix/torrent — sequential piece download
              ├── io.Pipe → ffmpeg — MKV → MP4/WebM on the fly
              └── chunked HTTP → browser <video>

        ALL torrent traffic routes through VPN (Gluetun/WireGuard)
```

---

## The Torrent Pipeline

```
magnet link
  → anacrolix/torrent (SetReadahead)
  → io.Pipe reader
  → ffmpeg -i pipe:0 -movflags frag_keyframe+empty_moov -f mp4 pipe:1
  → chunked HTTP response
  → browser <video>
```

`-movflags frag_keyframe+empty_moov` is required — without it ffmpeg writes the `moov` atom at the end and the browser cannot play until fully downloaded.

---

## VPN

The torrent service shares the VPN container's network namespace. All its traffic goes through the tunnel.

```yaml
vpn:
  image: qmcgaw/gluetun
  cap_add: [NET_ADMIN]

torrent-stream:
  network_mode: "service:vpn"
  depends_on:
    vpn:
      condition: service_healthy
```

VPN is opt-in locally via `--profile vpn` to keep dev iteration fast.

---

## Docker Compose

```bash
docker compose up           # dev, no VPN
docker compose --profile vpn up   # with VPN
```

| Container | Port |
|---|---|
| `postgres` | 5432 |
| `api` | 8080 |
| `torrent-stream` | 8081 (via vpn container when profile active) |
| `frontend` | 4200 |

---

## Repository Structure

```
/
├── README.md
├── docker-compose.yml
├── Makefile
├── services/
│   ├── api/
│   │   ├── Dockerfile
│   │   ├── main.go
│   │   └── internal/
│   │       ├── auth/
│   │       ├── users/
│   │       ├── movies/
│   │       └── comments/
│   └── torrent-stream/
│       ├── Dockerfile
│       ├── main.go
│       └── internal/
│           ├── torrent/
│           ├── stream/
│           └── transcode/
├── shared/
│   ├── db/
│   └── models/
├── frontend/
│   └── Dockerfile
└── .github/
    └── workflows/
        └── ci.yml
```

---

## API

OAuth2-protected. `GET /movies` is the only public endpoint.

```
POST   /oauth/token

GET    /users
GET    /users/:id
PATCH  /users/:id

GET    /movies
GET    /movies/:id
GET    /movies/:id/comments
POST   /movies/:id/comments

GET    /comments
GET    /comments/:id
POST   /comments
PATCH  /comments/:id
DELETE /comments/:id
```

---

## Auth

- Registration: email, username, first name, last name, password (bcrypt)
- OAuth2: 42 intranet + GitHub
- Password reset via email
- JWT on login, validated on every protected request

---

## Security

Eliminatory per the 42 subject — any breach scores zero:

- Passwords hashed (bcrypt)
- Parameterized queries (no SQL injection)
- Escaped output (no XSS)
- Form and upload validation
- Credentials in `.env`, never committed