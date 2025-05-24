package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mjefferson-whs/listener/internal/data"
	"github.com/tidwall/gjson"
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
	DateTime         int               `json:"datetime,omitempty"`
	FullSync         int               `json:"fullsync,omitempty"`
	InfoURL          string            `json:"infourl,omitempty"`
	PrivacyStatement string            `json:"privacystatement,omitempty"`
	Schools          []School          `json:"schools,omitempty"`
	SMS              string            `json:"sms,omitempty"`
	Sync             string            `json:"sync,omitempty"`
	Version          int               `json:"version,omitempty"`
	Assessments      AssessmentsField  `json:"assessments,omitempty"`
	Attendance       AttendanceField   `json:"attendance,omitempty"`
	ClassEfforts     ClassEffortsField `json:"classefforts,omitempty"`
	Notices          NoticesField      `json:"notices,omitempty"`
	Pastoral         PastoralField     `json:"pastoral,omitempty"`
	Recognitions     RecognitionsField `json:"recognitions,omitempty"`
	Results          ResultsField      `json:"results,omitempty"`
	// Use pointers to the three fields below, to allow easy checking for nil values (rather than 0 values), seeing as they come under a "full" or "part" key rather than their name
	Staff      *StaffField     `json:"staff,omitempty"`
	Students   *StudentsField  `json:"students,omitempty"`
	Subjects   *SubjectsField  `json:"subjects,omitempty"`
	Timetables TimetablesField `json:"timetables,omitempty"`
}

type AssessmentsField struct {
	Count int               `json:"count,omitempty"`
	Data  []data.Assessment `json:"data,omitempty"`
}
type AttendanceField struct {
	Count int               `json:"count,omitempty"`
	Data  []data.Attendance `json:"data,omitempty"`
}
type ClassEffortsField struct {
	Count int                `json:"count,omitempty"`
	Data  []data.ClassEffort `json:"data,omitempty"`
}
type NoticesField struct {
	Count int           `json:"count,omitempty"`
	Data  []data.Notice `json:"data,omitempty"`
}
type PastoralField struct {
	Count int             `json:"count,omitempty"`
	Data  []data.Pastoral `json:"data,omitempty"`
}
type RecognitionsField struct {
	Count int                `json:"count,omitempty"`
	Data  []data.Recognition `json:"data,omitempty"`
}
type ResultsField struct {
	Count int           `json:"count,omitempty"`
	Data  []data.Result `json:"data,omitempty"`
}
type StaffField struct {
	Count int          `json:"count,omitempty"`
	Data  []data.Staff `json:"data,omitempty"`
}
type StudentsField struct {
	Count int            `json:"count,omitempty"`
	Data  []data.Student `json:"data,omitempty"`
}
type SubjectsField struct {
	Count int            `json:"count,omitempty"`
	Data  []data.Subject `json:"data,omitempty"`
}
type TimetablesField struct {
	Count int              `json:"count,omitempty"`
	Data  []data.Timetable `json:"data,omitempty"`
}

