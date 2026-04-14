doc for the API

---

## GET /movies

Returns a selected list of movies available in the system.

### Response

```json
{
  "data": [
    {
      "id": "string",
      "title": "string",
      "poster_url": "string",
      "backdrop_url": "string"
    }
  ],
  "meta": {
    "total": 11,
    "page": 1,
    "per_page": 11
  }
}
```

> Note: only `id`, `title`, `poster_url`, and `backdrop_url` are returned by this endpoint. Full movie details (rating, genres, cast, etc.) are available via `GET /movies/:id` (not yet implemented).

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

## GET /movies

Returns a list of movies available in the system.

### Response

```json
{
  "data": [
    {
      "id": "string",
      "title": "string",
      "poster_url": "string",
      "backdrop_url": "string"
    }
  ],
  "meta": {
    "total": 11,
    "page": 1,
    "per_page": 11
  }
}
```

> Note: only `id`, `title`, `poster_url`, and `backdrop_url` are returned by this endpoint. Full movie details (rating, genres, cast, etc.) are available via `GET /movies/:id` (not yet implemented).

### Error response

```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "failed to load movies"
  }
}
```