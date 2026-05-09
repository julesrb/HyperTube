Torrent-stream service documentation

Base path: *(no prefix — routes are mounted at root)*

Default port: `8081`

---

## GET /health

Health check.

### Response

`200 OK` — no body.

---

## GET /stream/{id}

Starts the HLS torrenting and transcoding pipeline for a movie and waits until the output playlist is ready. The client should poll `GET /stream/{id}/index` once this returns 200.

### Path parameters

| Parameter | Type   | Description                       |
|-----------|--------|-----------------------------------|
| `id`      | string | IMDb ID of the movie to stream    |

### How it works

1. Locates the source file for `{id}`.
2. Runs `ffmpeg` to segment it into 5-second HLS chunks (`-hls_time 5`, codec copy).
3. Returns `200` when the playlist file is written and the first segments are ready.

### Response

`200 OK` — no body.

### Error responses

```
500 Internal Server Error — "failed to start stream"
```

---

## GET /stream/{id}/index

Returns the HLS master playlist (`index.m3u8`) for the given stream.

The browser `<video>` element (or any HLS-capable player) fetches this file first, then requests segments listed inside it.

### Path parameters

| Parameter | Type   | Description                    |
|-----------|--------|--------------------------------|
| `id`      | string | IMDb ID of the active stream   |

### Response

`200 OK`

```
Content-Type: application/vnd.apple.mpegurl
```

Body is the raw `.m3u8` playlist:

```
#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:5
#EXTINF:5.000000,
index0.ts
#EXTINF:5.000000,
index1.ts
...
#EXT-X-ENDLIST
```

### Error responses

```
500 Internal Server Error — "failed to read index file"
```

---

## GET /stream/{id}/{segment}

Serves a single HLS transport-stream segment (`.ts` file).

The player fetches these URLs automatically from the playlist returned by `GET /stream/{id}/index`. Clients do not need to construct segment URLs manually.

### Path parameters

| Parameter  | Type   | Description                              |
|------------|--------|------------------------------------------|
| `id`       | string | IMDb ID of the active stream             |
| `segment`  | string | Segment filename, e.g. `index0.ts`       |

### Response

`200 OK`

```
Content-Type: video/mp2t
```

Body is the raw binary `.ts` segment.

### Error responses

```
500 Internal Server Error — "failed to read segment file"
```

