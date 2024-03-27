package main

import (
	"net/http"
)

func (app *application) kamarResponse(w http.ResponseWriter, r *http.Request, status int, j map[string]interface{}) {
	w.Header().Set("Server", "WHS KAMAR Refresh/1.0")
	w.Header().Set("Connection", "close")

	env := envelope{"SMSDirectoryData": j}

	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}

// The two responses below meet the requirements of KAMAR by adding expected headers and the expected JSON body - only these two responses should ever be sent to KAMAR.
func (app *application) successResponse(w http.ResponseWriter, r *http.Request) {
	j := map[string]interface{}{
		"error":  0,
		"result": "OK",
	}

	app.kamarResponse(w, r, http.StatusOK, j)
}

// NOTE: The expected failed response here: https://directoryservices.kamar.nz/?listening-service/standard-response - includes a Content-Length: 123 header, whereas Content-Length is only 82 with this response.
func (app *application) authFailedResponse(w http.ResponseWriter, r *http.Request) {
	j := map[string]interface{}{
		"error":  403,
		"result": "Authentication Failed",
	}

	app.kamarResponse(w, r, http.StatusForbidden, j)
}

func (app *application) noCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	j := map[string]interface{}{
		"error":  401,
		"result": "No Credentials Provided",
	}

	app.kamarResponse(w, r, http.StatusUnauthorized, j)
}

func (app *application) checkResponse(w http.ResponseWriter, r *http.Request) {
	j := map[string]interface{}{
		"error":            0,
		"result":           "OK",
		"status":           "Ready",
		"infourl":          "none",
		"privacystatement": "To be stated.",
		"options": map[string]interface{}{
			"common": map[string]interface{}{
				"results": true,
			},
		},
	}

	app.kamarResponse(w, r, http.StatusOK, j)
}
