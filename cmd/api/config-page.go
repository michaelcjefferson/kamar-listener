package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	views "github.com/mjefferson-whs/listener/ui/views"
)

func (app *application) configPageHandler(c echo.Context) error {
	config, err := app.models.Config.GetAll()
	if err != nil {
		return app.serverErrorResponse(c, err)
	}

	return app.Render(c, http.StatusOK, views.ConfigPage(config))
}
