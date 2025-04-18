package main

import (
	"errors"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
	w.RecordsToday = 258
	w.TotalRecords = 30247

	aDBStat, err := os.Stat(app.config.app_db_path)
	if err != nil {
		app.serverErrorResponse(c, err)
		panic(err)
	}
	lDBStat, err := os.Stat(app.config.kamar_data_db_path)
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

	w.TotalLogs = int64(logMeta.TotalRecords)
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

	w.TotalErrors = int64(errLogMeta.TotalRecords)
	w.RecentErrorLogs = errLogs

	return app.Render(c, http.StatusOK, views.DashboardPage(u, true, w))
}

func (app *application) openDataFolderHandler(c echo.Context) error {
	appDBpath := app.config.app_db_path
	folderPath := filepath.Dir(appDBpath)

	// Check whether app DB exists - if not, return an error
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		return app.badRequestResponse(c, errors.New("Can't find database folder on your device - make sure that you are using the device that the application is running on."))
	}

	// Open DB folder in whichever file explorer the user's OS uses
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", folderPath)
	case "darwin":
		cmd = exec.Command("open", folderPath)
	case "linux":
		cmd = exec.Command("xdg-open", folderPath)
	default:
		return app.badRequestResponse(c, errors.New("Unsupported operating system"))
	}

	if err := cmd.Start(); err != nil {
		return app.serverErrorResponse(c, err)
	}

	return c.NoContent(http.StatusOK)
}
