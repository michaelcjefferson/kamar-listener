package main

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mjefferson-whs/listener/internal/data"
	views "github.com/mjefferson-whs/listener/ui/views"
)

type ConfigUpdateRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type"`
}

func (app *application) configPageHandler(c echo.Context) error {
	config, err := app.models.Config.GetAll()
	if err != nil {
		return app.serverErrorResponse(c, err)
	}

	return app.Render(c, http.StatusOK, views.ConfigPage(config))
}

func (app *application) updateConfigHandler(c echo.Context) error {
	var req ConfigUpdateRequest
	user := app.contextGetUser(c)

	if err := c.Bind(&req); err != nil {
		return app.errorResponse(c, http.StatusBadRequest, "Invalid request")
	}

	// TODO: Validate

	// TODO: Hash password
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

	// TODO: Don't attach password values
	if err != nil {
		app.logger.PrintError(err, map[string]any{
			"key":     req.Key,
			"value":   req.Value,
			"user_id": user.ID,
		})
		return app.errorResponse(c, http.StatusInternalServerError, "Failed to update config")
	}

	// TODO: Don't attach password values
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
