package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/mjefferson-whs/listener/internal/data"
	"github.com/mjefferson-whs/listener/internal/validator"
)

type userInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (app *application) registerPageHandler(w http.ResponseWriter, r *http.Request) {
	if app.userExists {
		user := app.contextGetUser(r)

		if user.IsAnonymous() {
			app.authenticationRequiredRedirectResponse(w, r)
			return
		}
	}

	// Update URL path to reflect file path in embedded file system
	r.URL.Path = "register.html"

	app.assetHandler.ServeHTTP(w, r)
}

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	// If app.userExists, ensure user is authenticated
	if app.userExists {
		user := app.contextGetUser(r)

		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}
	}

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

	err = app.createAndSetAdminTokenCookie(w, user.ID, app.config.tokens.expiry)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.userExists = true

	err = app.writeJSON(w, http.StatusAccepted, envelope{"authenticated": true}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) signInPageHandler(w http.ResponseWriter, r *http.Request) {
	// Update URL path to reflect file path in embedded file system
	r.URL.Path = "sign-in.html"

	app.assetHandler.ServeHTTP(w, r)
}

func (app *application) signInUserHandler(w http.ResponseWriter, r *http.Request) {
	input := userInput{}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	data.ValidateUsername(v, input.Username)
	data.ValidatePasswordPlaintext(v, input.Password)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Retrieve row from users table that matches the provided username
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

	err = app.createAndSetAdminTokenCookie(w, user.ID, app.config.tokens.expiry)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) logoutUserHandler(w http.ResponseWriter, r *http.Request) {
	u := app.contextGetUser(r)
	id := u.ID

	err := app.models.Tokens.DeleteAllForUser(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsReponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Set the user for this request session to an anonymous user, and expire previously set cookies
	r = app.contextSetUser(r, data.AnonymousUser)
	http.SetCookie(w, &http.Cookie{
		Name:     "listener_admin_auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	http.Redirect(w, r, "/sign-in", http.StatusSeeOther)
	// err = app.writeJSON(w, http.StatusAccepted, nil, nil)
	// if err != nil {
	// 	app.serverErrorResponse(w, r, err)
	// }
}

func (app *application) getUserCount() (int, error) {
	var userCount int

	userCount, err := app.models.Users.GetUserCount()
	if err != nil {
		app.logger.PrintFatal(err, nil)
	}

	return userCount, err
}

// TODO: https://www.alexedwards.net/blog/working-with-cookies-in-go - add features
func (app *application) createAndSetAdminTokenCookie(w http.ResponseWriter, id int64, ttl time.Duration) error {
	token, err := app.models.Tokens.New(id, ttl)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "listener_admin_auth_token",
		Value:    token.Plaintext,
		Path:     "/",
		HttpOnly: true,
		Secure:   app.config.https_on,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(ttl),
	})

	return nil
}
