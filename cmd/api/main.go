package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mjefferson-whs/listener/internal/data"
	"github.com/mjefferson-whs/listener/internal/getip"
	"github.com/mjefferson-whs/listener/internal/jsonlog"
	"github.com/mjefferson-whs/listener/internal/setfiledirs"
	"github.com/mjefferson-whs/listener/internal/sslcerts"

	// _ "modernc.org/sqlite"
	_ "github.com/mattn/go-sqlite3"
)

type config struct {
	ip                   string
	port                 int
	env                  string
	dblogs_on            bool
	kamar_db_table_names []string
	https_on             bool
	basePath             string
	dbPaths              struct {
		appDB      string
		dbDir      string
		listenerDB string
	}
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
	tlsPaths struct {
		cert   string
		key    string
		tlsDir string
	}
	tokens struct {
		expiry  time.Duration
		refresh time.Duration
	}
}

type application struct {
	appMetrics   appMetrics
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
	flag.StringVar(&cfg.env, "env", "production", "Environment (development|staging|production).")

	flag.StringVar(&cfg.dbPaths.dbDir, "db-dir-path", "./db", "Path to directory holding sqlite db files")

	flag.StringVar(&cfg.dbPaths.appDB, "app-db-path", "./db/app.db", "Path to SQLite .db file holding user data, config, logs etc.")
	flag.StringVar(&cfg.dbPaths.listenerDB, "listener-db-path", "./db/listener.db", "Path to SQLite .db file holding data received from KAMAR directory service.")

	flag.StringVar(&cfg.tlsPaths.tlsDir, "tls-dir-path", "./tls", "Path to directory holding tls files")

	flag.StringVar(&cfg.tlsPaths.cert, "cert-path", "./tls/cert.pem", "Path to cert.pem TLS file.")
	flag.StringVar(&cfg.tlsPaths.key, "key-path", "./tls/key.pem", "Path to key.pem TLS file.")

	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 4, "Rate limiter maximum requests per second.")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 8, "Rate limiter maximum burst.")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", false, "Enable rate limiter.")

	// TODO: Remove once set up in database config OR process these into DB, and prompt client to update with their own
	flag.StringVar(&cfg.credentials.username, "username", "username", "For authentication from KAMAR.")
	flag.StringVar(&cfg.credentials.password, "password", "password", "For authentication from KAMAR.")

	flag.BoolVar(&cfg.https_on, "https_on", true, "Turn server-side HTTPS on or off.")
	flag.BoolVar(&cfg.dblogs_on, "dblogs_on", true, "Turn writing logs to database on or off.")

	flag.Parse()

	// TODO: Likely unnecessary, as TLS and DB paths are already set in writable static folders
	// if cfg.env == "production" {
	// 	checkrundir.EnforceRunLocation()
	// }

	dirs, err := setfiledirs.SetFileDirs("kamar-listener", []string{"db", "tls"})
	if err != nil {
		log.Fatalf("couldn't set up app data directories: %v", err)
	}

	// TODO: Move to helper function
	cfg.basePath = dirs.ApplicationDir

	cfg.dbPaths.dbDir = dirs.FileDirs["db"]
	cfg.dbPaths.appDB = filepath.Join(cfg.dbPaths.dbDir, "app.db")
	cfg.dbPaths.listenerDB = filepath.Join(cfg.dbPaths.dbDir, "listener.db")

	cfg.tlsPaths.tlsDir = dirs.FileDirs["tls"]
	cfg.tlsPaths.cert = filepath.Join(cfg.tlsPaths.tlsDir, "cert.pem")
	cfg.tlsPaths.key = filepath.Join(cfg.tlsPaths.tlsDir, "key.pem")

	cfg.credentials.full = strings.Join([]string{cfg.credentials.username, cfg.credentials.password}, ":")

	cfg.tokens.expiry = 24 * time.Hour
	cfg.tokens.refresh = 6 * time.Hour

	app := &application{
		config:         cfg,
		isShuttingDown: make(chan struct{}),
	}

	fmt.Println("attempting to set up SQLite db")

	appDB, userExists, err := openAppDB(cfg.dbPaths.appDB)
	if err != nil {
		fmt.Printf("error setting up app database: %v\n", err)
		time.Sleep(20 * time.Second)
	}

	defer appDB.Close()
	app.userExists = userExists

	listenerDB, err := openKamarDB(cfg.dbPaths.listenerDB)
	if err != nil {
		fmt.Printf("error setting up listener database: %v\n", err)
		time.Sleep(20 * time.Second)
	}

	defer listenerDB.Close()

	models := data.NewModels(appDB, listenerDB, app.background)
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

	err = app.UpdateRecordCountsFromDB()
	if err != nil {
		app.logger.PrintFatal(err, nil)
	}

	ip, err := getip.GetLocalIP()
	if err != nil {
		app.logger.PrintError(err, nil)
	}
	// TODO: Add local IP field in config db. Check if previous local IP exists, and if it does and it's different to the one returned by GetLocalIP, generate new SSL certs

	app.config.ip = ip
	app.logger.PrintInfo("ip address set", map[string]any{
		"ip_address": app.config.ip,
	})

	// if cfg.env == "production" {
	// 	err = sslcerts.GenerateSSLCert(app.logger)
	// 	if err != nil {
	// 		app.logger.PrintFatal(err, nil)
	// 	}
	// }

	err = sslcerts.GenerateSSLCert(app.config.tlsPaths.tlsDir, ip, app.logger)
	// If SSL certs couldn't be found or generated, serve over HTTP; otherwise, serve over HTTPS
	if err != nil {
		app.logger.PrintError(err, nil)
	}

	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
		time.Sleep(20 * time.Second)
	}
}
