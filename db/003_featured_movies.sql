CREATE TABLE IF NOT EXISTS featured_movies (
    movie_id TEXT    NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    PRIMARY KEY (movie_id)
);
