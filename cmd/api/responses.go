package main

import (
	"net/http"
)

func (app *application) kamarResponse(w http.ResponseWriter, r *http.Request, status int, j map[string]interface{}) {
	w.Header().Set("Server", "WHS KAMAR Refresh")
	w.Header().Set("Connection", "close")

	env := envelope{"SMSDirectoryData": j}

	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}

// The three responses below meet the requirements of KAMAR by adding expected headers and the expected JSON body - only these two responses should ever be sent to KAMAR.
func (app *application) successResponse(w http.ResponseWriter, r *http.Request) {
	j := map[string]interface{}{
		"error":   0,
		"result":  "OK",
		"service": "WHS KAMAR Refresh",
		"version": "1.0",
	}

	app.kamarResponse(w, r, http.StatusOK, j)
}

// NOTE: The expected failed response here: https://directoryservices.kamar.nz/?listening-service/standard-response - includes a Content-Length: 123 header, whereas Content-Length is only 82 with this response.
func (app *application) authFailedResponse(w http.ResponseWriter, r *http.Request) {
	j := map[string]interface{}{
		"error":   403,
		"result":  "Authentication Failed",
		"service": "WHS KAMAR Refresh",
		"version": "1.0",
	}

	app.kamarResponse(w, r, http.StatusForbidden, j)
}

func (app *application) noCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	j := map[string]interface{}{
		"error":   401,
		"result":  "No Credentials Provided",
		"service": "WHS KAMAR Refresh",
		"version": "1.0",
	}

	app.kamarResponse(w, r, http.StatusUnauthorized, j)
}

func (app *application) checkResponse(w http.ResponseWriter, r *http.Request) {
	j := map[string]any{
		"error":            0,
		"result":           "OK",
		"service":          "WHS KAMAR Refresh",
		"version":          "1.0",
		"status":           "Ready",
		"infourl":          "https://wakatipu.school.nz/",
		"privacystatement": "This service only collects results data, and stores it locally on a secure device. Only staff members of the school have access to the data.",
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

	app.kamarResponse(w, r, http.StatusOK, j)
}
