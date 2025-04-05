package data

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// TODO: Consider using an in-memory db like this one for db tests that create a new db for each test - look at integrating db connection string params
func newTestDBInMemory(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	return db
}

// TODO: Consider using a temp file for persistent dbs used in tests - add use of setup script
func newTestDBTempFile(t *testing.T) *sql.DB {
	// The * is filled with a random string
	tmpFile, err := os.CreateTemp("", "test-db-temp-*.db")
	if err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("sqlite3", tmpFile.Name()+"?_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		db.Close()
		os.Remove(tmpFile.Name())
	})

	return db
}

// Create new sqlite database with mock data for testing user model - database is removed once tests have been run. Only contains user and token tables
func newTestUserDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", "./testdata/test-user.db?_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL")
	if err != nil {
		t.Fatal(err)
	}

	script, err := os.ReadFile("./testdata/user-db-setup.sql")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(string(script))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		db.Close()
		os.Remove("./testdata/test-user.db")
	})

	return db
}

// Create new full sqlite database with mock data in all tables for testing - database is removed once tests have been run
func newTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", "./testdata/test-kamar-directory-service.db?_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL")
	if err != nil {
		t.Fatal(err)
	}

	script, err := os.ReadFile("./testdata/full-db-setup.sql")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(string(script))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		db.Close()
		os.Remove("./testdata/test-kamar-directory-service.db")
	})

	return db
}
