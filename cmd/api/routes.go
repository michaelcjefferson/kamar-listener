package main

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

//go:embed ui/*
var embeddedFiles embed.FS

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// navigate file system inside of ui directory, so that requests tags match file paths
	f, err := fs.Sub(embeddedFiles, "ui")
	if err != nil {
		panic(err)
	}

	// create asset handler for embedded files
	assetHandler := http.FileServer(http.FS(f))

	router.HandlerFunc(http.MethodGet, "/healthcheck", app.healthcheckHandler)

	// Wrap the /kamar-refresh handler in the authenticate middleware, to force an auth check on any request to this endpoint.
	router.HandlerFunc(http.MethodPost, "/kamar-refresh", app.authenticate(app.kamarRefreshHandler))

	// serve assets from /ui (httprouter's wild cards require a variable to be created, in this case filepath, to represent the value of the wildcard)
	router.HandlerFunc(http.MethodGet, "/assets/*filepath", func(w http.ResponseWriter, r *http.Request) {
		assetHandler.ServeHTTP(w, r)
	})
	// serve index.html from /ui
	router.Handler(http.MethodGet, "/", assetHandler)

	return app.recoverPanic(app.rateLimit(app.processCORS(router)))
}
