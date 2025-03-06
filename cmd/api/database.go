package main

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"time"
)

// SQLite config for reads and writes (avoid SQLITE BUSY error): https://kerkour.com/sqlite-for-servers
func openDB(dbpath string) (*sql.DB, bool, error) {
	// Either connect to or create (if it doesn't exist) the database at the provided path
	db, err := sql.Open("sqlite3", dbpath)
	if err != nil {
		return nil, false, err
	}

	// Create context with 5 second deadline so that we can ping the db and finish establishing a db connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, false, err
	}

	// Set up logs table
	err = createLogsTable(db)
	if err != nil {
		db.Close()
		return nil, false, err
	}

	// Set up users table
	err = createUserTable(db)
	if err != nil {
		db.Close()
		return nil, false, err
	}

	// Set up tokens table
	err = createTokenTable(db)
	if err != nil {
		db.Close()
		return nil, false, err
	}

	// Set up tables for data to be consumed from SMS
	err = createSMSTables(db)
	if err != nil {
		db.Close()
		return nil, false, err
	}

	// Check to see whether a user already exists in the database - if not, a user must be created before the admin dashboard can be used
	exists, err := userExists(db)
	if err != nil {
		log.Fatal(err)
	}

	return db, exists, nil
}

func createLogsTable(db *sql.DB) error {
	userTableStmt := `CREATE TABLE IF NOT EXISTS logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		level TEXT NOT NULL,
		time TEXT NOT NULL DEFAULT (datetime('now')),
		message TEXT NOT NULL,
		properties TEXT,
		trace TEXT
	)`

	_, err := db.Exec(userTableStmt)

	if err != nil {
		return err
	}

	// Create an indexed column based on any logs that come with an attached user id, to make it easier to query for logs regarding a specific user. VIRTUAL allows the column to store NULL values without errors, and NULL values are ignored in indexes
	// Check the datatype of $.userID before extracting, as otherwise in some cases the value of this column can be set to a single " and cause errors
	// NOTE: Even with the below query, properties with a userID value of "" still cause malformed JSON errors (specifically the logs created by logsPage query params)
	alterTableStmt := `ALTER TABLE logs ADD COLUMN userID INTEGER
	GENERATED ALWAYS AS (
    CASE 
			WHEN json_valid(properties) 
				AND json_type(json_extract(properties, '$.userID')) = 'integer' 
				AND json_extract(properties, '$.userID') != '' 
			THEN json_extract(properties, '$.userID') 
			ELSE NULL 
    END
	) VIRTUAL;`

	_, err = db.Exec(alterTableStmt)
	// Alter table doesn't support IF NOT EXISTS, so ignore the error thrown if this column already exists
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		return err
	}

	createIndexStmt := `
		CREATE INDEX IF NOT EXISTS idx_logs_userID ON logs(userID);
	`

	_, err = db.Exec(createIndexStmt)

	// FTS5 tables are optimised for text search, and allow in this case for more efficiently searching for logs containing key words
	// This particular FTS5 table is only set up to allow for searching for text in log messages only, as defined by the message param
	createFTSTableStmt := `
		CREATE VIRTUAL TABLE IF NOT EXISTS logs_fts
		USING fts5(message, content='logs', content_rowid='id');
	`

	_, err = db.Exec(createFTSTableStmt)

	createLogsMetadataTableStmt := `
		CREATE TABLE IF NOT EXISTS logs_metadata (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			level TEXT,
			userID INTEGER,
			count INTEGER NOT NULL DEFAULT 0,
			UNIQUE(level, userID)
		);`

	_, err = db.Exec(createLogsMetadataTableStmt)

	return err
}

func createUserTable(db *sql.DB) error {
	userTableStmt := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_at TEXT NOT NULL DEFAULT (datetime('now')),
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL
	);`

	_, err := db.Exec(userTableStmt)

	return err
}

func createTokenTable(db *sql.DB) error {
	tokenTableStmt := `CREATE TABLE IF NOT EXISTS tokens (
		hash BLOB PRIMARY KEY, 
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE, 
		expiry TEXT NOT NULL
	);`

	_, err := db.Exec(tokenTableStmt)

	return err
}

func createSMSTables(db *sql.DB) error {
	// Includes resultData and results fields
	resultTableStmt := `CREATE TABLE IF NOT EXISTS results (
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
	)`

	// If the results table doesn't already exist in the database, create it
	// tableStmt := `CREATE TABLE IF NOT EXISTS results (
	// 	code,			TEXT,
	// 	comment         TEXT,
	// 	course          TEXT,
	// 	curriculumlevel,
	// 	date            TEXT,
	// 	enrolled,		INTEGER,
	// 	id              INTEGER,
	// 	nsn             TEXT,
	// 	number          TEXT,
	// 	published,		INTEGER,
	// 	result          TEXT,
	// 	subject         TEXT,
	// 	type            TEXT,
	// 	version         INTEGER,
	// 	year            INTEGER,
	// 	yearlevel       INTEGER
	// )`

	_, err := db.Exec(resultTableStmt)

	return err
}

func userExists(db *sql.DB) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count > 0, err
}
