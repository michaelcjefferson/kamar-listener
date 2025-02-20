package main

import (
	"net/http"

	"golang.org/x/time/rate"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (app *application) routes() http.Handler {
	router := echo.New()

	router.Use(app.recoverPanicMiddleware)
	// Only use rate limiter if enabled, and use custom values in config
	if app.config.limiter.enabled {
		router.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{
				Rate:  rate.Limit(app.config.limiter.rps),
				Burst: app.config.limiter.burst,
			},
		)))
	}
	// TODO: Set up CORS

	router.GET("/healthcheck", app.healthcheckHandler)

	// Wrap the /kamar-refresh handler in the authenticate middleware, to force an auth check on any request to this endpoint.
	// router.HandlerFunc(http.MethodPost, "/kamar-refresh", app.authenticateKAMAR(app.kamarRefreshHandler))

	// router.HandlerFunc(http.MethodGet, "/register", app.authenticateUser(app.registerPageHandler))
	// router.HandlerFunc(http.MethodPost, "/register", app.authenticateUser(app.registerUserHandler))

	// router.HandlerFunc(http.MethodGet, "/sign-in", app.authenticateUser(app.signInPageHandler))
	// router.HandlerFunc(http.MethodPost, "/sign-in", app.authenticateUser(app.signInUserHandler))

	// The user must be authenticated in order to be logged out successfully
	// router.HandlerFunc(http.MethodPost, "/log-out", app.authenticateUser(app.requireAuthenticatedUser(app.logoutUserHandler)))

	// router.HandlerFunc(http.MethodPost, "/tokens/authentication", app.authenticateUser(app.createAuthenticationTokenHandler))

	// serve assets from /ui (httprouter's wild cards require a variable to be created, in this case filepath, to represent the value of the wildcard)
	// router.HandlerFunc(http.MethodGet, "/assets/*filepath", app.authenticateUser(app.requireAuthenticatedUser(func(w http.ResponseWriter, r *http.Request) {
	// 	app.assetHandler.ServeHTTP(w, r)
	// })))
	// serve index.html from /ui if user is authenticated
	// router.HandlerFunc(http.MethodGet, "/", app.authenticateUser(app.requireAuthenticatedUser(func(w http.ResponseWriter, r *http.Request) {
	// 	app.assetHandler.ServeHTTP(w, r)
	// })))

	return router
	// return app.recoverPanic(app.rateLimit(app.processCORS(router)))
}
