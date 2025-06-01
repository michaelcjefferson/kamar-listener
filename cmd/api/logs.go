package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/michaelcjefferson/kamar-listener/internal/data"
	"github.com/michaelcjefferson/kamar-listener/internal/validator"
	"github.com/michaelcjefferson/kamar-listener/ui/components"
	views "github.com/michaelcjefferson/kamar-listener/ui/views"
)

func (app *application) getFilteredLogsPageHandler(c echo.Context) error {
	user := app.contextGetUser(c)

	filters := data.Filters{
		LogFilters: data.LogFilters{
			Level:   []string{},
			Message: "",
			UserID:  []int{},
		},
		Page:         1,
		PageSize:     10,
		Sort:         "-time",
		SortSafeList: []string{"level", "time", "user_id", "-level", "-time", "-user_id"},
	}

	filters.LogFilters.Level = c.QueryParams()["level"]
	filters.LogFilters.Message = c.QueryParam("message")
	if uids := c.QueryParams()["user_id"]; len(uids) != 0 {
		for _, val := range uids {
			u, err := strconv.Atoi(val)
			if err != nil {
				app.badRequestResponse(c, err)
				return err
			}
			filters.LogFilters.UserID = append(filters.LogFilters.UserID, u)
		}
	}

	// TODO: Refer to movies.go in greenlight to allow client-based sorting etc.
	if p := c.QueryParam("page"); p != "" {
		p, err := strconv.Atoi(p)
		if err != nil {
			app.badRequestResponse(c, err)
			return err
		}
		filters.Page = p
	}

	v := validator.New()
	if data.ValidateFilters(v, filters); !v.Valid() {
		return app.failedValidationResponse(c, v.Errors)
	}

	logs, metadata, logsMetadata, err := app.models.Logs.GetAll(filters)
	if err != nil {
		return app.serverErrorResponse(c, err)
	}

	return app.Render(c, http.StatusOK, views.LogsPage(logs, metadata, logsMetadata, filters, user))
}

func (app *application) getIndividualLogPageHandler(c echo.Context) error {
	u := app.contextGetUser(c)
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

	referer := c.Request().Header.Get("Referer")

	return app.Render(c, http.StatusAccepted, views.IndividualLogPage(*log, referer, u))
}

func (app *application) deleteIndividualLogHandler(c echo.Context) error {
	id, err := app.readIDParam(c)
	if err != nil {
		app.notFoundResponse(c)
		return err
	}

	err = app.models.Logs.DeleteForID(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(c)
		default:
			app.serverErrorResponse(c, err)
		}
		return err
	}

	return app.redirectResponse(c, "/logs", http.StatusAccepted, "log successfully deleted")
}

func (app *application) getFilteredLogsHandler(c echo.Context) error {
	var filters data.Filters

	v := validator.New()

	filters.LogFilters.Level = c.QueryParams()["level"]
	filters.LogFilters.Message = c.QueryParam("message")
	if uids := c.QueryParams()["user_id"]; len(uids) != 0 {
		for _, val := range uids {
			u, err := strconv.Atoi(val)
			if err != nil {
				app.badRequestResponse(c, err)
				return err
			}
			filters.LogFilters.UserID = append(filters.LogFilters.UserID, u)
		}
	}

	// TODO: Refer to movies.go in greenlight to allow client-based sorting etc.
	if p := c.QueryParam("page"); p != "" {
		p, err := strconv.Atoi(p)
		if err != nil || p < 1 || p > 20 {
			app.badRequestResponse(c, err)
			return err
		}
		filters.Page = p
	} else {
		filters.Page = 1
	}

	filters.PageSize = 10
	filters.Sort = "time"
	filters.SortSafeList = []string{"level", "time", "user_id"}

	if data.ValidateFilters(v, filters); !v.Valid() {
		app.failedValidationResponse(c, v.Errors)
		return nil
	}

	logs, metadata, _, err := app.models.Logs.GetAll(filters)
	if err != nil {
		app.serverErrorResponse(c, err)
		return err
	}

	return app.Render(c, http.StatusAccepted, components.LogsContainer(logs, metadata))
}
