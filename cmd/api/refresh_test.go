package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mjefferson-whs/listener/internal/assert"
	"github.com/mjefferson-whs/listener/internal/data"
	"github.com/mjefferson-whs/listener/internal/jsonlog"
)

func setupTestDB(t *testing.T) *sql.DB {
	// Create in-memory database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	err = createSMSTables(db)
	if err != nil {
		t.Fatalf("Failed to create SMS tables in database: %v", err)
	}

	return db
}

func TestRefreshHandler(t *testing.T) {
	tests := []struct {
		name           string
		jsonFile       string
		setupFunc      func(*sql.DB) // Optional function to set up initial data
		username       string
		password       string
		includeAuth    bool
		expectedStatus int
		expectedBody   map[string]any
		expectedCount  int
		checkDB        func(*testing.T, int, *sql.DB)
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
							"passwords":       false,
							"photos":          false,
							"groups":          false,
							"awards":          false,
							"timetables":      false,
							"attendance":      false,
							"assessments":     true,
							"pastoral":        false,
							"learningsupport": false,
							"fields": map[string]any{
								"required": "firstname;lastname;gender;nsn",
								"optional": "username;caregivers;caregivers1;caregivers2;caregiver.name;caregiver.relationship;caregiver.mobile;caregiver.email",
							},
						},
						"common": map[string]any{
							"subjects": false,
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
			checkDB: func(t *testing.T, expectedCount int, db *sql.DB) {
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
			expectedCount: 2365,
			checkDB: func(t *testing.T, expectedCount int, db *sql.DB) {
				var actualCount int

				err := db.QueryRow("SELECT COUNT(*) FROM assessments;").Scan(&actualCount)
				if err != nil {
					t.Fatalf("error getting count of assessments from db: %v", err)
				}

				if actualCount != expectedCount {
					t.Errorf("unexpected number of assessments inserted into database: want %d got %d", expectedCount, actualCount)
				}
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
			expectedCount: 3982,
			checkDB: func(t *testing.T, expectedCount int, db *sql.DB) {
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
			checkDB: func(t *testing.T, expectedCount int, db *sql.DB) {
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
			checkDB: func(t *testing.T, expectedCount int, db *sql.DB) {
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
			var db *sql.DB

			db = setupTestDB(t)
			defer db.Close()

			if tt.setupFunc != nil {
				tt.setupFunc(db)
			}

			app := &application{}
			app.models = data.NewModels(db, app.background)

			// Set credentials to check against for basic auth
			// TODO: Middleware currently checks against username and password individually - for now set them here, but in future either make use of or remove credentials.full
			app.config.credentials.username = "username"
			app.config.credentials.password = "password"
			app.config.credentials.full = "username:password"
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

			if tt.checkDB != nil {
				tt.checkDB(t, tt.expectedCount, db)
			}
		})
	}
}
