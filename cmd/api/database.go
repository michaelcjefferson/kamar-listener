package main

import (
	"context"
	"database/sql"
	"time"
)

func openDB(dbpath string) (*sql.DB, error) {
	// Either connect to or create (if it doesn't exist) the database at the provided path
	db, err := sql.Open("sqlite3", dbpath)
	if err != nil {
		return nil, err
	}

	// Create context with 5 second deadline so that we can ping the db and finish establishing a db connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

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

	// Includes resultData and results fields
	tableStmt := `CREATE TABLE IF NOT EXISTS results (
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
	_, err = db.Exec(tableStmt)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
