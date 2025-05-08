package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func (app *application) logError(c echo.Context, err error) {
	app.logger.PrintError(err, map[string]any{
		"request_ip":     c.RealIP(),
		"request_method": c.Request().Method,
		"request_url":    c.Request().URL,
	})
}

func (app *application) logRequest(c echo.Context, message string) {
	app.logger.PrintInfo(message, map[string]any{
		"request_ip":     c.RealIP(),
		"request_method": c.Request().Method,
		"request_url":    c.Request().URL,
	})
}

func (app *application) errorResponse(c echo.Context, status int, message any) error {
	env := envelope{"error": message}
	// env := envelope{
	// 	"success": false,
	// 	"error":   message,
	// }

	err := c.JSON(status, env)
	if err != nil {
		app.logError(c, err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return nil
}

// If the client accepts JSON (which will be the case when using fetch() in the browser for API calls, CURL etc.), provide a JSON response including a status code, redirect path to follow if desired by the client, and error message - otherwise respond with an http.Redirect. http.Redirect will occur if a client tries to access a page via its URL - it is a GET request and doesn't include the Accepts JSON header
func (app *application) redirectErrorResponse(c echo.Context, path string, jsonStatus int, message any) error {
	var err error
	if strings.Contains(c.Request().Header.Get("Accept"), "application/json") {
		env := envelope{"error": message, "redirect": path}
		err = c.JSON(jsonStatus, env)
	} else {
		err = c.Redirect(http.StatusSeeOther, path)
	}
	if err != nil {
		app.logError(c, err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return err
}

func (app *application) serverErrorResponse(c echo.Context, err error) error {
	app.logError(c, err)

	message := "the server encountered a problem and could not process your request"
	return app.errorResponse(c, http.StatusInternalServerError, message)
}

func (app *application) notFoundResponse(c echo.Context) error {
	message := "the requested resource could not be found"
	return app.errorResponse(c, http.StatusNotFound, message)
}

func (app *application) methodNotAllowedResponse(c echo.Context) error {
	message := fmt.Sprintf("the %s method is not supported for this resource", c.Request().Method)
	return app.errorResponse(c, http.StatusMethodNotAllowed, message)
}

func (app *application) rateLimitExceededResponse(c echo.Context) error {
	message := "rate limit exceeded"
	app.logRequest(c, message)
	return app.errorResponse(c, http.StatusTooManyRequests, message)
}

func (app *application) invalidCredentialsResponse(c echo.Context) error {
	message := "invalid authentication credentials"
	app.logRequest(c, "log in attempt failed")
	return app.errorResponse(c, http.StatusUnauthorized, message)
}

func (app *application) invalidAuthenticationTokenResponse(c echo.Context) error {
	// This header informs the user that a bearer token should be used to authenticate.
	// w.Header().Set("WWW-Authenticate", "Bearer")
	c.SetCookie(&http.Cookie{
		Name:     "listener_admin_auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	message := "invalid or missing authentication token"
	return app.redirectErrorResponse(c, "/sign-in", http.StatusUnauthorized, message)
}

// For browser requests - redirect the client to the sign-in page
func (app *application) authenticationRequiredResponse(c echo.Context) error {
	message := "you must be authenticated to access this resource"
	return app.redirectErrorResponse(c, "/sign-in", http.StatusForbidden, message)
}

// For browser requests - redirect the client to the sign-in page
func (app *application) signOutRequiredResponse(c echo.Context) error {
	message := "you must be signed out to access this resource"
	return app.redirectErrorResponse(c, "/", http.StatusForbidden, message)
}

func (app *application) badRequestResponse(c echo.Context, err error) error {
	return app.errorResponse(c, http.StatusBadRequest, err.Error())
}

func (app *application) failedValidationResponse(c echo.Context, errors map[string]string) error {
	return app.errorResponse(c, http.StatusUnprocessableEntity, errors)
}
