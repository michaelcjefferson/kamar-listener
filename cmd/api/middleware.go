package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mjefferson-whs/listener/internal/data"
	"github.com/mjefferson-whs/listener/internal/validator"
)

func (app *application) recoverPanicMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Deferring a function ensures it will execute as the stack is unwound in the case of a panic. It won't be used elsewhere, so an anonymous function works well.
		defer func() {
			// recover() is a built-in function that checks whether or not there has been a panic.
			if err := recover(); err != nil {
				// Set header which tells the server to close the connection after this has been sent.
				c.Response().Header().Set(echo.HeaderConnection, "close")
				app.logger.PrintError(err.(error), map[string]interface{}{
					"OHNO": "couldn't recover from this error",
				})
				c.Error(fmt.Errorf("%v", err))
			}
		}()

		return next(c)
	}
}

// Authenticate requests received from KAMAR itself, using the required Basic authentication
func (app *application) authenticateKAMAR(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			app.logger.PrintInfo("listener: failed at authHeader", nil)
			app.kamarNoCredentialsResponse(c)
			return errors.New("failed at authHeader")
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Basic" {
			app.logger.PrintInfo("listener: failed at headerParts", nil)
			app.kamarAuthFailedResponse(c)
			return errors.New("failed at headerParts")
		}

		decodedAuth, err := base64.StdEncoding.DecodeString(headerParts[1])
		if err != nil {
			app.logger.PrintInfo("listener: failed at decodedAuth", nil)
			app.kamarAuthFailedResponse(c)
			return errors.New("failed at decodedAuth")
		}

		// TODO: Compare authCredentials to config from DB, rather than app.config.credentials
		authCredentials := strings.Split(string(decodedAuth), ":")
		if len(authCredentials) != 2 || authCredentials[0] != app.config.credentials.username || authCredentials[1] != app.config.credentials.password {
			logInfo := make(map[string]interface{})
			logInfo["creds"] = authCredentials
			logInfo["creds_length"] = len(authCredentials)
			logInfo["app_user"] = app.config.credentials.username
			logInfo["app_pass"] = app.config.credentials.password
			logInfo["req_user"] = authCredentials[0]
			logInfo["req_pass"] = authCredentials[1]
			app.logger.PrintInfo("listener: failed at authCredentials", logInfo)
			app.kamarAuthFailedResponse(c)
			return errors.New("failed at authCredentials")
		}

		app.logger.PrintInfo("listener: successfully authenticated request from KAMAR", nil)
		return next(c)
	}
}

// If a valid auth token is provided, set "user" value in request context to a struct containing the corresponding user's data. If an invalid token is provided, send an error.
func (app *application) authenticateUser(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		// TODO: Add a check for r.Context().Value("isAuthenticated").(bool) to prevent extra look-ups if user is already authenticated, and set the isAuthenticated value below once user has been found. Ensure that isAuthenticated doesn't lead to leaky security, where this can be parsed as true even if user has no or an expired token.

		// Get the http-only cookie containing the token from the request, and convert to a string
		cookie, err := c.Cookie("listener_admin_auth_token")

		// If the cookie can't be found, the user is not authenticated and should be set as an anonymous user
		if err == http.ErrNoCookie {
			app.contextSetUser(c, data.AnonymousUser)
			return next(c)
		}

		token := cookie.Value

		v := validator.New()

		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			return app.invalidAuthenticationTokenResponse(c)
		}

		// Retrieve user data from user table based on the token provided.
		user, tokenExpiry, err := app.models.Users.GetForToken(token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				return app.invalidAuthenticationTokenResponse(c)
			default:
				return app.serverErrorResponse(c, err)
			}
		}

		// If token has only a short time before expiry, create a new token for that user
		expiryTime, err := time.Parse(time.RFC3339, tokenExpiry)
		if err != nil {
			return app.serverErrorResponse(c, err)
		}

		expiryTimeFrame := time.Now().Add(app.config.tokens.refresh)

		// Check if the token expiry is within the timeframe, and if so, generate a new token and return it
		if expiryTime.Before(expiryTimeFrame) {
			app.logger.PrintInfo("token near expiry - creating new token and sending to user", map[string]interface{}{
				"user id":           user.ID,
				"expiry time":       tokenExpiry,
				"expiry time frame": expiryTimeFrame,
			})
			app.createAndSetAdminTokenCookie(c, user.ID, app.config.tokens.expiry)
		}

		// Attach user data to context
		app.contextSetUser(c, user)

		// Call next handler in the chain.
		return next(c)
	}
}

// TODO: Add activation and requireActivatedUser for new user registrations after initial set-up - admins can go to an add user page, and enter an email address to send an activation code to. This creates an activation token in the database, and provides it as part of a link for the admin to copy and paste into an email to the new user. The new user can follow that link to be brought to an activation page, where they create a username and password, and a new account is created. Activation tokens valid for 24 (?) hours

// Runs after authenticate, only needed on protected routes - checks the context for the value of the user set by authenticate, and at this point only ensures that one exists, as it means that someone is logged in and can access protected routes
func (app *application) requireAuthenticatedUser(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := app.contextGetUser(c)

		if user.IsAnonymous() {
			return app.authenticationRequiredResponse(c)
		}

		return next(c)
	}
}

// // TODO: Add to config for app, including instructions to find IP address of KAMAR instance
// func (app *application) processCORS(next http.Handler) http.Handler {
// 	c := cors.New(cors.Options{
// 		AllowedOrigins: []string{"https://localhost", "https://0.0.0.0"},
// 		// AllowedOrigins:   []string{"https://localhost", "https://10.100"},
// 		AllowCredentials: true,
// 		AllowedHeaders:   []string{"Origin", "Authorization", "Content-Type"},
// 		AllowedMethods:   []string{"GET", "POST"},
// 		// AllowedMethods:   []string{"POST"},
// 		Debug: true,
// 	})

// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		c.Handler(next).ServeHTTP(w, r)
// 	})
// }
