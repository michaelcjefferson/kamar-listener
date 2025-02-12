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
	app.assetHandler = http.FileServer(http.FS(f))

	router.HandlerFunc(http.MethodGet, "/healthcheck", app.healthcheckHandler)

	// Wrap the /kamar-refresh handler in the authenticate middleware, to force an auth check on any request to this endpoint.
	router.HandlerFunc(http.MethodPost, "/kamar-refresh", app.authenticateKAMAR(app.kamarRefreshHandler))

	router.HandlerFunc(http.MethodGet, "/register", app.authenticateUser(app.registerPageHandler))
	router.HandlerFunc(http.MethodPost, "/register", app.authenticateUser(app.registerUserHandler))

	router.HandlerFunc(http.MethodGet, "/sign-in", app.authenticateUser(app.signInPageHandler))
	router.HandlerFunc(http.MethodPost, "/sign-in", app.authenticateUser(app.signInUserHandler))

	router.HandlerFunc(http.MethodPost, "/tokens/authentication", app.authenticateUser(app.createAuthenticationTokenHandler))

	// serve assets from /ui (httprouter's wild cards require a variable to be created, in this case filepath, to represent the value of the wildcard)
	router.HandlerFunc(http.MethodGet, "/assets/*filepath", app.authenticateUser(app.requireAuthenticatedUser(func(w http.ResponseWriter, r *http.Request) {
		app.assetHandler.ServeHTTP(w, r)
	})))
	// serve index.html from /ui if user is authenticated
	router.HandlerFunc(http.MethodGet, "/", app.authenticateUser(app.requireAuthenticatedUser(func(w http.ResponseWriter, r *http.Request) {
		app.assetHandler.ServeHTTP(w, r)
	})))

	return app.recoverPanic(app.rateLimit(app.processCORS(router)))
}
