package main

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mjefferson-whs/listener/internal/data"
	views "github.com/mjefferson-whs/listener/ui/views"
)

func (app *application) getLogHandler(c echo.Context) error {
	id, err := app.readIDParam(c)
	if err != nil {
		app.notFoundResponse(c)
		return err
	}

	log, err := app.models.Logs.GetForID(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(c)
		default:
			app.serverErrorResponse(c, err)
		}
		return err
	}

	return app.Render(c, http.StatusAccepted, views.IndividualLogPage(*log))
}

func (app *application) listLogsHandler(c echo.Context) {
	// var input struct {
	// 	data.Filters
	// }
}
