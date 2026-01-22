-- PostgreSQL test schema

DROP TABLE IF EXISTS posts CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS comments CASCADE;

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    age INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    published BOOLEAN DEFAULT FALSE,
    published_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE comments (
    id SERIAL PRIMARY KEY,
    post_id INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_posts_user_id ON posts(user_id);
CREATE INDEX idx_posts_published ON posts(published);
CREATE INDEX idx_comments_post_id ON comments(post_id);
CREATE INDEX idx_comments_user_id ON comments(user_id);

INSERT INTO users (email, name, age) VALUES
    ('alice@example.com', 'Alice Johnson', 28),
    ('bob@example.com', 'Bob Smith', 35),
    ('charlie@example.com', 'Charlie Brown', 42),
    ('diana@example.com', 'Diana Prince', 31),
    ('eve@example.com', 'Eve Wilson', 26);

INSERT INTO posts (user_id, title, content, published, published_at) VALUES
    (1, 'Getting Started with Go', 'Go is a powerful language...', TRUE, NOW()),
    (1, 'Advanced Go Patterns', 'In this post we explore...', TRUE, NOW()),
    (2, 'Web Development Tips', 'Best practices for web dev...', FALSE, NULL),
    (3, 'Database Design', 'How to design efficient schemas...', TRUE, NOW()),
    (4, 'Testing in Go', 'Unit testing is important...', TRUE, NOW());

INSERT INTO comments (post_id, user_id, content) VALUES
    (1, 2, 'Great introduction!'),
    (1, 3, 'Very helpful, thanks!'),
    (2, 4, 'Could you elaborate on interfaces?'),
    (3, 1, 'Looking forward to reading this'),
    (4, 5, 'Best post I have read'),
    (5, 2, 'Excellent coverage of testing'),
    (5, 3, 'Any recommendations for test frameworks?');
