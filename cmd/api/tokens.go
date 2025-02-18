package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/mjefferson-whs/listener/internal/data"
	"github.com/mjefferson-whs/listener/internal/validator"
)

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// Process JSON from client request into "input" struct
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	// Validate fields in "input" struct
	data.ValidateUsername(v, input.Username)
	data.ValidatePasswordPlaintext(v, input.Password)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Retrieve row from users table in database that matches the provided email
	user, err := app.models.Users.GetByUsername(input.Username)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsReponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	// Compare plaintext password provided by the client with hashed password from row retrieved from user table (SECURITY RISK - MAN-IN-MIDDLE/FAKE WEBSITE ATTACK??)
	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if !match {
		app.invalidCredentialsReponse(w, r)
		return
	}

	// Create an "authentication" token valid for 24 hours
	token, err := app.models.Tokens.New(user.ID, 24*time.Hour)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Return auth token to client
	err = app.writeJSON(w, http.StatusCreated, envelope{"authentication_token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) initiateTokenDeletionCycle() {
	app.background(func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				deleted, err := app.models.Tokens.DeleteExpiredTokens()
				if err != nil {
					app.logger.PrintError(err, nil)
				}
				app.logger.PrintInfo("purged expired tokens", map[string]interface{}{
					"tokensDeleted": deleted,
				})
			case <-app.isShuttingDown:
				app.logger.PrintInfo("token deletion cycle ending - shut down signal received", nil)
				return
			}
		}
	})
}
