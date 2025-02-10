package main

import (
	"errors"
	"net/http"

	"github.com/mjefferson-whs/listener/internal/data"
	"github.com/mjefferson-whs/listener/internal/validator"
)

type userInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (app *application) registerPageHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.PrintInfo("registerPageHandler hit", nil)
	if app.userExists {
		http.Redirect(w, r, "/sign-in", http.StatusForbidden)
		return
	}

	// Update URL path to reflect file path in embedded file system
	r.URL.Path = "register.html"

	// TODO: serve register page
	app.assetHandler.ServeHTTP(w, r)
}

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: if app.userExists, ensure user is authenticated
	input := userInput{}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Username: input.Username,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrUserAlreadyExists):
			v.AddError("username", "a user with this username already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	app.userExists = true

	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getUserCount() (int, error) {
	var userCount int

	userCount, err := app.models.Users.GetUserCount()
	if err != nil {
		app.logger.PrintFatal(err, nil)
	}

	return userCount, err
}
