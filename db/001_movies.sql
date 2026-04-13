CREATE TABLE IF NOT EXISTS movies (
    id              TEXT PRIMARY KEY,
    title           TEXT        NOT NULL,
    year            INTEGER,
    poster_url      TEXT,
    backdrop_url    TEXT,
    imdb_rating     REAL,
    genres          TEXT[],
    runtime_minutes INTEGER,
    summary         TEXT,
    director        TEXT,
    "cast"          TEXT[],
    watched         BOOLEAN     NOT NULL DEFAULT FALSE,
    seeders         INTEGER     NOT NULL DEFAULT 0
);
