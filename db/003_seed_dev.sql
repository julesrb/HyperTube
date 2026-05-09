-- Dev seed: users and comments for local development
INSERT INTO users (username) VALUES
    ('alice'),
    ('bob'),
    ('charlie'),
    ('diana')
ON CONFLICT (username) DO NOTHING;

INSERT INTO movies (imdbid, tmdbid, title, year, poster_url, backdrop_url, note, genre, summary) VALUES
    ('tt1612774', '45649', 'Rubber', '2010',
     'https://image.tmdb.org/t/p/w500/mWgCORI5IC5vOvB2cDVQe0YNtXZ.jpg',
     'https://image.tmdb.org/t/p/w500/cqfaq9uczoHBBug74eORBH4Jcpe.jpg',
     5.9, ARRAY[35, 18, 14, 27, 9648],
     'A tire named Robert comes to life in the California desert and discovers it has psychic powers capable of destroying anything it encounters — for no reason at all.')
ON CONFLICT (imdbid) DO NOTHING;

INSERT INTO direct_stream_movies (imdbid) VALUES
    ('tt1612774')
ON CONFLICT (imdbid) DO NOTHING;

INSERT INTO featured_movies (imdbid, position) VALUES
    ('tt1612774', 0)
ON CONFLICT (imdbid) DO UPDATE SET position = 0;
