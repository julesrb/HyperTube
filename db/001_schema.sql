CREATE TABLE IF NOT EXISTS movies (
    imdbid          TEXT        PRIMARY KEY,
    tmdbid          TEXT        NOT NULL,
    title           TEXT        NOT NULL,
    year            TEXT        NOT NULL,
    poster_url      TEXT        NOT NULL,
    backdrop_url    TEXT        NOT NULL,
    note            REAL        NOT NULL,
    genre           INTEGER[]   NOT NULL,
    runtime_minutes INTEGER     NOT NULL,
    summary         TEXT        NOT NULL,
    director        TEXT        NOT NULL,
    "cast"          TEXT[]      NOT NULL
);

CREATE TABLE IF NOT EXISTS torrents (
    id          SERIAL PRIMARY KEY,
    imdbid      TEXT    NOT NULL REFERENCES movies(imdbid) ON DELETE CASCADE,
    source      TEXT    NOT NULL,
    url         TEXT    NOT NULL,
    quality     TEXT    NOT NULL,
    size        FLOAT   NOT NULL,
    language    TEXT    NOT NULL,
    seeds       INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS featured_movies (
    imdbid       TEXT    NOT NULL REFERENCES movies(imdbid) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    PRIMARY KEY (imdbid)
);

CREATE TABLE IF NOT EXISTS users (
    id       SERIAL PRIMARY KEY,
    username TEXT   NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS watch_history (
    id          SERIAL PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    movie_id    TEXT    NOT NULL REFERENCES movies(imdbid) ON DELETE CASCADE,
    watched_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
