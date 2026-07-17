# torrent-transcode

Internal service responsible for torrent downloading and HLS transcoding.  
**Not exposed to the internet** — called only by the `api` service over the Docker internal network.

**Default port:** `8081` (override with `PORT`)

---

## GET /health

Health check.

**Response:** `200 OK` — no body.

---

## POST /download/{id}

*(Not yet implemented)* Start a torrent download and prepare for transcoding.

| Parameter | Description          |
|-----------|----------------------|
| `id`      | IMDb ID of the movie |

---

## POST /transcode/{id}

Start the HLS transcoding pipeline for an already-downloaded torrent.

| Parameter | Description          |
|-----------|----------------------|
| `id`      | IMDb ID of the movie |

### How it works

1. Looks for the source file at `/data/torrents/{id}/rubber.mp4` *(path is hardcoded for now)*.
2. Creates the output directory `/data/videos/{id}/`.
3. Runs `ffmpeg` to segment the file into 5-second HLS chunks, writing `stream.m3u8` and `stream0.ts`, `stream1.ts` … into the output directory.
4. Returns `200` when done.

### Response

`200 OK` — no body.

### Error responses

```
500 Internal Server Error — "failed to create stream directory"
500 Internal Server Error — "failed to start stream"
```

---

## Environment variables

| Variable | Default | Description      |
|----------|---------|------------------|
| `PORT`   | `8081`  | Listening port   |

---

## Data volumes

| Path              | Description                                       |
|-------------------|---------------------------------------------------|
| `/data/torrents/` | Source torrent files — read by this service       |
| `/data/videos/`   | HLS output — written by this service, read by api |
