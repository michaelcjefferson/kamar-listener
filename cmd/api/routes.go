package main

import (
	"embed"
	"net/http"

	"golang.org/x/time/rate"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//go:embed assets/*
var embeddedAssets embed.FS

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

	// TO SET UP EMBEDDED ASSETS OR FILES WITH ECHO,
	// 1) the assets must be in the same directory as main.go, so the go:embed can find them (as far as I can tell)
	// 2) use echo.MustSubFS() to clean route-file matching, eg. the html is searching for url.com/style.css - seeing as that file is in the /assets folder, MustSubFS navigates the file system inside of /assets
	// 3) then, use router.StaticFS to attach the files as static routes
	assetFS := echo.MustSubFS(embeddedAssets, "assets")

	router.GET("/healthcheck", app.healthcheckHandler)

	// Routes that require authentication
	// To set middleware on any route branching off of "/", including the "/" route itself, echo requires the route group to be set on "" rather than "/" as it appends a trailing slash.
	authGroup := router.Group("", app.authenticateUser)
	authGroup.GET("/register", app.registerPageHandler)
	authGroup.POST("/register", app.registerUserHandler)

	authGroup.GET("/sign-in", app.signInPageHandler)
	authGroup.POST("/sign-in", app.signInUserHandler)

	authGroup.StaticFS("/", assetFS)
	authGroup.GET("/", app.dashboardHandler)

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
