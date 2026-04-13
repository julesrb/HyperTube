doc for the API

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

---

## Planned endpoints

For non-logged-in users:
- `GET /movies` — list of movies already in the DB and processed by the hypertube system

For logged-in users:
- `GET /movies` — same as above
- `GET /movies/continue-watching` — movies related to the user's watch history
- `GET /movies/popular` — paginated popularity list, extracted from torrent sources, augmented by TMDB info