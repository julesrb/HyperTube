CREATE TABLE IF NOT EXISTS movies (
    imdbid          TEXT PRIMARY KEY,
    tmdbid          TEXT        NOT NULL,
    title           TEXT        NOT NULL,
    year            TEXT        NOT NULL,
    poster_url      TEXT        NOT NULL,
    backdrop_url    TEXT        NOT NULL,
    imdb_rating     REAL        NOT NULL,
    genres          TEXT[]      NOT NULL,
    runtime_minutes INTEGER     NOT NULL,
    summary         TEXT        NOT NULL,
    director        TEXT        NOT NULL,
    "cast"          TEXT[]      NOT NULL, 
    -- watched         BOOLEAN     NOT NULL DEFAULT FALSE,
    -- progression     FLOAT       NOT NULL DEFAULT 0,
    seeders         INTEGER     NOT NULL DEFAULT 0
);
