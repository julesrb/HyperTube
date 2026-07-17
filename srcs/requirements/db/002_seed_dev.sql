-- Dev seed: users and comments for local development

INSERT INTO users (email, username, first_name, last_name, password_hash, profile_picture) VALUES
    ('dev@hypertube.local', 'dev_user', 'Dev', 'User', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', NULL),
    ('alice@example.local', 'alice', 'Alice', 'Example', '', NULL),
    ('bob@example.local', 'bob', 'Bob', 'Example', '', NULL),
    ('charlie@example.local', 'charlie', 'Charlie', 'Example', '', NULL),
    ('diana@example.local', 'diana', 'Diana', 'Example', '', NULL);

