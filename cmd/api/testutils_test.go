package main

import "testing"

func newTestApplication(t *testing.T) *application {
	cfg := config{
		port:        8085,
		env:         "development",
		dblogs_on:   false,
		app_db_path: "../../test/test-full.db",
		https_on:    true,
		limiter: struct {
			rps     float64
			burst   int
			enabled bool
		}{
			rps:     4,
			burst:   8,
			enabled: false,
		},
	}

	return &application{
		assetHandler: nil,
		config:       cfg,
	}
}
