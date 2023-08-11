package main

import (
	"fmt"
	"net/http"
)

// The two responses below meet the requirements of KAMAR by adding expected headers and the expected JSON body - only these two responses should ever be sent to KAMAR.
func (app *application) successResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", "WHS KAMAR Refresh/1.0")
	w.Header().Set("Connection", "close")

	j := map[string]interface{}{
		"error":  0,
		"result": "OK",
	}

	env := envelope{"SMSDirectoryData": j}

	err := app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}

// NOTE: The expected failed response here: https://directoryservices.kamar.nz/?listening-service/standard-response - includes a Content-Length: 123 header, whereas Content-Length is only 82 with this response.
func (app *application) failedResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", "WHS KAMAR Refresh/1.0")
	w.Header().Set("Connection", "close")

	j := map[string]interface{}{
		"error":  403,
		"result": "Authentication Failed",
	}

	env := envelope{"SMSDirectoryData": j}

	err := app.writeJSON(w, http.StatusForbidden, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}

func (app *application) logError(r *http.Request, err error) {
	app.logger.PrintError(err, map[string]string{
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
	app.errorResponse(w, r, http.StatusTooManyRequests, message)
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}