func (app *application) kamarRefreshHandler(c echo.Context) error {
	// If write_to_json is on, send context to json writing handler instead
	if app.config.kamar_write_to_json {
		return app.kamarRefreshJSONHandler(c)
	} else {
		var kamarData KAMARData

		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			app.logger.PrintError(errors.New("listener: failed to read data from directory service"), map[string]any{
				// "request body": c.Request().Body,
				"error": err.Error(),
			})
			return app.kamarUnprocessableEntityResponse(c)
		}

		err = json.Unmarshal(body, &kamarData)

		// err := c.Bind(&kamarData)

		// Response if KAMAR JSON is malformed/incomplete
		if err != nil {
			app.logger.PrintError(errors.New("listener: failed to bind data from KAMAR to Go structs"), map[string]any{
				// "request body": c.Request().Body,
				"error": err.Error(),
			})
			return app.kamarUnprocessableEntityResponse(c)
		}

		app.logRequest(c, fmt.Sprintf("hit kamarRefreshHandler, with syncType: %v", kamarData.Data.Sync))

		// Establish the type of request from KAMAR
		syncType := kamarData.Data.Sync
		if syncType == "" {
			app.logger.PrintError(errors.New("listener: failed to get syncType from input"), map[string]any{
				"data": kamarData.Data,
			})
			return app.kamarUnprocessableEntityResponse(c)
		}

		count := 0

		// Check sync type, and respond accordingly
		switch syncType {
		// "check" requests are sent when service is first set up/reestablished, and once a day between 4am and 5am, to verify that the service is up and what fields it is listening for
		case "check":
			app.logger.PrintInfo("listener: received and processed check request", map[string]any{
				"data": kamarData.Data,
			})
			return app.kamarCheckResponse(c)
		case "assessments":
			count = kamarData.Data.Assessments.Count
			app.logger.PrintInfo("listener: attempting to write assessments to database...", map[string]any{
				"count": count,
				"sync":  syncType,
			})
			err = app.models.Assessments.InsertManyAssessments(kamarData.Data.Assessments.Data)
		case "attendance":
			count = kamarData.Data.Attendance.Count
			app.logger.PrintInfo("listener: attempting to write attendance to database...", map[string]any{
				"count": count,
				"sync":  syncType,
			})
			err = app.models.Attendance.InsertManyAttendance(kamarData.Data.Attendance.Data)
		case "classefforts":
			count = kamarData.Data.ClassEfforts.Count
			app.logger.PrintInfo("listener: attempting to write classefforts to database...", map[string]any{
				"count": count,
				"sync":  syncType,
			})
			err = app.models.ClassEfforts.InsertManyClassEfforts(kamarData.Data.ClassEfforts.Data)
		case "full", "part":
			// TODO: Issue here where there may be multiple errors but only the last one will be returned/recorded
			if kamarData.Data.Staff != nil {
				count += kamarData.Data.Staff.Count
				app.logger.PrintInfo("listener: attempting to write staff to database...", map[string]any{
					"count": kamarData.Data.Staff.Count,
					"sync":  syncType,
				})
				err = app.models.Staff.InsertManyStaff(kamarData.Data.Staff.Data)
			}
			if kamarData.Data.Students != nil {
				count += kamarData.Data.Students.Count
				app.logger.PrintInfo("listener: attempting to write students to database...", map[string]any{
					"count": kamarData.Data.Students.Count,
					"sync":  syncType,
				})
				err = app.models.Students.InsertManyStudents(kamarData.Data.Students.Data)
			}
			if kamarData.Data.Subjects != nil {
				count += kamarData.Data.Subjects.Count
				app.logger.PrintInfo("listener: attempting to write subjects to database...", map[string]any{
					"count": kamarData.Data.Subjects.Count,
					"sync":  syncType,
				})
				err = app.models.Subjects.InsertManySubjects(kamarData.Data.Subjects.Data)
			}
		case "notices":
			count += kamarData.Data.Notices.Count
			app.logger.PrintInfo("listener: attempting to write notices records to database...", map[string]any{
				"count": count,
				"sync":  syncType,
			})
			err = app.models.Notices.InsertManyNotices(kamarData.Data.Notices.Data)
		case "pastoral":
			count = kamarData.Data.Pastoral.Count
			app.logger.PrintInfo("listener: attempting to write pastoral records to database...", map[string]any{
				"count": count,
				"sync":  syncType,
			})
			err = app.models.Pastoral.InsertManyPastoral(kamarData.Data.Pastoral.Data)
		case "recognitions":
			count = kamarData.Data.Recognitions.Count
			app.logger.PrintInfo("listener: attempting to write recognitions to database...", map[string]any{
				"count": count,
				"sync":  syncType,
			})
			err = app.models.Recognitions.InsertManyRecognitions(kamarData.Data.Recognitions.Data)
		case "results":
			count = kamarData.Data.Results.Count
			app.logger.PrintInfo("listener: attempting to write results to database...", map[string]any{
				"count": count,
				// "data":  kamarData.Data.Results.Data,
				"sync": syncType,
				// "schools": kamarData.Data.Schools,
			})
			err = app.models.Results.InsertManyResults(kamarData.Data.Results.Data)
		case "studenttimetables", "stafftimetables":
			count = kamarData.Data.Timetables.Count
			app.logger.PrintInfo("listener: attempting to write timetables to database...", map[string]any{
				"count": count,
				// "data":  kamarData.Data.Results.Data,
				"sync": syncType,
				// "schools": kamarData.Data.Schools,
			})
			err = app.models.Timetables.InsertManyTimetables(kamarData.Data.Timetables.Data)
		// If synctype doesn't match any of these cases, return an unprocessable entity error
		default:
			app.logger.PrintError(errors.New("listener: synctype not available"), map[string]any{
				"sync": syncType,
			})
			return app.kamarUnprocessableEntityResponse(c)
		}

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
		- recognitions (unavailable in test data)
		- classefforts (unavailable in test data)
		- full/part (json keys="staff", "students")
		- subjects
		- studenttimetables/stafftimetables (json key="timetables")
		*/

		if err != nil {
			app.logError(c, err)
			return app.kamarUnprocessableEntityResponse(c)
		}

		app.logger.PrintInfo("listener: data successfully received from KAMAR and written to the SQLite database", nil)

		e := data.ListenerEvent{
			ReqType:       "insert",
			WasSuccessful: true,
			Message:       "sync type: " + syncType,
		}

		app.appMetrics.SetLastInsertTime(time.Now())
		app.models.ListenerEvents.Insert(&e)
		// TODO: Count represents written AND updated AND ignored - make this a more useful metric
		app.appMetrics.IncreaseRecordCount(count)

		return app.kamarSuccessResponse(c)

		// IMPORTANT: Instead of the results being written to SQLite as above, the following will write to a JSON file in the same directory as the binary - useful for testing?

		// This is a drawn out and unnecessary attempt to get a slice of results to send to the InsertMany function - hopefully solved by defining better data structs at the top of this file
		// resultsInterface, ok := kamarData.Data["results"].(map[string]any)
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
}

func (app *application) kamarRefreshJSONHandler(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		app.logger.PrintError(err, nil)
		return app.kamarUnprocessableEntityResponse(c)
	}

	// Use json decoder to process as little json as possible to find the sync type
	syncType := gjson.GetBytes(body, "SMSDirectoryData.sync").String()

	if syncType == "" {
		app.logger.PrintError(errors.New("couldn't get sync type from raw JSON data"), nil)
		return app.kamarUnprocessableEntityResponse(c)
	}

	if syncType == "check" {
		app.logger.PrintInfo("listener: received and processed check", nil)
		return app.kamarCheckResponse(c)
	}

	outputDir := "./kamar-json"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		app.logger.PrintError(err, nil)
		return app.kamarUnprocessableEntityResponse(c)
	}

	filename := fmt.Sprintf("%s/%s_%s.json", outputDir, syncType, time.Now().Format("20060102_150405_00000"))
	if err := os.WriteFile(filename, body, 0644); err != nil {
		app.logger.PrintError(err, nil)
		return app.kamarUnprocessableEntityResponse(c)
	}

	app.logger.PrintInfo("listener: successfully wrote data to json file", map[string]any{
		"sync type": syncType,
		"file name": filename,
	})

	return app.kamarSuccessResponse(c)
}
