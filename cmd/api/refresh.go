package main

import (
	"encoding/json"
	"net/http"
	"os"
)

func (app *application) kamarRefreshHandler(w http.ResponseWriter, r *http.Request) {
	var input interface{}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.failedResponse(w, r)
		return
	}

	// app.logger.PrintInfo("request received:", map[string]string{
	// 	"Data": fmt.Sprintf("%+v", input),
	// })

	// app.logger.PrintInfo("request received:", map[string]string{
	// 	"Type": fmt.Sprintf("%s", input.SMSDirectoryData.sync),
	// })

	// MarshalIndent should be used for testing as it indents the JSON that it creates, making it more readable. However, Marshal should be used in practise instead because it consumes less resources (especially important due to the size of the JSON file that is created).
	output, _ := json.MarshalIndent(input, "", "\t")
	// output, _ := json.Marshal(input)

	// Check WriteFile summary - can leave file in partially written state if an error occurs. Check whether os.Create and io.Copy might be a better solution.
	err = os.WriteFile("refresh.json", output, 0644)
	if err != nil {
		app.failedResponse(w, r)
		return
	}

	// Make an empty http.header map, and then add headers to meet KAMAR's requirements.
	headers := make(http.Header)
	headers.Set("Server", "WHS KAMAR Refresh/1.0")
	headers.Set("Connection", "close")

	app.successResponse(w, r)
}

// type requestMetadata struct {
// 	TimeStamp          int
// 	RequestContentType string
// 	SMS                string
// 	KAMARVersion       int
// }

// func (app *application) kamarRefreshHandler(w http.ResponseWriter, r *http.Request) {
// 	// Create struct to be populated with known keys in all KAMAR requests. RequestType (sync) will dictate which steps will follow based on the type of request that was sent by KAMAR.
// 	var input struct {
// 		Data struct {
// 			TimeStamp    int    `json:"timestamp"`
// 			RequestType  string `json:"sync"`
// 			SMS          string `json:"sms"`
// 			KAMARVersion int    `json:"version"`
// 		} `json: "SMSDirectoryData"`
// 	}

// 	// Pass request (r) to new JSON decoder in readJSON where it will parse r.Body, and pass reference to &input to be populated by Decode method
// 	err := app.readJSON(w, r, &input)
// 	if err != nil {
// 		app.badRequestResponse(w, r, err)
// 		return
// 	}

// 	app.logger.PrintInfo("request received:", map[string]string{
// 		"Request Type": input.Data.RequestType,
// 		"SMS":          input.Data.SMS,
// 	})
// }
