package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	views "github.com/mjefferson-whs/listener/ui/views"
)

func (app *application) dashboardHandler(c echo.Context) error {
	app.logger.PrintInfo("hit dashboardHandler", nil)
	return app.Render(c, http.StatusOK, views.Dashboard())
}
