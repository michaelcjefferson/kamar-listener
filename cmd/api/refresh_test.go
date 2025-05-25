package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/michaelcjefferson/kamar-listener/internal/assert"
	"github.com/michaelcjefferson/kamar-listener/internal/data"
	"github.com/michaelcjefferson/kamar-listener/internal/jsonlog"
)

func setupTestDB(t *testing.T) (*sql.DB, *sql.DB) {
	// Create in-memory database for testing
	listenerDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	err = createSMSTables(listenerDB)
	if err != nil {
		t.Fatalf("Failed to create SMS tables in database: %v", err)
	}

	appDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	err = createConfigTable(appDB)
	if err != nil {
		t.Fatalf("Failed to create SMS tables in database: %v", err)
	}

	p := data.Password{}
	p.Set("password")
	h := p.Hash()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	appDB.ExecContext(ctx, `UPDATE config
	SET value = "username"
	WHERE key = "listener_username";`)

	appDB.ExecContext(ctx, `UPDATE config
	SET value = ?
	WHERE key = "listener_password";`, h)

	return listenerDB, appDB
}

// TODO: Remove *sql.DB from checkDB function and run everything from app.models - this requires creating a count() query for each model
func TestRefreshHandler(t *testing.T) {
	tests := []struct {
		name           string
		jsonFile       string
		setupFunc      func(*sql.DB, *sql.DB) // Optional function to set up initial data
		username       string
		password       string
		includeAuth    bool
		expectedStatus int
		expectedBody   map[string]any
		expectedCount  int
		checkDB        func(*testing.T, int, *sql.DB, *application)
	}{
		// TODO: Update to get config data from actual DB? Or include tests.dbSetup func which inserts appropriate config?
		// TODO: Use setupFunc and checkDB functions to build more varied tests
		{
			name:           "Valid Check, Valid Credentials",
			jsonFile:       "check.json",
			username:       "username",
			password:       "password",
			includeAuth:    true,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]any{
				"SMSDirectoryData": map[string]any{
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
							"passwords":       true,
							"photos":          false,
							"groups":          true,
							"awards":          true,
							"timetables":      true,
							"attendance":      true,
							"assessments":     true,
							"pastoral":        true,
							"recognitions":    true,
							"classefforts":    true,
							"learningsupport": true,
							"fields": map[string]string{
								"required": "firstname;lastname;gender;gendercode;nsn;uniqueid",
								"optional": "schoolindex;firstnamelegal;lastnamelegal;forenames;forenameslegal;genderpreferred;username;mobile;email;house;whanau;boarder;byodinfo;ece;esol;ors;languagespoken;datebirth;startingdate;startschooldate;created;leavingdate;leavingreason;leavingschool;leavingactivity;res;resa;resb;res.title;res.salutation;res.email;res.numFlatUnit;res.numStreet;res.ruralDelivery;res.suburb;res.town;res.postcode;caregivers;caregivers1;caregivers2;caregivers3;caregivers4;caregiver.name;caregiver.relationship;caregiver.status;caregiver.address,caregiver.mobile;caregiver.email;emergency;emergency1;emergency2;emergency.name;emergency.relationship;emergency.mobile;moetype;ethnicityL1;ethnicityL2;ethnicity;iwi;yearlevel;fundinglevel;tutor;timetablebottom1;timetablebottom2;timetablebottom3;timetablebottom4;timetabletop1;timetabletop2;timetabletop3;timetabletop4;maorilevel;pacificlanguage;pacificlevel;flags;flag.alert;flag.conditions;flag.dietary;flag.general;flag.ibuprofen;flag.medical;flag.notes;flag.paracetamol;flag.pastoral;flag.reactions;flag.specialneeds;flag.vaccinations;flag.eotcconsent;flag.eotcform;custom;custom.custom1;custom.custom2;custom.custom3;custom.custom4;custom.custom5;siblinglink;photocopierid;signedagreement;accountdisabled;networkaccess;altdescription;althomedrive",
							},
						},
						"staff": map[string]any{
							"details":    true,
							"photos":     false,
							"timetables": true,
							"fields": map[string]string{
								"required": "uniqueid;firstname;lastname;username;gender;email",
								"optional": "schoolindex;title;mobile;extension;classification;position;house;tutor;groups;groups.departments;datebirth;created;leavingdate;startingdate;eslguid;moenumber;photocopierid;registrationnumber;custom;custom.custom1;custom.custom2;custom.custom3;custom.custom4;custom.custom5",
							},
						},
						"common": map[string]bool{
							"subjects": true,
							"notices":  false,
							"calendar": false,
							"bookings": false,
						},
					},
				},
			},
		},
		{
			name:           "Valid Check, Invalid Credentials",
			jsonFile:       "check.json",
			username:       "userframe",
			password:       "bassword",
			includeAuth:    true,
			expectedStatus: http.StatusForbidden,
			expectedBody: map[string]any{
				"SMSDirectoryData": map[string]any{
					"error":   403,
					"result":  "Authentication Failed",
					"service": "WHS KAMAR Refresh",
					"version": "1.0",
				},
			},
		},
		{
			name:           "Valid Check, Missing Credentials",
			jsonFile:       "check.json",
			username:       "",
			password:       "",
			includeAuth:    false,
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]any{
				"SMSDirectoryData": map[string]any{
					"error":   401,
					"result":  "No Credentials Provided",
					"service": "WHS KAMAR Refresh",
					"version": "1.0",
				},
			},
		},
		{
			name:           "Malformed Check, Valid Credentials",
			jsonFile:       "malformed.json",
			username:       "username",
			password:       "password",
			includeAuth:    true,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody: map[string]any{
				"SMSDirectoryData": map[string]any{
					"error":   422,
					"result":  "Request From KAMAR Was Malformed",
					"service": "WHS KAMAR Refresh",
					"version": "1.0",
				},
			},
		},
		{
			name:           "Valid Results Data",
			jsonFile:       "refresh-test.json",
			username:       "username",
			password:       "password",
			includeAuth:    true,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]any{
				"SMSDirectoryData": map[string]any{
					"error":   0,
					"result":  "OK",
					"service": "WHS KAMAR Refresh",
					"version": "1.0",
				},
			},
			expectedCount: 5,
			checkDB: func(t *testing.T, expectedCount int, db *sql.DB, app *application) {
				var actualCount int

				err := db.QueryRow("SELECT COUNT(*) FROM results;").Scan(&actualCount)
				if err != nil {
					t.Fatalf("error getting count of results from db: %v", err)
				}

				if actualCount != expectedCount {
					t.Errorf("unexpected number of results inserted into database: want %d got %d", expectedCount, actualCount)
				}
			},
		},
		{
			name:           "Valid Results Data, Invalid Credentials",
			jsonFile:       "refresh-test.json",
			username:       "",
			password:       "password",
			includeAuth:    true,
			expectedStatus: http.StatusForbidden,
			expectedBody: map[string]any{
				"SMSDirectoryData": map[string]any{
					"error":   403,
					"result":  "Authentication Failed",
					"service": "WHS KAMAR Refresh",
					"version": "1.0",
				},
			},
		},
		{
			name:           "Malformed Results Data",
			jsonFile:       "malformed-refresh-test.json",
			username:       "username",
			password:       "password",
			includeAuth:    true,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody: map[string]any{
				"SMSDirectoryData": map[string]any{
					"error":   422,
					"result":  "Request From KAMAR Was Malformed",
					"service": "WHS KAMAR Refresh",
					"version": "1.0",
				},
			},
		},
		// {
		// 	name:     "Results Data with Incorrect (attendance) Sync Label",
		// 	jsonFile: "refresh-test.json",
		// },
		{
			name:           "Valid Assessments Data",
			jsonFile:       "actual-requests/assessments_18122024_161942.json",
			username:       "username",
			password:       "password",
			includeAuth:    true,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]any{
				"SMSDirectoryData": map[string]any{
					"error":   0,
					"result":  "OK",
					"service": "WHS KAMAR Refresh",
					"version": "1.0",
				},
			},
			// TODO: Actual count is 2365. Assessment titles T_TP and T_UE, with _ replaced with numbers 1 through 4, have duplicate entries, so any write after the first is skipped
			expectedCount: 1969,
			checkDB: func(t *testing.T, expectedCount int, db *sql.DB, app *application) {
				var actualCount int

				err := db.QueryRow("SELECT COUNT(*) FROM assessments;").Scan(&actualCount)
				if err != nil {
					t.Fatalf("error getting count of assessments from db: %v", err)
				}

				if actualCount != expectedCount {
					t.Errorf("unexpected number of assessments inserted into database: want %d got %d", expectedCount, actualCount)
				}

				ass, err := app.models.Assessments.GetByAssessmentNumber("91402")
				if err != nil {
					t.Fatalf("error getting assessment from db: %v", err)
				}

				// TODO: Move to test table - optional data verification check, with two test table parameters - key (eg. assmnt number for select query) and expected returned struct values
				if *ass.Type != "A" {
					t.Errorf("unexpected value in db for assessment.type: want %v got %v", "A", ass.Type)
				}
				if *ass.Number != "91402" {
					t.Errorf("unexpected value in db for assessment.number: want %v got %v", "91402", ass.Number)
				}
				if *ass.Version != 3 {
					t.Errorf("unexpected value in db for assessment.version: want %v got %v", 3, ass.Version)
				}
				if *ass.Level != 3 {
					t.Errorf("unexpected value in db for assessment.level: want %v got %v", 3, ass.Level)
				}
				if *ass.Credits != 5 {
					t.Errorf("unexpected value in db for assessment.credits: want %v got %v", 5, ass.Credits)
				}
				if ass.Weighting != nil {
					t.Errorf("unexpected value in db for assessment.weighting: want %v got %v", nil, ass.Weighting)
				}
				if ass.Points != nil {
					t.Errorf("unexpected value in db for assessment.points: want %v got %v", nil, ass.Points)
				}
				if *ass.Title != "Economics 3.4 - Demonstrate understanding of government interventions where the market fails to deliver efficient or equitable outcomes" {
					t.Errorf("unexpected value in db for assessment.title: want %v got %v", "Economics 3.4 - Demonstrate understanding of government interventions where the market fails to deliver efficient or equitable outcomes", ass.Title)
				}
				if ass.Description != nil {
					t.Errorf("unexpected value in db for assessment.description: want %v got %v", "", ass.Description)
				}
				if ass.Purpose != nil {
					t.Errorf("unexpected value in db for assessment.purpose: want %v got %v", "", ass.Purpose)
				}
				if *ass.Subfield != "Economic Theory and Practice" {
					t.Errorf("unexpected value in db for assessment.subfield: want %v got %v", "Economic Theory and Practice", ass.Subfield)
				}
				if *ass.Internalexternal != "I" {
					t.Errorf("unexpected value in db for assessment.internalexternal: want %v got %v", "I", ass.Internalexternal)
				}

				// "type": "A",
				// "number": "91402",
				// "version": 3,
				// "level": 3,
				// "credits": 5,
				// "weighting": null,
				// "points": null,
				// "title": "Economics 3.4 - Demonstrate understanding of government interventions where the market fails to deliver efficient or equitable outcomes",
				// "description": null,
				// "purpose": null,
				// "subfield": "Economic Theory and Practice",
				// "internalexternal": "I"
			},
		},
		{
			name:           "Valid Attendance Data",
			jsonFile:       "actual-requests/attendance_18122024_152843.json",
			username:       "username",
			password:       "password",
			includeAuth:    true,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]any{
				"SMSDirectoryData": map[string]any{
					"error":   0,
					"result":  "OK",
					"service": "WHS KAMAR Refresh",
					"version": "1.0",
				},
			},
			// TODO: This only represents the total number of unique student IDs added from this attendance request, and doesn't match the "count" value from the request - "count" is 3982 (total number of attendance records), and attendance_values will have a total count much higher than this, as each record contains up to 5 values
			expectedCount: 1471,
			checkDB: func(t *testing.T, expectedCount int, db *sql.DB, app *application) {
				var actualCount int

				err := db.QueryRow("SELECT COUNT(*) FROM attendance;").Scan(&actualCount)
				if err != nil {
					t.Fatalf("error getting count of attendance from db: %v", err)
				}

				if actualCount != expectedCount {
					t.Errorf("unexpected number of attendance records inserted into database: want %d got %d", expectedCount, actualCount)
				}
			},
		},
		{
			name:           "Valid Pastoral Data",
			jsonFile:       "actual-requests/pastoral_19122024_154553.json",
			username:       "username",
			password:       "password",
			includeAuth:    true,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]any{
				"SMSDirectoryData": map[string]any{
					"error":   0,
					"result":  "OK",
					"service": "WHS KAMAR Refresh",
					"version": "1.0",
				},
			},
			expectedCount: 3262,
			checkDB: func(t *testing.T, expectedCount int, db *sql.DB, app *application) {
				var actualCount int

				err := db.QueryRow("SELECT COUNT(*) FROM pastoral;").Scan(&actualCount)
				if err != nil {
					t.Fatalf("error getting count of pastoral from db: %v", err)
				}

				if actualCount != expectedCount {
					t.Errorf("unexpected number of pastoral records inserted into database: want %d got %d", expectedCount, actualCount)
				}
			},
		},
		{
			name:           "Valid Student Details (Full) Data",
			jsonFile:       "actual-requests/full_18122024_151311.json",
			username:       "username",
			password:       "password",
			includeAuth:    true,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]any{
				"SMSDirectoryData": map[string]any{
					"error":   0,
					"result":  "OK",
					"service": "WHS KAMAR Refresh",
					"version": "1.0",
				},
			},
			expectedCount: 1777,
			checkDB: func(t *testing.T, expectedCount int, db *sql.DB, app *application) {
				var actualCount int

				err := db.QueryRow("SELECT COUNT(*) FROM students;").Scan(&actualCount)
				if err != nil {
					t.Fatalf("error getting count of students from db: %v", err)
				}

				if actualCount != expectedCount {
					t.Errorf("unexpected number of students records inserted into database: want %d got %d", expectedCount, actualCount)
				}
			},
		},
		{
			name:           "Valid Student Details (Part) Data",
			jsonFile:       "actual-requests/part_18122024_201702.json",
			username:       "username",
			password:       "password",
			includeAuth:    true,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]any{
				"SMSDirectoryData": map[string]any{
					"error":   0,
					"result":  "OK",
					"service": "WHS KAMAR Refresh",
					"version": "1.0",
				},
			},
			expectedCount: 1,
			checkDB: func(t *testing.T, expectedCount int, db *sql.DB, app *application) {
				var actualCount int

				err := db.QueryRow("SELECT COUNT(*) FROM students;").Scan(&actualCount)
				if err != nil {
					t.Fatalf("error getting count of students from db: %v", err)
				}

				if actualCount != expectedCount {
					t.Errorf("unexpected number of students records inserted into database: want %d got %d", expectedCount, actualCount)
				}
			},
		},
		{
			name:           "Valid Staff, Student, and Subject Details (Full) Data",
			jsonFile:       "actual-requests/full_19122024_154006.json",
			username:       "username",
			password:       "password",
			includeAuth:    true,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]any{
				"SMSDirectoryData": map[string]any{
					"error":   0,
					"result":  "OK",
					"service": "WHS KAMAR Refresh",
					"version": "1.0",
				},
			},
			expectedCount: 1,
			checkDB: func(t *testing.T, expectedCount int, db *sql.DB, app *application) {
				var actualStaffCount, actualStudentCount, actualSubjectCount int

				err := db.QueryRow("SELECT COUNT(*) FROM staff;").Scan(&actualStaffCount)
				if err != nil {
					t.Fatalf("error getting count of staff from db: %v", err)
				}

				if actualStaffCount != 227 {
					t.Errorf("unexpected number of staff records inserted into database: want %d got %d", 227, actualStaffCount)
				}

				err = db.QueryRow("SELECT COUNT(*) FROM students;").Scan(&actualStudentCount)
				if err != nil {
					t.Fatalf("error getting count of students from db: %v", err)
				}

				if actualStudentCount != 1777 {
					t.Errorf("unexpected number of students records inserted into database: want %d got %d", 1777, actualStudentCount)
				}

				err = db.QueryRow("SELECT COUNT(*) FROM subjects;").Scan(&actualSubjectCount)
				if err != nil {
					t.Fatalf("error getting count of subjects from db: %v", err)
				}

				if actualSubjectCount != 255 {
					t.Errorf("unexpected number of subject records inserted into database: want %d got %d", 255, actualSubjectCount)
				}
			},
		},
		{
			name:           "Valid Student Timetables Data",
			jsonFile:       "actual-requests/studenttimetables_18122024_152357.json",
			username:       "username",
			password:       "password",
			includeAuth:    true,
			expectedStatus: http.StatusOK,
			expectedBody: map[string]any{
				"SMSDirectoryData": map[string]any{
					"error":   0,
					"result":  "OK",
					"service": "WHS KAMAR Refresh",
					"version": "1.0",
				},
			},
			expectedCount: 8,
			checkDB: func(t *testing.T, expectedCount int, db *sql.DB, app *application) {
				var actualCount int

				err := db.QueryRow("SELECT COUNT(*) FROM timetables;").Scan(&actualCount)
				if err != nil {
					t.Fatalf("error getting count of timetables from db: %v", err)
				}

				if actualCount != expectedCount {
					t.Errorf("unexpected number of timetables inserted into database: want %d got %d", expectedCount, actualCount)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var listenerDB, appDB *sql.DB

			listenerDB, appDB = setupTestDB(t)
			defer listenerDB.Close()
			defer appDB.Close()

			if tt.setupFunc != nil {
				tt.setupFunc(listenerDB, appDB)
			}

			app := &application{}
			app.models = data.NewModels(appDB, listenerDB, app.background)

			// logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo, nil)
			// io.Discard means logs won't be written to terminal when tests are run - replace with the line above to see logs during testing
			logger := jsonlog.New(io.Discard, jsonlog.LevelInfo, nil)
			app.logger = logger

			jsonPath := filepath.Join("../../test", tt.jsonFile)
			jsonData, err := os.ReadFile(jsonPath)
			if err != nil {
				t.Fatalf("Failed to read test data from %s: %v", jsonPath, err)
			}

			e := echo.New()

			req := httptest.NewRequest(http.MethodPost, "/kamar-refresh", bytes.NewBuffer(jsonData))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			if tt.includeAuth {
				req.SetBasicAuth(tt.username, tt.password)
			}

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// handler := echo.HandlerFunc(app.authenticateKAMAR(app.kamarRefreshHandler))
			// handler.ServeHTTP(rec, req)

			handler := app.authenticateKAMAR(app.kamarRefreshHandler)
			err = handler(c)
			assert.NilError(t, err)

			if status := rec.Code; status != tt.expectedStatus {
				t.Errorf("handler returned unexpected status code: got %v want %v", status, tt.expectedStatus)
			}

			var responseBody map[string]any
			err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
			if err != nil {
				t.Fatalf("error unmarshalling json response: %v", err)
			}

			// Convert tt.expectedBody to JSON and back, to ensure datatypes within the map reflect the ones that will be unmarshalled from the JSON response
			expectedBody, err := NormaliseJSONMapTypes(tt.expectedBody)
			if err != nil {
				t.Fatalf("error normalising expectedBody: %v", err)
			}

			if !reflect.DeepEqual(responseBody, expectedBody) {
				t.Errorf("handler returned unexpected body: got %v want %v", responseBody, expectedBody)
			}

			// responseBody := bytes.TrimSpace(rec.Body.Bytes())
			// if string(responseBody) != tt.expectedBody {
			// 	t.Errorf("handler returned unexpected body: got %v want %v", string(responseBody), tt.expectedBody)
			// }

			// Log the names of each table that exists in the database
			// rows, err := db.Query(`SELECT name FROM sqlite_master WHERE type='table' ORDER BY name;`)
			// if err != nil {
			// 	log.Fatal(err)
			// }
			// defer rows.Close()

			// for rows.Next() {
			// 	var name string
			// 	if err := rows.Scan(&name); err != nil {
			// 		log.Fatal(err)
			// 	}
			// 	fmt.Println(name)
			// }
			// if err := rows.Err(); err != nil {
			// 	log.Fatal(err)
			// }

			if tt.checkDB != nil {
				tt.checkDB(t, tt.expectedCount, listenerDB, app)
			}
		})
	}
}
