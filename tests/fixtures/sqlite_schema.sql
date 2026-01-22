-- SQLite test schema

DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS posts;
DROP TABLE IF EXISTS users;

CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    age INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    content TEXT,
    published BOOLEAN DEFAULT 0,
    published_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
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
    (1, 'Getting Started with Go', 'Go is a powerful language...', 1, datetime('now')),
    (1, 'Advanced Go Patterns', 'In this post we explore...', 1, datetime('now')),
    (2, 'Web Development Tips', 'Best practices for web dev...', 0, NULL),
    (3, 'Database Design', 'How to design efficient schemas...', 1, datetime('now')),
    (4, 'Testing in Go', 'Unit testing is important...', 1, datetime('now'));

INSERT INTO comments (post_id, user_id, content) VALUES
    (1, 2, 'Great introduction!'),
    (1, 3, 'Very helpful, thanks!'),
    (2, 4, 'Could you elaborate on interfaces?'),
    (3, 1, 'Looking forward to reading this'),
    (4, 5, 'Best post I have read'),
    (5, 2, 'Excellent coverage of testing'),
    (5, 3, 'Any recommendations for test frameworks?');
