package main

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mjefferson-whs/listener/internal/data"
	"github.com/mjefferson-whs/listener/internal/validator"
	views "github.com/mjefferson-whs/listener/ui/views"
	"golang.org/x/crypto/bcrypt"
)

func (app *application) configPageHandler(c echo.Context) error {
	u := app.contextGetUser(c)
	config, err := app.models.Config.GetAll()
	if err != nil {
		return app.serverErrorResponse(c, err)
	}

	return app.Render(c, http.StatusOK, views.ConfigPage(config, u))
}

func (app *application) updateConfigHandler(c echo.Context) error {
	var req data.ConfigEntry
	user := app.contextGetUser(c)

	if err := c.Bind(&req); err != nil {
		return app.errorResponse(c, http.StatusBadRequest, "Invalid request")
	}

	if req.Key == "listener_password" {
		return app.badRequestResponse(c, errors.New("password cannot be updated at this endpoint"))
	}

	v := validator.New()
	data.ValidateConfigUpdate(v, req)

	if !v.Valid() {
		return app.failedValidationResponse(c, v.Errors)
	}

	currentConfigVal, err := app.models.Config.GetByKey(req.Key)
	if err != nil && err != sql.ErrNoRows {
		return app.errorResponse(c, http.StatusNotFound, "Failed to get existing config")
	}

	newConfig := data.ConfigEntry{
		Key:         req.Key,
		Value:       req.Value,
		Type:        req.Type,
		Description: currentConfigVal.Description,
	}

	err = app.models.Config.Set(newConfig)

	if err != nil {
		app.logger.PrintError(err, map[string]any{
			"key":     req.Key,
			"value":   req.Value,
			"user_id": user.ID,
		})
		return app.errorResponse(c, http.StatusInternalServerError, "Failed to update config")
	}

	app.logger.PrintInfo("config updated", map[string]any{
		"key":     req.Key,
		"value":   req.Value,
		"user_id": user.ID,
	})

	// TODO: Add "success" field to all responses?
	updatedConfig, err := app.models.Config.GetByKey(req.Key)
	if err != nil {
		env := envelope{"success": true}
		return c.JSON(http.StatusAccepted, env)
	}

	env := envelope{
		"success":   true,
		"updatedAt": updatedConfig.UpdatedAt,
	}

	return c.JSON(http.StatusAccepted, env)
}

func (app *application) updateConfigPasswordPageHandler(c echo.Context) error {
	u := app.contextGetUser(c)

	return app.Render(c, http.StatusAccepted, views.UpdateListenerPasswordPage(u))
}

func (app *application) updateConfigPasswordHandler(c echo.Context) error {
	user := app.contextGetUser(c)
	input := PasswordUpdateInput{}

	err := c.Bind(&input)
	if err != nil {
		app.logger.PrintError(err, map[string]any{
			"message": "error updating listener password",
			"key":     "listener_password",
			"user_id": user.ID,
		})
		return app.badRequestResponse(c, err)
	}

	currPassConf, err := app.models.Config.GetByKey("listener_password")
	if err != nil {
		return app.serverErrorResponse(c, err)
	}

	// If there is already a password set in config, ensure it matches the CurrentPassword provided by the client
	if currPassConf.Value != "" {
		currPassHash := currPassConf.Value

		err := bcrypt.CompareHashAndPassword([]byte(currPassHash), []byte(input.CurrentPassword))
		if err != nil {
			app.logger.PrintError(err, map[string]any{
				"message": "error updating listener password",
				"key":     "listener_password",
				"user_id": user.ID,
			})
			if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
				return app.invalidCredentialsResponse(c)
			}
			return app.serverErrorResponse(c, err)
		}
	}

	v := validator.New()

	if data.ValidatePasswordPlaintext(v, input.NewPassword); !v.Valid() {
		app.logger.PrintError(errors.New("updating listener password - failed validation"), map[string]any{
			"message": "error updating listener password",
			"key":     "listener_password",
			"user_id": user.ID,
		})
		return app.failedValidationResponse(c, v.Errors)
	}

	p := data.Password{}
	if err = p.Set(input.NewPassword); err != nil {
		app.logger.PrintError(err, map[string]any{
			"message": "error updating listener password",
			"key":     "listener_password",
			"user_id": user.ID,
		})
		return app.serverErrorResponse(c, err)
	}

	configEntry := data.ConfigEntry{
		Key:         "listener_password",
		Value:       string(p.Hash()),
		Type:        "password",
		Description: currPassConf.Description,
	}

	err = app.models.Config.Set(configEntry)
	if err != nil {
		app.logger.PrintError(err, map[string]any{
			"message": "error updating listener password",
			"key":     "listener_password",
			"user_id": user.ID,
		})
		return app.serverErrorResponse(c, err)
	}

	app.logger.PrintInfo("config updated", map[string]any{
		"message": "successfully updated listener password",
		"key":     "listener_password",
		"user_id": user.ID,
	})

	env := envelope{
		"success": true,
	}

	return c.JSON(http.StatusAccepted, env)
}
