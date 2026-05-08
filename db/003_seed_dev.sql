-- Dev seed: users and comments for local development
INSERT INTO users (username) VALUES
    ('alice'),
    ('bob'),
    ('charlie'),
    ('diana')
ON CONFLICT (username) DO NOTHING;
