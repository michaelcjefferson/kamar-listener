package main

import (
	"context"
	"database/sql"
	"log"
	"time"
)

func openDB(dbpath string) (*sql.DB, bool, error) {
	// Either connect to or create (if it doesn't exist) the database at the provided path
	db, err := sql.Open("sqlite", dbpath)
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

	// Set up users table
	err = createUserTable(db)
	if err != nil {
		db.Close()
		return nil, false, err
	}

	// Check to see whether a user already exists in the database - if not, a user must be created before the admin dashboard can be used
	exists, err := userExists(db)
	if err != nil {
		log.Fatal(err)
	}

	// Set up tables for data to be consumed from SMS
	err = createSMSTables(db)
	if err != nil {
		db.Close()
		return nil, false, err
	}

	return db, exists, nil
}

func createUserTable(db *sql.DB) error {
	userTableStmt := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_at TEXT NOT NULL DEFAULT (datetime('now')),
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL
	)`

	_, err := db.Exec(userTableStmt)

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
