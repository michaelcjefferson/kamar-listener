package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mjefferson-whs/listener/internal/data"
	"github.com/mjefferson-whs/listener/internal/validator"
)

func (app *application) createAuthenticationTokenHandler(c echo.Context) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// Process JSON from client request into "input" struct
	err := c.Bind(&input)
	if err != nil {
		app.badRequestResponse(c, err)
		return
	}

	v := validator.New()

	// Validate fields in "input" struct
	data.ValidateUsername(v, input.Username)
	data.ValidatePasswordPlaintext(v, input.Password)

	if !v.Valid() {
		app.failedValidationResponse(c, v.Errors)
		return
	}

	// Retrieve row from users table in database that matches the provided email
	user, err := app.models.Users.GetByUsername(input.Username)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsReponse(c)
		default:
			app.serverErrorResponse(c, err)
		}

		return
	}

	// Compare plaintext password provided by the client with hashed password from row retrieved from user table (SECURITY RISK - MAN-IN-MIDDLE/FAKE WEBSITE ATTACK??)
	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}

	if !match {
		app.invalidCredentialsReponse(c)
		return
	}

	// Create an "authentication" token valid for 24 hours
	token, err := app.models.Tokens.New(user.ID, 24*time.Hour)
	if err != nil {
		app.serverErrorResponse(c, err)
		return
	}

	// Return auth token to client
	err = c.JSON(http.StatusCreated, envelope{"authentication_token": token})
	if err != nil {
		app.serverErrorResponse(c, err)
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
