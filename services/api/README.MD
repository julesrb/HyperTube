doc for the API

---

## GET /movies

Returns a curated list of featured movies.

### Response

```json
{
  "data": [
    {
      "imdb_id": "string",
      "title": "string",
      "year": "string",
      "poster_url": "string",
      "backdrop_url": "string"
    }, 
    ...
  ],
  "meta": {
    "total": 11,
    "page": 1,
    "per_page": 11
  }
}
```

> Only `imdb_id`, `title`, `year`, `poster_url`, and `backdrop_url` are returned. Full details are available via `GET /movies/:id`.

### Error response

```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "failed to load movies"
  }
}
```

---

## GET /movies/:id

Returns full metadata for a single movie. Summary, director, and cast are fetched live from TMDB.

### Path parameters

| Parameter | Type   | Description            |
|-----------|--------|------------------------|
| `id`      | string | IMDb ID of the movie   |

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
    "runtime_minutes": 167,
    "summary": "string",
    "director": "string",
    "cast": ["string"],
    "watched": false,
    "progression": 0.0
  }
}
```

### Error responses

```json
{ "error": { "code": "NOT_FOUND", "message": "movie not found" } }
```
```json
{ "error": { "code": "INTERNAL_ERROR", "message": "failed to load movie" } }
```
```json
{ "error": { "code": "INTERNAL_ERROR", "message": "failed to fetch movie details" } }
```

---

## GET /movies/search?title=

Searches for movies by title across tracker sources, enriches results with TMDB metadata, and persists new entries to the database.

### Query parameters

| Parameter | Type   | Required | Description          |
|-----------|--------|----------|----------------------|
| `title`   | string | yes      | Title to search for  |

### Response

```json
{
  "data": [
    {
      "imdb_id": "string",
      "title": "string",
      "year": "string",
      "poster_url": "string",
      "backdrop_url": "string"
    }
  ],
  "meta": {
    "total": 10,
    "page": 1,
    "per_page": 10
  }
}
```

> Results are capped at 10 to stay within TMDB rate limitsfor now.

### Error responses

```json
{ "error": { "code": "VALIDATION_ERROR", "message": "title query parameter is required" } }
```
```json
{ "error": { "code": "INTERNAL_ERROR", "message": "failed to search movies" } }
```

---

## GET /movies/:id/torrents

> **Not yet implemented.**

Returns available torrent sources for a movie. Sources are resolved by crawling the tracker URLs stored during search.

### Path parameters

| Parameter | Type   | Description          |
|-----------|--------|----------------------|
| `id`      | string | IMDb ID of the movie |

### Planned response

```json
{
  "data": [
    {
      "url": "string",
      "quality": "string",
      "size": "string",
      "language": "string",
      "seeds": "string",
      "source": "string"
    }
  ],
  "meta": {
    "total": 11,
    "page": 1,
    "per_page": 11
  }
}
```

---

## GET /movies/:id/subtitles

> **Not yet implemented.**

Returns available subtitle tracks for a movie.

### Path parameters

| Parameter | Type   | Description          |
|-----------|--------|----------------------|
| `id`      | string | IMDb ID of the movie |

### Planned response

```json
{
  "data": [
    {
      "url": "string",
      "language": "string"
    }
  ],
  "meta": {
    "total": 11,
    "page": 1,
    "per_page": 11
  }
}
```

---

## GET /movies/:id/comments

> **Not yet implemented.**

Returns comments posted on a movie.

### Path parameters

| Parameter | Type   | Description          |
|-----------|--------|----------------------|
| `id`      | string | IMDb ID of the movie |

### Planned response

```json
{
  "data": [
    {
      "id": "string",
      "content": "string",
      "posted_at": "2025-12-17T00:00:00Z",
      "author": {
        "id": "string",
        "username": "string",
        "picture_url": "string"
      }
    }
  ],
  "meta": {
    "total": 5,
    "page": 1,
    "per_page": 5
  }
}
```

### Planned error responses

```json
{ "error": { "code": "NOT_FOUND", "message": "movie not found" } }
```
