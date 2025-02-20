package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (app *application) healthcheckHandler(c echo.Context) error {
	return c.JSON(http.StatusAccepted, map[string]any{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
		},
	})
}

// func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
// 	env := envelope{
// 		"status": "available",
// 		"system_info": map[string]string{
// 			"environment": app.config.env,
// 		},
// 	}

// 	err := app.writeJSON(w, http.StatusOK, env, nil)
// 	if err != nil {
// 		app.serverErrorResponse(w, r, err)
// 	}
// }
