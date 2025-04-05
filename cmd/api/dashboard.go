package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	views "github.com/mjefferson-whs/listener/ui/views"
)

func (app *application) dashboardPageHandler(c echo.Context) error {
	// TODO: If KAMAR auth hasn't been set, render DashboardPage with parameter true

	u := app.contextGetUser(c)
	return app.Render(c, http.StatusOK, views.DashboardPage(u, true))
}
