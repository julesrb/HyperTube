CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS movies (
    imdbid          TEXT        PRIMARY KEY,
    tmdbid          TEXT        NOT NULL,
    title           TEXT        NOT NULL,
    year            TEXT        NOT NULL,
    poster_url      TEXT        NOT NULL,
    backdrop_url    TEXT        NOT NULL,
    note            REAL        NOT NULL,
    genre           INTEGER[]   NOT NULL,
    summary         TEXT        NOT NULL
);

CREATE TABLE IF NOT EXISTS torrents (
    id          SERIAL PRIMARY KEY,
    imdbid      TEXT    NOT NULL REFERENCES movies(imdbid) ON DELETE CASCADE,
    title       TEXT    NOT NULL,
    year        INTEGER NOT NULL,
    source      TEXT    NOT NULL,
    url         TEXT    NOT NULL,
    quality     TEXT    NOT NULL,
    size        FLOAT    NOT NULL,
    language    TEXT    NOT NULL,
    seeds       TEXT    NOT NULL,
    UNIQUE (imdbid, url)
);

CREATE TABLE IF NOT EXISTS featured_movies (
    imdbid       TEXT    NOT NULL REFERENCES movies(imdbid) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    UNIQUE (position, imdbid),
    PRIMARY KEY (imdbid)
);

CREATE TABLE IF NOT EXISTS users (
    id            SERIAL PRIMARY KEY,
    email         CITEXT      NOT NULL UNIQUE,
    username      TEXT        NOT NULL UNIQUE,
    first_name    TEXT        NOT NULL,
    last_name     TEXT        NOT NULL,
    password_hash TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS oauth_accounts (
    id               SERIAL PRIMARY KEY,
    user_id          INTEGER     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider         TEXT        NOT NULL,
    provider_user_id TEXT        NOT NULL,
    provider_email   CITEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (provider, provider_user_id),
    UNIQUE (user_id, provider)
);

CREATE TABLE IF NOT EXISTS watch_history (
    id          SERIAL PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    imdbid      TEXT    NOT NULL REFERENCES movies(imdbid) ON DELETE CASCADE,
    watched_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS direct_stream_movies (
    imdbid      TEXT    NOT NULL REFERENCES movies(imdbid) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (imdbid)
);

CREATE TABLE IF NOT EXISTS comments (
    id          SERIAL  PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    movie_id    TEXT    NOT NULL REFERENCES movies(imdbid) ON DELETE CASCADE,
    content     TEXT    NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS movie_searches (
    query       TEXT        NOT NULL,
    imdbid      TEXT        NOT NULL REFERENCES movies(imdbid) ON DELETE CASCADE,
    rank        INTEGER     NOT NULL,
    searched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (query, imdbid)
)