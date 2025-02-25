package main

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// ------------General Responses------------ //
func (app *application) redirectResponse(c echo.Context, path string, jsonStatus int, message interface{}) error {
	var err error
	if strings.Contains(c.Request().Header.Get("Accept"), "application/json") {
		env := envelope{"message": message, "redirect": path}
		err = c.JSON(jsonStatus, env)
	} else {
		err = c.Redirect(http.StatusSeeOther, path)
	}
	if err != nil {
		app.logError(c, err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return err
}

// ------------KAMAR Responses------------ //
func (app *application) kamarResponse(c echo.Context, status int, j map[string]interface{}) error {
	c.Response().Header().Set(echo.HeaderServer, "WHS KAMAR Refresh")
	c.Response().Header().Set(echo.HeaderConnection, "close")

	env := envelope{"SMSDirectoryData": j}

	err := c.JSON(status, env)
	if err != nil {
		app.logError(c, err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return nil
}

// The three responses below meet the requirements of KAMAR by adding expected headers and the expected JSON body - only these two responses should ever be sent to KAMAR.
func (app *application) successResponse(c echo.Context) error {
	j := map[string]interface{}{
		"error":   0,
		"result":  "OK",
		"service": "WHS KAMAR Refresh",
		"version": "1.0",
	}

	return app.kamarResponse(c, http.StatusOK, j)
}

// NOTE: The expected failed response here: https://directoryservices.kamar.nz/?listening-service/standard-response - includes a Content-Length: 123 header, whereas Content-Length is only 82 with this response.
func (app *application) authFailedResponse(c echo.Context) error {
	j := map[string]interface{}{
		"error":   403,
		"result":  "Authentication Failed",
		"service": "WHS KAMAR Refresh",
		"version": "1.0",
	}

	return app.kamarResponse(c, http.StatusForbidden, j)
}

func (app *application) noCredentialsResponse(c echo.Context) error {
	j := map[string]interface{}{
		"error":   401,
		"result":  "No Credentials Provided",
		"service": "WHS KAMAR Refresh",
		"version": "1.0",
	}

	return app.kamarResponse(c, http.StatusUnauthorized, j)
}

func (app *application) checkResponse(c echo.Context) error {
	j := map[string]any{
		"error":             0,
		"result":            "OK",
		"service":           "WHS KAMAR Refresh",
		"version":           "1.1",
		"status":            "Ready",
		"infourl":           "https://wakatipu.school.nz/",
		"privacystatement":  "This service only collects results data, and stores it locally on a secure device. Only staff members of the school have access to the data.",
		"countryDataStored": "New Zealand",
		"options": map[string]any{
			"ics": true,
			"students": map[string]any{
				"details":         true,
				"passwords":       false,
				"photos":          false,
				"groups":          false,
				"awards":          false,
				"timetables":      false,
				"attendance":      false,
				"assessments":     true,
				"pastoral":        false,
				"learningsupport": false,
				"fields": map[string]string{
					"required": "firstname;lastname;gender;nsn",
					"optional": "username;caregivers;caregivers1;caregivers2;caregiver.name;caregiver.relationship;caregiver.mobile;caregiver.email",
				},
			},
			"common": map[string]bool{
				"subjects": false,
				"notices":  false,
				"calendar": false,
				"bookings": false,
			},
		},
	}

	return app.kamarResponse(c, http.StatusOK, j)
}
