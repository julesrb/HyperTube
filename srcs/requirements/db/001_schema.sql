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
    original_language TEXT      NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS torrents (
    id          TEXT    PRIMARY KEY,
    imdbid      TEXT    NOT NULL REFERENCES movies(imdbid) ON DELETE CASCADE,
    title       TEXT    NOT NULL,
    year        INTEGER NOT NULL,
    source      TEXT    NOT NULL,
    url         TEXT    NOT NULL,
    hash        TEXT,
    quality     TEXT    NOT NULL,
    size        FLOAT   NOT NULL,
    language    TEXT    NOT NULL,
    seeds       TEXT    NOT NULL,
    status      TEXT    NOT NULL DEFAULT 'zero' CHECK (status IN ('zero', 'in_progress', 'failed', 'finished')),
    finished_at TIMESTAMPTZ,
    UNIQUE (url)
);

CREATE TABLE IF NOT EXISTS featured_movies (
    imdbid       TEXT    NOT NULL REFERENCES movies(imdbid) ON DELETE CASCADE PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS default_movies (
    imdbid       TEXT    NOT NULL REFERENCES movies(imdbid) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    PRIMARY KEY (imdbid)
);

CREATE TABLE IF NOT EXISTS users (
    id            SERIAL PRIMARY KEY,
    email         CITEXT      NOT NULL,
    username      TEXT        NOT NULL UNIQUE,
    first_name    TEXT        NOT NULL,
    last_name     TEXT        NOT NULL,
    profile_picture TEXT,
    password_hash TEXT,
    color         TEXT        NOT NULL DEFAULT 'purple' CHECK (color IN ('yellow', 'pink', 'green', 'purple', 'blue', 'red')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS users_password_email_key
    ON users (email)
    WHERE COALESCE(password_hash, '') <> '';

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

CREATE TABLE IF NOT EXISTS oauth_applications (
    id                 SERIAL PRIMARY KEY,
    owner_user_id      INTEGER     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name               TEXT        NOT NULL,
    scope              TEXT        NOT NULL DEFAULT '',
    redirect_uri       TEXT        NOT NULL DEFAULT '',
    client_id          TEXT        NOT NULL UNIQUE,
    client_secret_hash TEXT        NOT NULL,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS oauth_applications_owner_user_id_idx
    ON oauth_applications (owner_user_id);

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id         SERIAL      PRIMARY KEY,
    user_id    INTEGER     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT        NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS password_reset_tokens_user_id_idx
    ON password_reset_tokens (user_id);

CREATE INDEX IF NOT EXISTS password_reset_tokens_valid_idx
    ON password_reset_tokens (token_hash, expires_at)
    WHERE used_at IS NULL;


CREATE TABLE IF NOT EXISTS watch_history (
    id          SERIAL PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    imdbid      TEXT    NOT NULL REFERENCES movies(imdbid) ON DELETE CASCADE,
    progress    INTEGER NOT NULL DEFAULT 0 CHECK (progress >= 0),
    complete    BOOLEAN NOT NULL DEFAULT FALSE,
    pourcent    INTEGER NOT NULL DEFAULT 0 CHECK (pourcent >= 0 AND pourcent <= 100),
    watched_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT watch_history_user_id_imdbid_key UNIQUE (user_id, imdbid)
);

CREATE TABLE IF NOT EXISTS comments (
    id          SERIAL  PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    movie_id    TEXT    NOT NULL REFERENCES movies(imdbid) ON DELETE CASCADE,
    content     TEXT    NOT NULL,
    edited      BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS movie_searches (
    query       TEXT        NOT NULL,
    imdbid      TEXT        NOT NULL REFERENCES movies(imdbid) ON DELETE CASCADE,
    rank        INTEGER     NOT NULL,
    searched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (query, imdbid)
);

CREATE TABLE IF NOT EXISTS app_state (
    key         TEXT        PRIMARY KEY,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
