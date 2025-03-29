package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mjefferson-whs/listener/internal/data"
	"github.com/mjefferson-whs/listener/internal/validator"
	views "github.com/mjefferson-whs/listener/ui/views"
)

type userInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type PasswordUpdateInput struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (app *application) registerPageHandler(c echo.Context) error {
	if app.userExists {
		user := app.contextGetUser(c)

		if user.IsAnonymous() {
			app.authenticationRequiredResponse(c)
			return nil
		}
	}

	return app.Render(c, http.StatusAccepted, views.RegisterPage())
}

func (app *application) registerUserHandler(c echo.Context) error {
	// If app.userExists, ensure user is authenticated
	if app.userExists {
		user := app.contextGetUser(c)

		if user.IsAnonymous() {
			app.authenticationRequiredResponse(c)
			return nil
		}
	}

	input := userInput{}

	err := c.Bind(&input)
	if err != nil {
		app.badRequestResponse(c, err)
		return nil
	}

	user := &data.User{
		Username: input.Username,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(c, err)
		return nil
	}

	v := validator.New()

	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(c, v.Errors)
		return nil
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrUserAlreadyExists):
			v.AddError("username", "a user with this username already exists")
			app.failedValidationResponse(c, v.Errors)
		default:
			app.serverErrorResponse(c, err)
		}
		return nil
	}

	err = app.createAndSetAdminTokenCookie(c, user.ID, app.config.tokens.expiry)
	if err != nil {
		app.serverErrorResponse(c, err)
		return nil
	}

	app.userExists = true

	app.logger.PrintInfo("new user registered", map[string]interface{}{
		"user_id":    user.ID,
		"created at": user.CreatedAt,
		"username":   user.Username,
	})

	// return c.JSON(http.StatusAccepted, envelope{"authenticated": true})
	return app.redirectResponse(c, "/", http.StatusAccepted, envelope{"user": user, "authenticated": true})
}

// Prevent user from accessing sign in page and handler if they are already logged in
func (app *application) signInPageHandler(c echo.Context) error {
	user := app.contextGetUser(c)

	if !user.IsAnonymous() {
		return app.signOutRequiredResponse(c)
	}

	return app.Render(c, http.StatusAccepted, views.SignInPage())
}

func (app *application) signInUserHandler(c echo.Context) error {
	u := app.contextGetUser(c)

	if !u.IsAnonymous() {
		return app.signOutRequiredResponse(c)
	}

	input := userInput{}

	err := c.Bind(&input)
	if err != nil {
		return app.badRequestResponse(c, err)
	}

	v := validator.New()

	data.ValidateUsername(v, input.Username)
	data.ValidatePasswordPlaintext(v, input.Password)

	if !v.Valid() {
		return app.failedValidationResponse(c, v.Errors)
	}

	// Retrieve row from users table that matches the provided username
	user, err := app.models.Users.GetByUsername(input.Username)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			return app.invalidCredentialsResponse(c)
		default:
			return app.serverErrorResponse(c, err)
		}
	}

	// Compare plaintext password provided by the client with hashed password from row retrieved from user table (SECURITY RISK - MAN-IN-MIDDLE/FAKE WEBSITE ATTACK??)
	match, err := user.Password.Matches(input.Password)
	if err != nil {
		return app.serverErrorResponse(c, err)
	}

	if !match {
		return app.invalidCredentialsResponse(c)
	}

	err = app.createAndSetAdminTokenCookie(c, user.ID, app.config.tokens.expiry)
	if err != nil {
		return app.serverErrorResponse(c, err)
	}

	app.logger.PrintInfo("user logged in", map[string]interface{}{
		"user_id": user.ID,
	})

	return app.redirectResponse(c, "/", http.StatusAccepted, envelope{"user": user})
}

func (app *application) logoutUserHandler(c echo.Context) error {
	user := app.contextGetUser(c)

	if user.IsAnonymous() {
		return app.invalidAuthenticationTokenResponse(c)
	}

	deleted, err := app.models.Tokens.DeleteAllForUser(user.ID)
	if err != nil {
		switch {
		// TODO: This case currently cannot be triggered as DeleteAllForUser doesn't return this type of error
		case errors.Is(err, data.ErrRecordNotFound):
			return app.invalidCredentialsResponse(c)
		default:
			return app.serverErrorResponse(c, err)
		}
	}

	app.logger.PrintInfo("user logged out", map[string]interface{}{
		"user_id":        user.ID,
		"tokens deleted": deleted,
	})

	// Set the user for this request session to an anonymous user, and expire previously set cookies
	// TODO: Refactor into function
	app.contextSetUser(c, data.AnonymousUser)
	c.SetCookie(&http.Cookie{
		Name:     "listener_admin_auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	return app.redirectResponse(c, "/sign-in", http.StatusAccepted, "successfully logged out")
}

func (app *application) updateUserPasswordPageHandler(c echo.Context) error {
	u := app.contextGetUser(c)

	return app.Render(c, http.StatusAccepted, views.UpdatePasswordPage(u, false))
}

func (app *application) updateUserPasswordHandler(c echo.Context) error {
	// Ensure user making request is the same as user that password will be updated for
	user := app.contextGetUser(c)
	input := PasswordUpdateInput{}

	err := c.Bind(&input)
	if err != nil {
		app.logger.PrintError(err, map[string]any{
			"message": "error updating user password",
			"user_id": user.ID,
		})
		return app.badRequestResponse(c, err)
	}

	// Ensure user's password matches the CurrentPassword provided by the client
	match, err := user.Password.Matches(input.CurrentPassword)
	if err != nil {
		return app.serverErrorResponse(c, err)
	}
	if !match {
		return app.invalidCredentialsResponse(c)
	}

	v := validator.New()

	if data.ValidatePasswordPlaintext(v, input.NewPassword); !v.Valid() {
		return app.failedValidationResponse(c, v.Errors)
	}

	if err = user.Password.Set(input.NewPassword); err != nil {
		return app.serverErrorResponse(c, err)
	}

	err = app.models.Users.Update(*user)
	if err != nil {
		return app.serverErrorResponse(c, err)
	}

	env := envelope{
		"success": true,
	}

	return c.JSON(http.StatusAccepted, env)
}

func (app *application) getUsersPageHandler(c echo.Context) error {
	u := app.contextGetUser(c)

	users, err := app.models.Users.GetAll()
	if err != nil {
		return app.serverErrorResponse(c, err)
	}

	return app.Render(c, http.StatusAccepted, views.UsersPage(users, u))
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
func (app *application) createAndSetAdminTokenCookie(c echo.Context, id int64, ttl time.Duration) error {
	token, err := app.models.Tokens.New(id, ttl)
	if err != nil {
		return err
	}

	c.SetCookie(&http.Cookie{
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
