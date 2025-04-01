CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  last_authenticated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  username TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL
);

-- test_user_1 pass: password
-- geoff-the-test pass: fish6182*five
INSERT INTO users (username, password_hash) VALUES 
  ('test_user_1', '$2a$12$WUwQdlE7K7pVhLifm4xgBe7p6MuyJBiG.0.P0FxYLrYfZqCJqjGGO'),
  ('geoff-the-test', '$2a$12$uqf0L5oWsuVrKf9S1Rrf0uVTj6wbSEXfC01obOXSJb77glJhvbghS');

CREATE TABLE IF NOT EXISTS tokens (
  hash BLOB PRIMARY KEY, 
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE, 
  expiry TEXT NOT NULL
);