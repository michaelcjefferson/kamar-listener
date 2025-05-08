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

	router.Pre(middleware.RemoveTrailingSlash())

	// TO SET UP EMBEDDED ASSETS OR FILES WITH ECHO,
	// 1) the assets must be in the same directory as main.go, so the go:embed can find them (as far as I can tell)
	// 2) use echo.MustSubFS() to clean route-file matching, eg. the html is searching for url.com/style.css - seeing as that file is in the /assets folder, MustSubFS navigates the file system inside of /assets
	// 3) then, use router.StaticFS to attach the files as static routes
	assetFS := echo.MustSubFS(embeddedAssets, "assets")

	router.GET("/healthcheck", app.healthcheckHandler)

	// Routes that require authentication - clients accessing these routes will pass through authentication, being set as either the user associated with their cookie or an anonymous user
	// To set middleware on any route branching off of "/", including the "/" route itself, echo requires the route group to be set on "" rather than "/" as it appends a trailing slash.
	authGroup := router.Group("", app.authenticateUser)
	authGroup.GET("/register", app.registerPageHandler)
	authGroup.POST("/register", app.registerUserHandler)

	authGroup.GET("/sign-in", app.signInPageHandler)
	authGroup.POST("/sign-in", app.signInUserHandler)

	// Auth not required to request assets
	// TODO: Separate assets for authed and unauthed, and only serve unauthed assets to unauthed users
	authGroup.StaticFS("/", assetFS)

	// Routes that require the user to be successfully authenticated
	// All routes in this group first pass through authenticateUser, as it is defined on top of the authGroup.Group
	isAuthenticatedGroup := authGroup.Group("", app.requireAuthenticatedUser)
	// The user must be authenticated in order to be logged out successfully, and to reach the dashboard
	isAuthenticatedGroup.POST("/log-out", app.logoutUserHandler)

	isAuthenticatedGroup.POST("/config/set/auth", app.setKamarAuthHandler)

	// Routes in this group will be redirected to a KAMAR auth set up page if a username and password for KAMAR directory service haven't been set in the database
	kamarAuthSetGroup := isAuthenticatedGroup.Group("", app.requireKAMARAuthSetUp)
	kamarAuthSetGroup.GET("/config/update/password", app.updateConfigPasswordPageHandler)
	kamarAuthSetGroup.POST("/config/update/password", app.updateConfigPasswordHandler)
	kamarAuthSetGroup.POST("/config/update", app.updateConfigHandler)
	kamarAuthSetGroup.GET("/config", app.configPageHandler)

	isAuthenticatedGroup.GET("/logs/partial", app.getFilteredLogsHandler)
	isAuthenticatedGroup.GET("/logs/:id", app.getIndividualLogPageHandler)
	isAuthenticatedGroup.GET("/logs", app.getFilteredLogsPageHandler)

	isAuthenticatedGroup.GET("/users/update/password", app.updateUserPasswordPageHandler)
	isAuthenticatedGroup.POST("/users/update/password", app.updateUserPasswordHandler)
	isAuthenticatedGroup.GET("/users/delete", app.deleteUserHandler)
	isAuthenticatedGroup.GET("/users", app.getUsersPageHandler)

	isAuthenticatedGroup.GET("/help", app.getHelpPageHandler)

	isAuthenticatedGroup.GET("/opendatafolder", app.openDataFolderHandler)

	kamarAuthSetGroup.GET("/", app.getDashboardPageHandler)

	// Wrap the /kamar-refresh handler in the authenticate middleware, to force an auth check on any request to this endpoint.
	kamarAuthGroup := router.Group("/kamar-refresh", app.authenticateKAMAR)
	kamarAuthGroup.POST("", app.kamarRefreshHandler)

	// router.HandlerFunc(http.MethodPost, "/tokens/authentication", app.authenticateUser(app.createAuthenticationTokenHandler))

	return router
}
