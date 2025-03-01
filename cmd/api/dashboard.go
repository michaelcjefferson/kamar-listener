package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	views "github.com/mjefferson-whs/listener/ui/views"
)

func (app *application) dashboardPageHandler(c echo.Context) error {
	return app.Render(c, http.StatusOK, views.DashboardPage())
}

func (app *application) logsPageHandler(c echo.Context) error {
	return app.Render(c, http.StatusOK, views.LogsPage(nil))
}

func (app *application) configPageHandler(c echo.Context) error {
	return app.Render(c, http.StatusOK, views.ConfigPage())
}
