package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mjefferson-whs/listener/internal/data"
	"github.com/mjefferson-whs/listener/internal/jsonlog"

	_ "github.com/mattn/go-sqlite3"
)

type config struct {
	port     int
	env      string
	dbpath   string
	https_on bool
	// rps (requests per second) must be float, burst must be int for limiter. enabled allows turning off the rate limiter for, for example load testing.
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	credentials struct {
		username string
		password string
		full     string
	}
}

type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	wg     sync.WaitGroup
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 443, "API server port.")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production).")

	flag.StringVar(&cfg.dbpath, "db-path", "./kamar-directory-service.db", "Path to SQLite .db (database) file.")

	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second.")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst.")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", false, "Enable rate limiter.")

	flag.StringVar(&cfg.credentials.username, "username", "username", "For authentication from KAMAR.")
	flag.StringVar(&cfg.credentials.password, "password", "password", "For authentication from KAMAR.")

	flag.BoolVar(&cfg.https_on, "https_on", true, "Turn server-side HTTPS on or off.")

	flag.Parse()

	cfg.credentials.full = strings.Join([]string{cfg.credentials.username, cfg.credentials.password}, ":")

	// Instantiate logger that will log anything at or above info level. To write from a different level, change this parameter.
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	logger.PrintInfo("attempting to set up SQLite db", nil)

	db, err := openDB(cfg.dbpath)
	if err != nil {
		logger.PrintFatal(err, nil)
		time.Sleep(20 * time.Second)
	}

	defer db.Close()

	logger.PrintInfo("database connection established", nil)

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
		time.Sleep(20 * time.Second)
	}
}

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
		resultData,
		results,
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
