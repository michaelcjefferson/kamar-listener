CREATE TABLE IF NOT EXISTS logs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  level TEXT NOT NULL,
  time TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  message TEXT NOT NULL,
  properties TEXT,
  trace TEXT
);

ALTER TABLE logs ADD COLUMN user_id INTEGER
GENERATED ALWAYS AS (
  CASE 
    WHEN json_valid(properties) 
      AND json_type(json_extract(properties, '$.user_id')) = 'integer' 
      AND json_extract(properties, '$.user_id') != '' 
    THEN json_extract(properties, '$.user_id') 
    ELSE NULL 
  END
) VIRTUAL;

CREATE INDEX IF NOT EXISTS idx_logs_user_id ON logs(user_id);

CREATE VIRTUAL TABLE IF NOT EXISTS logs_fts
  USING fts5(message, content='logs', content_rowid='id');

CREATE TABLE IF NOT EXISTS logs_metadata (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  type TEXT NOT NULL,
  level TEXT UNIQUE,
  user_id INTEGER UNIQUE,
  count INTEGER NOT NULL DEFAULT 0
);

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
  ('test_user_1', '$2a$12$WUwQdlE7K7pVhLifm4xgBe7p6MuyJBiG.0.P0FxYLrYfZqCJqjGGO')
  ('geoff-the-test', '$2a$12$uqf0L5oWsuVrKf9S1Rrf0uVTj6wbSEXfC01obOXSJb77glJhvbghS')

CREATE TABLE IF NOT EXISTS tokens (
  hash BLOB PRIMARY KEY, 
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE, 
  expiry TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS config (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL,
  type TEXT NOT NULL,
  description TEXT NOT NULL,
  updated_at TEXT DEFAULT CURRENT_TIMESTAMP
);

INSERT OR IGNORE INTO config (key, value, type, description) VALUES 
  ('service_name', 'KAMAR Listener Service', 'string', 'Use the acronym/name of your school, eg. "WHS KAMAR Listener Service"'),
  ('info_url', 'https://www.educationcounts.govt.nz/directories/list-of-nz-schools', 'string', 'Website where people can contact you/read about how you use this service, eg. https://schoolname.school.nz'),
  ('privacy_statement', 'This service only collects results data, and stores it locally on a secure device. Only staff members of the school have access to the data.', 'string', 'Minimum 100 characters: a description of how you use the data from this listener service'),
  ('listener_username', 'username', 'string', 'Username entered into KAMAR when setting up listener service'),
  ('listener_password', '', 'password', 'Password entered into KAMAR when setting up listener service'),
  ('details', 'true', 'bool', 'Enable/disable details'),
  ('passwords', 'true', 'bool', 'Enable/disable passwords'),
  ('photos', 'true', 'bool', 'Enable/disable photos'),
  ('groups', 'true', 'bool', 'Enable/disable groups'),
  ('awards', 'true', 'bool', 'Enable/disable awards'),
  ('timetables', 'true', 'bool', 'Enable/disable timetables'),
  ('attendance', 'false', 'bool', 'Enable/disable attendance'),
  ('assessments', 'true', 'bool', 'Enable/disable results and assessments'),
  ('pastoral', 'true', 'bool', 'Enable/disable pastoral'),
  ('learning_support', 'true', 'bool', 'Enable/disable learning support'),
  ('subjects', 'true', 'bool', 'Enable/disable subjects'),
  ('notices', 'true', 'bool', 'Enable/disable notices'),
  ('calendar', 'false', 'bool', 'Enable/disable calendar'),
  ('bookings', 'true', 'bool', 'Enable/disable bookings');

CREATE TABLE IF NOT EXISTS results (
  code			TEXT,
  comment         TEXT,
  course          TEXT,
  curriculumlevel,
  date            TEXT,
  enrolled		INTEGER,
  id              INTEGER,
  nsn             TEXT,
  number          TEXT,
  published		INTEGER,
  result          TEXT,
  resultData TEXT,
  results TEXT,
  subject         TEXT,
  tnv 			TEXT,
  type            TEXT,
  version         INTEGER,
  year            INTEGER,
  yearlevel       INTEGER
);