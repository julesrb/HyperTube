INSERT INTO featured_movies (movie_id, position) VALUES
('tt1757678',  1),  -- Avatar: Fire and Ash
('tt12042730', 2),  -- Project Hail Mary
('tt15239678', 3),  -- Dune: Part Two
('tt28650488', 4),  -- The Super Mario Galaxy Movie
('tt33071426', 5),  -- The Drama
('tt26443616', 6),  -- Hoppers
('tt30851137', 7),  -- The Bride!
('tt8036976',  8),  -- Send Help
('tt27613895', 9)   -- GOAT
ON CONFLICT (movie_id) DO NOTHING;
