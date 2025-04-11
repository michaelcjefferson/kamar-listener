package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mjefferson-whs/listener/internal/data"
	"github.com/mjefferson-whs/listener/internal/jsonlog"

	// _ "modernc.org/sqlite"
	_ "github.com/mattn/go-sqlite3"
)

type config struct {
	port               int
	env                string
	dblogs_on          bool
	app_db_path        string
	kamar_data_db_path string
	https_on           bool
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
	tls struct {
		certPath string
		keyPath  string
	}
	tokens struct {
		expiry  time.Duration
		refresh time.Duration
	}
}

type application struct {
	assetHandler http.Handler
	config       config
	// Allows processes, eg. token deletion cycle, to respond to this channel closing (and eg. perform tidy up operations)
	isShuttingDown chan struct{}
	logger         *jsonlog.Logger
	models         data.Models
	userExists     bool
	wg             sync.WaitGroup
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 8085, "API server port.")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production).")

	flag.StringVar(&cfg.app_db_path, "app-db-path", "./db/app.db", "Path to SQLite .db file holding user data, config, logs etc.")
	flag.StringVar(&cfg.kamar_data_db_path, "kamar-data-db-path", "./db/listener.db", "Path to SQLite .db file holding data from KAMAR directory service.")

	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 4, "Rate limiter maximum requests per second.")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 8, "Rate limiter maximum burst.")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", false, "Enable rate limiter.")

	// TODO: Remove once set up in database config OR process these into DB, and prompt client to update with their own
	flag.StringVar(&cfg.credentials.username, "username", "username", "For authentication from KAMAR.")
	flag.StringVar(&cfg.credentials.password, "password", "password", "For authentication from KAMAR.")

	flag.BoolVar(&cfg.https_on, "https_on", true, "Turn server-side HTTPS on or off.")
	flag.BoolVar(&cfg.dblogs_on, "dblogs_on", true, "Turn writing logs to database on or off.")

	flag.Parse()

	cfg.credentials.full = strings.Join([]string{cfg.credentials.username, cfg.credentials.password}, ":")

	cfg.tls.certPath = "./tls/cert.pem"
	cfg.tls.keyPath = "./tls/key.pem"

	cfg.tokens.expiry = 24 * time.Hour
	cfg.tokens.refresh = 6 * time.Hour

	app := &application{
		config:         cfg,
		isShuttingDown: make(chan struct{}),
	}

	fmt.Println("attempting to set up SQLite db")

	appdb, userExists, err := openAppDB(cfg.app_db_path)
	if err != nil {
		fmt.Printf("error setting up app database: %v\n", err)
		time.Sleep(20 * time.Second)
	}

	defer appdb.Close()
	app.userExists = userExists

	kamardb, err := openKamarDB(cfg.kamar_data_db_path)
	if err != nil {
		fmt.Printf("error setting up app database: %v\n", err)
		time.Sleep(20 * time.Second)
	}

	defer appdb.Close()
	app.userExists = userExists

	models := data.NewModels(appdb, kamardb, app.background)
	app.models = models

	// Instantiate logger that will log anything at or above info level. To write from a different level, change this parameter.
	var logger *jsonlog.Logger
	if cfg.dblogs_on {
		logger = jsonlog.New(os.Stdout, jsonlog.LevelInfo, &models.Logs)
	} else {
		// If cfg.dblogs_on is false, logger will only write to stdout
		logger = jsonlog.New(os.Stdout, jsonlog.LevelInfo, nil)
	}
	app.logger = logger
	app.logger.PrintInfo("database connection established", nil)

	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
		time.Sleep(20 * time.Second)
	}
}
