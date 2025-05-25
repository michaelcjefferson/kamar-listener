package main

import (
	"errors"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/labstack/echo/v4"
	"github.com/michaelcjefferson/kamar-listener/internal/data"
	views "github.com/michaelcjefferson/kamar-listener/ui/views"
)

func (app *application) getDashboardPageHandler(c echo.Context) error {
	u := app.contextGetUser(c)

	w := data.WidgetData{}

	w.LastCheckTime, w.LastInsertTime, w.RecordsToday, w.TotalRecords, w.CountByType = app.appMetrics.Snapshot()

	w.IP = app.config.ip

	aDBStat, err := os.Stat(app.config.dbPaths.appDB)
	if err != nil {
		app.serverErrorResponse(c, err)
		panic(err)
	}
	lDBStat, err := os.Stat(app.config.dbPaths.listenerDB)
	if err != nil {
		app.serverErrorResponse(c, err)
		panic(err)
	}
	appDBSize := float64(aDBStat.Size()) / 1024.0 / 1024.0
	listenerDBSize := float64(lDBStat.Size()) / 1024.0 / 1024.0
	w.DBSize = appDBSize + listenerDBSize

	logFilters := data.Filters{
		LogFilters:   data.LogFilters{},
		Page:         1,
		PageSize:     5,
		Sort:         "-time",
		SortSafeList: []string{"level", "time", "user_id", "-level", "-time", "-user_id"},
	}
	logs, logMeta, _, err := app.models.Logs.GetAll(logFilters)
	if err != nil {
		return app.serverErrorResponse(c, err)
	}

	w.TotalLogs = logMeta.TotalRecords
	w.RecentLogs = logs

	errLogFilters := data.Filters{
		LogFilters: data.LogFilters{
			Level: []string{"ERROR"},
		},
		Page:         1,
		PageSize:     5,
		Sort:         "-time",
		SortSafeList: []string{"level", "time", "user_id", "-level", "-time", "-user_id"},
	}
	errLogs, errLogMeta, _, err := app.models.Logs.GetAll(errLogFilters)
	if err != nil {
		return app.serverErrorResponse(c, err)
	}

	w.TotalErrors = errLogMeta.TotalRecords
	w.RecentErrorLogs = errLogs

	events, err := app.models.ListenerEvents.GetLastWeek()
	if err != nil {
		app.logger.PrintError(err, map[string]any{
			"message": "couldn't get listener events from database",
		})
	}
	w.Events = events

	w.JSONEnabled = app.config.kamar_write_to_json

	return app.Render(c, http.StatusOK, views.DashboardPage(u, w))
}

// Open the directory that holds the application databases on the client's computer, if it exists
func (app *application) openDataFolderHandler(c echo.Context) error {
	// Check whether app database directory exists - if not, return an error
	if _, err := os.Stat(app.config.dbPaths.dbDir); os.IsNotExist(err) {
		return app.badRequestResponse(c, errors.New("Can't find database folder on your device - make sure that you are using the device that the application is running on."))
	}

	// Open DB folder in whichever file explorer the user's OS uses
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", app.config.dbPaths.dbDir)
	case "darwin":
		cmd = exec.Command("open", app.config.dbPaths.dbDir)
	case "linux":
		cmd = exec.Command("xdg-open", app.config.dbPaths.dbDir)
	default:
		return app.badRequestResponse(c, errors.New("Unsupported operating system"))
	}

	if err := cmd.Start(); err != nil {
		return app.serverErrorResponse(c, err)
	}

	return c.NoContent(http.StatusOK)
}

type JSONSwitchRequest struct {
	Enabled bool `json:"enabled"`
}

func (app *application) jsonSwitchHandler(c echo.Context) error {
	var req JSONSwitchRequest

	if err := c.Bind(&req); err != nil {
		return app.serverErrorResponse(c, err)
	}

	app.config.kamar_write_to_json = req.Enabled
	return c.NoContent(http.StatusOK)
}
