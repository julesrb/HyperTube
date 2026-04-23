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

> Note: only `id`, `title`, `poster_url`, and `backdrop_url` are returned by this endpoint. Full movie details are available via `GET /movies/:id`.

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

Returns full metadata for a single movie already ingested in the system.

### Path parameters

| Parameter | Type   | Description      |
|-----------|--------|------------------|
| `id`      | string | The movie ID     |

### Response

```json
{
  "data": {
    "id": "string",
    "title": "string",
    "year": 2024,
    "poster_url": "string",
    "backdrop_url": "string",
    "imdb_rating": 8.1,
    "genres": ["Sci-Fi", "Adventure"],
    "runtime_minutes": 167,
    "summary": "string",
    "director": "string",
    "cast": ["string"],
    "watched": false,
    "seeders": 1200
  }
}
```

### Error responses

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "movie not found"
  }
}
```

```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "failed to load movie"
  }
}
```