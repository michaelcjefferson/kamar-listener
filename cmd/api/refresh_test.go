package main

import (
	"bytes"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mjefferson-whs/listener/internal/data"
	"github.com/mjefferson-whs/listener/internal/jsonlog"
)

func setupTestDB(t *testing.T) *sql.DB {
	// Create in-memory database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	// Initialize schema
	_, err = db.Exec(`CREATE TABLE results (
		code			TEXT,
		comment         TEXT,
		course          TEXT,
		curriculumlevel,
		date            TEXT,
		enrolled		INTEGER,
		id              INTEGER,
		nsn             TEXT,
		number          TEXT,
		published		INTEGER,
		result          TEXT,
		resultData TEXT,
		results TEXT,
		subject         TEXT,
		tnv 			TEXT,
		type            TEXT,
		version         INTEGER,
		year            INTEGER,
		yearlevel       INTEGER
	)`)
	if err != nil {
		t.Fatalf("Failed to create schema: %v", err)
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
		expectedBody   string
		checkDB        func(*testing.T, *sql.DB)
	}{
		// TODO: Update to get config data from actual DB? Or include tests.dbSetup func which inserts appropriate config?
		{
			name:           "Valid Check, Valid Credentials",
			jsonFile:       "check.json",
			username:       "username",
			password:       "password",
			includeAuth:    true,
			expectedStatus: http.StatusOK,
			expectedBody: `{
				"SMSDirectoryData": {
					"error":             0,
					"result":            "OK",
					"service":           "WHS KAMAR Refresh",
					"version":           "1.1",
					"status":            "Ready",
					"infourl":           "https://wakatipu.school.nz/",
					"privacystatement":  "This service only collects results data, and stores it locally on a secure device. Only staff members of the school have access to the data.",
					"countryDataStored": "New Zealand",
					"options": {
						"ics": true,
						"students": {
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
							"fields": {
								"required": "firstname;lastname;gender;nsn",
								"optional": "username;caregivers;caregivers1;caregivers2;caregiver.name;caregiver.relationship;caregiver.mobile;caregiver.email",
							}
						},
						"common": {
							"subjects": false,
							"notices":  false,
							"calendar": false,
							"bookings": false,
						}
					}
				}
			}`,
		},
		{
			name:           "Valid Check, Invalid Credentials",
			jsonFile:       "check.json",
			username:       "userframe",
			password:       "bassword",
			includeAuth:    true,
			expectedStatus: http.StatusForbidden,
			expectedBody: `{
				"SMSDirectoryData": {
					"error":   403,
					"result":  "Authentication Failed",
					"service": "WHS KAMAR Refresh",
					"version": "1.0",
				}
			}`,
		},
		{
			name:           "Valid Check, Missing Credentials",
			jsonFile:       "check.json",
			username:       "",
			password:       "",
			includeAuth:    false,
			expectedStatus: http.StatusUnauthorized,
			expectedBody: `{
				"SMSDirectoryData": {
					"error":   401,
					"result":  "No Credentials Provided",
					"service": "WHS KAMAR Refresh",
					"version": "1.0",
				}
			}`,
		},
		{
			name:           "Malformed Check, Valid Credentials",
			jsonFile:       "malformed.json",
			username:       "username",
			password:       "password",
			includeAuth:    true,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody: `{
				"SMSDirectoryData": {
					"error":   422,
					"result":  "Request From KAMAR Was Malformed",
					"service": "WHS KAMAR Refresh",
					"version": "1.0",
				}
			}`,
		},
		{
			name:           "Valid Results Data",
			jsonFile:       "refresh-test.json",
			username:       "username",
			password:       "password",
			includeAuth:    true,
			expectedStatus: http.StatusOK,
			expectedBody: `{
				"SMSDirectoryData": {
					"error":   0,
					"result":  "OK",
					"service": "WHS KAMAR Refresh",
					"version": "1.0",
				}
			}`,
		},
		{
			name:           "Valid Results Data, Invalid Credentials",
			jsonFile:       "refresh-test.json",
			username:       "",
			password:       "password",
			includeAuth:    true,
			expectedStatus: http.StatusForbidden,
			expectedBody: `{
				"SMSDirectoryData": {
					"error":   403,
					"result":  "Authentication Failed",
					"service": "WHS KAMAR Refresh",
					"version": "1.0",
				}
			}`,
		},
		{
			name:           "Malformed Results Data",
			jsonFile:       "refresh-test.json",
			username:       "username",
			password:       "password",
			includeAuth:    true,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody: `{
				"SMSDirectoryData": {
					"error":   422,
					"result":  "Request From KAMAR Was Malformed",
					"service": "WHS KAMAR Refresh",
					"version": "1.0",
				}
			}`,
		},
		// {
		// 	name:     "Results Data with Incorrect (attendance) Sync Label",
		// 	jsonFile: "refresh-test.json",
		// },
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
			// TODO: Middleware currently checks against username and password individually, so set them here
			app.config.credentials.full = "username:password"
			logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo, nil)
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
			if err != nil {
				t.Fatal(err)
			}

			if status := rec.Code; status != tt.expectedStatus {
				t.Errorf("handler returned unexpected status code: got %v want %v", status, tt.expectedStatus)
			}

			if tt.expectedBody != "" {
				responseBody := bytes.TrimSpace(rec.Body.Bytes())
				if string(responseBody) != tt.expectedBody {
					t.Errorf("handler returned unexpected body: got %v want %v", string(responseBody), tt.expectedBody)
				}
			}

			if tt.checkDB != nil {
				tt.checkDB(t, db)
			}
		})
	}
}
