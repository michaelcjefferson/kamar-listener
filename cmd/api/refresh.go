package main

import (
	"errors"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/mjefferson-whs/listener/internal/data"
)

type School struct {
	Authoritative bool   `json:"authoritative,omitempty"`
	Index         int    `json:"index,omitempty"`
	MOECode       string `json:"moecode,omitempty"`
	Name          string `json:"name,omitempty"`
	SchoolIndex   int    `json:"schoolindex,omitempty"`
	Type          int    `json:"type,omitempty"`
}

type KAMARData struct {
	Data SMSDirectoryData `json:"SMSDirectoryData"`
}

type SMSDirectoryData struct {
	DateTime         int          `json:"datetime,omitempty"`
	FullSync         int          `json:"fullsync,omitempty"`
	InfoURL          string       `json:"infourl,omitempty"`
	Results          ResultsField `json:"results,omitempty"`
	PrivacyStatement string       `json:"privacystatement,omitempty"`
	Schools          []School     `json:"schools,omitempty"`
	SMS              string       `json:"sms,omitempty"`
	Sync             string       `json:"sync,omitempty"`
	Version          int          `json:"version,omitempty"`
}

type ResultsField struct {
	Count int           `json:"count,omitempty"`
	Data  []data.Result `json:"data,omitempty"`
}

func (app *application) kamarRefreshHandler(c echo.Context) error {
	var kamarData KAMARData

	// Response if KAMAR JSON is malformed/incomplete
	err := c.Bind(&kamarData)
	// err := app.readJSON(c, &kamarData)
	if err != nil {
		app.logger.PrintError(errors.New("listener: failed to bind data from KAMAR to Go structs"), map[string]interface{}{
			"request body": c.Request().Body,
		})
		return app.badRequestResponse(c, err)
	}

	// Establish the type of request from KAMAR
	syncType := kamarData.Data.Sync
	if syncType == "" {
		app.logger.PrintError(errors.New("listener: failed to get syncType from input"), map[string]interface{}{
			"data": kamarData.Data,
		})
		return app.authFailedResponse(c)
	}

	// "check" requests are sent when service is first set up/reestablished, and once a day between 4am and 5am, to verify that the service is up and what fields it is listening for
	if syncType == "check" {
		app.logger.PrintInfo("listener: received and processed check request", map[string]interface{}{
			"data": kamarData.Data,
		})
		return app.checkResponse(c)
	}

	// If syncType is populated, but its value is not "check", then it contains data. syncType will indicate the type of data included
	app.logger.PrintInfo("listener: attempting to write results to database...", map[string]interface{}{
		"count": kamarData.Data.Results.Count,
		// "data":  kamarData.Data.Results.Data,
		"sync": syncType,
		// "schools": kamarData.Data.Schools,
	})

	// TODO: either check the value of "sync", or check for the existence of various keys in the struct, to decide which database queries to run
	// ? Seeing as the struct will omit empty fields in the JSON file, it may be more efficient to just write everything that exists to the DB without checking which fields have been received
	/* "sync" contains the type of data that the message includes.
	it also reflects which keys exist in the SMSDirectoryData JSON file.

	"sync" types:
	- assessments
	- results
	- attendance
	- bookings
	- calendar (json key="calendars")
	- notices
	- pastoral
	- photos/staffphotos (json keys="photos", "staffphotos")
	- full/part (json keys="staff", "students")
	- subjects
	- studenttimetables/stafftimetables (json key="timetables")
	*/
	switch syncType {
	case "results":
		err = app.models.Results.InsertManyResults(kamarData.Data.Results.Data)
	default:
		err = fmt.Errorf("sync type wasn't recognised. sync type: %s", syncType)
	}
	if err != nil {
		app.logError(c, err)
		return app.authFailedResponse(c)
	}

	app.logger.PrintInfo("listener: data successfully received from KAMAR and written to the SQLite database", nil)
	return app.successResponse(c)

	// IMPORTANT: Instead of the results being written to SQLite as above, the following will write to a JSON file in the same directory as the binary - useful for testing?

	// This is a drawn out and unnecessary attempt to get a slice of results to send to the InsertMany function - hopefully solved by defining better data structs at the top of this file
	// resultsInterface, ok := kamarData.Data["results"].(map[string]interface{})
	// if !ok {
	// 	app.logger.PrintError(errors.New("no 'results' field found in JSON"), nil)
	// 	app.authFailedResponse(w, r)
	// 	return
	// }

	// dataInterface, ok := resultsInterface["data"]
	// if !ok {
	// 	app.logger.PrintError(errors.New("no 'data' field found in JSON"), nil)
	// 	app.authFailedResponse(w, r)
	// 	return
	// }

	// var results []data.Result
	// switch dataInterface := dataInterface.(type) {
	// case []any:
	// 	for _, resultInterface := range dataInterface {
	// 		resultMap, ok := resultInterface.(map[string]any)
	// 		if !ok {
	// 			app.logger.PrintError(errors.New("invalid result format"), nil)
	// 			app.authFailedResponse(w, r)
	// 			return
	// 		}
	// 		var result data.Result
	// 		// Any data converted in the following 18 lines, eg. resultMap["comment"].(string), should be error-checked in the same way
	// 		result.Code = resultMap["code"]
	// 		result.Comment, ok = resultMap["comment"].(string)
	// 		if !ok {
	// 			result.Comment = ""
	// 			app.logger.PrintError(errors.New("the 'comment' field must be a string"), nil)
	// 		}
	// 		result.Course, ok = resultMap["course"].(string)
	// 		if !ok {
	// 			result.Course = ""
	// 		}
	// 		result.CurriculumLevel = resultMap["curriculumlevel"]
	// 		result.Date = resultMap["date"].(string)
	// 		result.Enrolled = resultMap["enrolled"]
	// 		result.ID = resultMap["id"].(int)
	// 		result.NSN = resultMap["nsn"].(string)
	// 		result.Number = resultMap["number"].(string)
	// 		result.Published = resultMap["published"]
	// 		result.Result = resultMap["result"].(string)
	// 		result.ResultData = resultMap["resultData"]
	// 		result.Results = resultMap["results"]
	// 		result.Subject = resultMap["subject"].(string)
	// 		result.Type = resultMap["type"].(string)
	// 		result.Version = resultMap["version"].(int)
	// 		result.Year = resultMap["year"].(int)
	// 		result.YearLevel = resultMap["yearlevel"].(int)
	// 		results = append(results, result)
	// 	}
	// default:
	// 	app.logger.PrintError(errors.New("'data' field must be an array"), nil)
	// 	app.authFailedResponse(w, r)
	// }

	// MarshalIndent should be used for testing as it indents the JSON that it creates, making it more readable. However, Marshal should be used in practise instead because it consumes less resources (especially important due to the size of the JSON file that is created).
	// output, _ := json.MarshalIndent(kamarData, "", "\t")
	// Use this one instead to write less-readable but more efficient JSON
	// output, _ := json.Marshal(kamarData)

	// Check WriteFile summary - can leave file in partially written state if an error occurs. Check whether os.Create and io.Copy might be a better solution.
	// err = os.WriteFile("refresh.json", output, 0644)
	// if err != nil {
	// 	app.serverErrorResponse(w, r, err)
	// 	return
	// }
	// app.logger.PrintInfo("data successfully received from KAMAR and written to refresh.json", nil)
	// app.successResponse(w, r)
}
