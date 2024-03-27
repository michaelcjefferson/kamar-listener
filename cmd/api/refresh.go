package main

import (
	"encoding/json"
	"net/http"
	"os"
)

type KAMARData struct {
	Data map[string]interface{} `json:"SMSDirectoryData"`
}

func (app *application) kamarRefreshHandler(w http.ResponseWriter, r *http.Request) {
	// var input map[string]interface{}
	var kamarData KAMARData

	err := app.readJSON(w, r, &kamarData)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// syncType, ok := input["SMSDirectoryData"]["sync"].(string)
	syncType, ok := kamarData.Data["sync"].(string)
	if !ok || syncType == "" {
		app.logger.PrintInfo("failed to get syncType from input", kamarData.Data)
		app.authFailedResponse(w, r)
		return
	}

	if syncType == "check" {
		app.checkResponse(w, r)
		app.logger.PrintInfo("received and processed check request", kamarData.Data)
		return
	}

	// MarshalIndent should be used for testing as it indents the JSON that it creates, making it more readable. However, Marshal should be used in practise instead because it consumes less resources (especially important due to the size of the JSON file that is created).
	output, _ := json.MarshalIndent(kamarData, "", "\t")
	// output, _ := json.Marshal(input)

	// Check WriteFile summary - can leave file in partially written state if an error occurs. Check whether os.Create and io.Copy might be a better solution.
	err = os.WriteFile("refresh.json", output, 0644)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.logger.PrintInfo("data successfully received from KAMAR and written to refresh.json", nil)
	app.successResponse(w, r)
}
