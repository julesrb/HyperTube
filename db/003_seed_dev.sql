-- Dev seed: users and comments for local development
INSERT INTO users (username) VALUES
    ('alice'),
    ('bob'),
    ('charlie'),
    ('diana')
ON CONFLICT (username) DO NOTHING;

-- INSERT INTO comments (user_id, movie_id, content, updated_at)
-- SELECT u.id, m.imdbid, c.content, c.updated_at
-- FROM (VALUES
--     ('alice',   'tt26443616', 'Loved the animation style, very refreshing!',              NOW() - INTERVAL '2 days'),
--     ('bob',     'tt26443616', 'Great movie for kids and adults alike.',                   NOW() - INTERVAL '1 day'),
--     ('charlie', 'tt26443616', 'The story felt a bit rushed in the second half.',          NOW() - INTERVAL '3 hours'),
--     ('diana',   'tt26443616', 'Visually stunning, would watch again.',                    NOW()),
--     ('alice',   'tt28650488', 'Not as good as the first Mario movie but still fun.',      NOW() - INTERVAL '5 days'),
--     ('bob',     'tt28650488', 'My kids absolutely loved it!',                             NOW() - INTERVAL '4 days'),
--     ('charlie', 'tt10078772', 'Kept me on the edge of my seat the whole time.',           NOW() - INTERVAL '1 day'),
--     ('diana',   'tt10078772', 'Predictable plot but great performances.',                 NOW() - INTERVAL '12 hours'),
--     ('alice',   'tt10078772', 'The third act twist genuinely surprised me.',              NOW() - INTERVAL '2 hours'),
--     ('bob',     'tt42373968', 'Weird but in a good way.',                                 NOW() - INTERVAL '6 days'),
--     ('charlie', 'tt42373968', 'Not sure what I just watched but I enjoyed it.',           NOW() - INTERVAL '3 days'),
--     ('diana',   'tt34624872', 'Finally a sequel that does justice to the original.',      NOW() - INTERVAL '1 day')
-- ) AS c(username, movie_id, content, updated_at)
-- JOIN users u ON u.username = c.username
-- JOIN movies m ON m.imdbid = c.movie_id;
