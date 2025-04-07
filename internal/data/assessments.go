package data

import (
	"database/sql"
	"strconv"
	"strings"
)

type Assessment struct {
	Type             string `json:"type,omitempty"`
	Number           string `json:"number,omitempty"`
	Version          int    `json:"version,omitempty"`
	TNV              string
	Level            int    `json:"level,omitempty"`
	Credits          int    `json:"credits,omitempty"`
	Weighting        any    `json:"weighting,omitempty"`
	Points           any    `json:"points,omitempty"`
	Title            string `json:"title,omitempty"`
	Description      any    `json:"description,omitempty"`
	Purpose          any    `json:"purpose,omitempty"`
	Subfield         string `json:"subfield,omitempty"`
	Internalexternal string `json:"internalexternal,omitempty"`
}

type AssessmentModel struct {
	DB *sql.DB
}

func (a *Assessment) CreateTNV() {
	tnv := strings.Join([]string{a.Type, a.Number, strconv.Itoa(a.Version)}, "_")
	a.TNV = tnv
}

func (m AssessmentModel) InsertManyAssessments(assessments []Assessment) error {
	// Start a transaction (tx)
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback transaction if there's an error

	stmt, err := tx.Prepare(`
	INSERT INTO assessments (credits, description, internalexternal, level, number, points, purpose, subfield, title, tnv, type, version, weighting) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Insert entries in batches - extrapolate into own function
	batchSize := 100 // adjust as needed
	for i := 0; i < len(assessments); i += batchSize {
		// Create a slice of 100 results
		batch := assessments[i:min(i+batchSize, len(assessments))]

		// Insert each entry
		for _, assessment := range batch {
			assessment.CreateTNV()
			_, err := stmt.Exec(assessment.Credits, assessment.Description, assessment.Internalexternal, assessment.Level, assessment.Number, assessment.Points, assessment.Purpose, assessment.Subfield, assessment.Title, assessment.TNV, assessment.Type, assessment.Version, assessment.Weighting)
			if err != nil {
				return err
			}
		}
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return err
	}

	// Database insert succeeded
	return nil
}
