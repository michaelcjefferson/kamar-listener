package main

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mjefferson-whs/listener/internal/data"
	views "github.com/mjefferson-whs/listener/ui/views"
)

func (app *application) getDashboardPageHandler(c echo.Context) error {
	// TODO: If KAMAR auth hasn't been set, render DashboardPage with parameter false

	u := app.contextGetUser(c)

	w := data.WidgetData{}
	w.LastUpdated = time.Now()
	w.TotalRecords = 30247

	return app.Render(c, http.StatusOK, views.DashboardPage(u, true, w))
}

func (app *application) getHelpPageHandler(c echo.Context) error {
	u := app.contextGetUser(c)

	return app.Render(c, http.StatusOK, views.HelpPage(u))
}
