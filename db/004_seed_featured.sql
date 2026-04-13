INSERT INTO featured_movies (movie_id, position) VALUES
('83533',   1),  -- Avatar: Fire and Ash
('687163',  2),  -- Project Hail Mary
('693134',  3),  -- Dune: Part Two
('1226863', 4),  -- The Super Mario Galaxy Movie
('1325734', 5),  -- The Drama
('1327819', 6),  -- Hoppers
('1159831', 7),  -- The Bride!
('1198994', 8),  -- Send Help
('1297842', 9)   -- GOAT
ON CONFLICT (movie_id) DO NOTHING;
