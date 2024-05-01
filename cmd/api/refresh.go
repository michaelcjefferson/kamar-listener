package main

import (
	"errors"
	"net/http"

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
	InforURL         string       `json:"infourl,omitempty"`
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

func (app *application) kamarRefreshHandler(w http.ResponseWriter, r *http.Request) {
	var kamarData KAMARData

	err := app.readJSON(w, r, &kamarData)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	syncType := kamarData.Data.Sync
	if syncType == "" {
		app.logger.PrintError(errors.New("failed to get syncType from input"), map[string]interface{}{
			"data": kamarData.Data,
		})
		app.authFailedResponse(w, r)
		return
	}

	if syncType == "check" {
		app.checkResponse(w, r)
		app.logger.PrintInfo("received and processed check request", map[string]interface{}{
			"data": kamarData.Data,
		})
		return
	}

	app.logger.PrintInfo("attempting to write results to database...", map[string]interface{}{
		"count": kamarData.Data.Results.Count,
		// "data":    kamarData.Data.Results.Data,
		// "schools": kamarData.Data.Schools,
	})
	err = app.models.Results.InsertMany(kamarData.Data.Results.Data)
	if err != nil {
		app.logger.PrintError(err, nil)
		app.authFailedResponse(w, r)
		return
	}

	app.logger.PrintInfo("data successfully received from KAMAR and written to the SQLite database", nil)
	app.successResponse(w, r)

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
