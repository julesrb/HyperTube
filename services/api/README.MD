API documentation

Base path: `/api/v1`

---

## GET /health

Health check.

### Response

`200 OK` — no body.

---

## GET /movies

Returns the curated list of featured movies.

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
  ]
}
```

> Only the card fields are returned. Full details are available via `GET /movies/{id}`.

### Error responses

```json
{ "error": { "code": "INTERNAL_ERROR", "message": "failed to load movies" } }
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
{ "error": { "code": "VALIDATION_ERROR", "message": "title query parameter is required" } }
```
```json
{ "error": { "code": "INTERNAL_ERROR", "message": "failed to search movies" } }
```

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
| `lang`    | string | no       | `en-US`  | TMDB language code for the details |

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
    "director": ["string"],
    "cast": ["string"],
    "extra_backdrops": ["string"],
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
  ]
}
```

### Error responses

```json
{ "error": { "code": "NOT_FOUND", "message": "no tracker source found for this movie" } }
```
```json
{ "error": { "code": "INTERNAL_ERROR", "message": "failed to load tracker source" } }
```

---

## GET /movies/{id}/comments

Returns comments posted on a movie, ordered by most recent first.

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
      "user_id": 2,
      "movie_id": "string",
      "content": "string",
      "updated_at": "2026-05-06T12:00:00Z"
    }
  ]
}
```

### Error responses

```json
{ "error": { "code": "NOT_FOUND", "message": "no comments" } }
```
```json
{ "error": { "code": "INTERNAL_ERROR", "message": "failed to acess comments" } }
```

---

## POST /movies/{id}/comments

Posts a new comment on a movie.

### Path parameters

| Parameter | Type   | Description          |
|-----------|--------|----------------------|
| `id`      | string | IMDb ID of the movie |

### Request body

```json
{
  "user_id": 1,
  "movie_id": "string",
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
    "updated_at": "2026-05-06T12:00:00Z"
  }
}
```

### Error responses

```json
{ "error": { "code": "VALIDATION_ERROR", "message": "invalid request body" } }
```
```json
{ "error": { "code": "INTERNAL_ERROR", "message": "failed to create comment" } }
```

---

## GET /comments

Returns all comments across all movies.

### Response

```json
{
  "data": [
    {
      "id": 1,
      "user_id": 2,
      "movie_id": "string",
      "content": "string",
      "updated_at": "2026-05-06T12:00:00Z"
    }
  ]
}
```

### Error responses

```json
{ "error": { "code": "NOT_FOUND", "message": "comments not found" } }
```
```json
{ "error": { "code": "INTERNAL_ERROR", "message": "failed to load comments" } }
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
    "updated_at": "2026-05-06T12:00:00Z"
  }
}
```

### Error responses

```json
{ "error": { "code": "NOT_FOUND", "message": "comment not found" } }
```
```json
{ "error": { "code": "INTERNAL_ERROR", "message": "failed to load comment" } }
```

---

## PATCH /comments/{id}

Updates the content of an existing comment.

### Path parameters

| Parameter | Type   | Description       |
|-----------|--------|-------------------|
| `id`      | string | ID of the comment |

### Request body

```json
{
  "user_id": 2,
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
    "updated_at": "2026-05-06T12:00:00Z"
  }
}
```

### Error responses

```json
{ "error": { "code": "VALIDATION_ERROR", "message": "invalid request body" } }
```
```json
{ "error": { "code": "INTERNAL_ERROR", "message": "failed to update comment" } }
```

---

## DELETE /comments/{id}

Deletes a comment.

### Path parameters

| Parameter | Type   | Description       |
|-----------|--------|-------------------|
| `id`      | string | ID of the comment |

### Request body

```json
{
  "user_id": 2
}
```

### Response

`200 OK`

```json
{ "data": null }
```

### Error responses

```json
{ "error": { "code": "VALIDATION_ERROR", "message": "invalid request body" } }
```
```json
{ "error": { "code": "INTERNAL_ERROR", "message": "failed to update comment" } }
```
