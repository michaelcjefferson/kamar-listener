package main

import (
	"fmt"
	"net/http"
)

func (app *application) logError(r *http.Request, err error) {
	app.logger.PrintError(err, map[string]interface{}{
		"request_ip":     r.RemoteAddr,
		"request_method": r.Method,
		"request_url":    r.URL.String(),
	})
}

func (app *application) logRequest(r *http.Request, message string) {
	app.logger.PrintInfo(message, map[string]interface{}{
		"request_ip":     r.RemoteAddr,
		"request_method": r.Method,
		"request_url":    r.URL.String(),
	})
}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	env := envelope{"error": message}

	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}

func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)

	message := "the server encountered a problem and could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}

func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	app.logRequest(r, message)
	app.errorResponse(w, r, http.StatusTooManyRequests, message)
}

func (app *application) invalidCredentialsReponse(w http.ResponseWriter, r *http.Request) {
	message := "invalid authentication credentials"
	app.logRequest(r, "log in attempt failed")
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

func (app *application) invalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	// This header informs the user that a bearer token should be used to authenticate.
	// w.Header().Set("WWW-Authenticate", "Bearer")
	w.Header().Set("Location", "/sign-in")
	http.SetCookie(w, &http.Cookie{
		Name:     "listener_admin_auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	message := "invalid or missing authentication token"
	app.errorResponse(w, r, http.StatusSeeOther, message)
	// app.errorResponse(w, r, http.StatusUnauthorized, message)
}

// For browser requests - redirect the client to the sign-in page
func (app *application) authenticationRequiredRedirectResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Location", "/sign-in")

	message := "you must be authenticated to access this resource"
	app.errorResponse(w, r, http.StatusSeeOther, message)
}

// For API requests - respond with Forbidden status code
func (app *application) authenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	message := "you must be authenticated to access this resource"
	app.errorResponse(w, r, http.StatusForbidden, message)
}

// For browser requests - redirect the client to the sign-in page
func (app *application) signOutRequiredRedirectResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Location", "/")

	message := "you must be signed out to access this resource"
	app.errorResponse(w, r, http.StatusSeeOther, message)
}

// For API requests - respond with Forbidden status code
func (app *application) signOutRequiredResponse(w http.ResponseWriter, r *http.Request) {
	message := "you must be signed out to access this resource"
	app.errorResponse(w, r, http.StatusForbidden, message)
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}
