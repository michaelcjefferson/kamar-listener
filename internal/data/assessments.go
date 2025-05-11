package data

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"
)

type Assessment struct {
	Type              *string `json:"type,omitempty"`
	Number            *string `json:"number,omitempty"`
	Version           *int    `json:"version,omitempty"`
	TNV               *string
	Level             *int    `json:"level,omitempty"`
	Credits           *int    `json:"credits,omitempty"`
	Weighting         *any    `json:"weighting,omitempty"`
	Points            *any    `json:"points,omitempty"`
	Title             *string `json:"title,omitempty"`
	Description       *any    `json:"description,omitempty"`
	Purpose           *any    `json:"purpose,omitempty"`
	SchoolRef         *any    `json:"schoolref,omitempty"`
	Subfield          *string `json:"subfield,omitempty"`
	Internalexternal  *string `json:"internalexternal,omitempty"`
	ListenerUpdatedAt string
}

type AssessmentModel struct {
	DB *sql.DB
}

func (a *Assessment) CreateTNV() {
	tnv := strings.Join([]string{*a.Type, *a.Number, strconv.Itoa(*a.Version)}, "_")
	a.TNV = &tnv
}

func (m *AssessmentModel) InsertManyAssessments(assessments []Assessment) error {
	// Start a transaction (tx)
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback transaction if there's an error

	stmt, err := tx.Prepare(`
	INSERT INTO assessments (credits, description, internalexternal, level, number, points, purpose, schoolref, subfield, title, tnv, type, version, weighting)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	ON CONFLICT(tnv) DO UPDATE SET
		credits = excluded.credits,
		description = excluded.description,
		internalexternal = excluded.internalexternal,
		level = excluded.level,
		points = excluded.points,
		purpose = excluded.purpose,
		schoolref = excluded.schoolref,
		subfield = excluded.subfield,
		title = excluded.title,
		weighting = excluded.weighting,
		listener_updated_at = (datetime('now'))
	;`)
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
			_, err := stmt.Exec(assessment.Credits, assessment.Description, assessment.Internalexternal, assessment.Level, assessment.Number, assessment.Points, assessment.Purpose, assessment.SchoolRef, assessment.Subfield, assessment.Title, assessment.TNV, assessment.Type, assessment.Version, assessment.Weighting)
			if err != nil {
				if err.Error() == `UNIQUE constraint failed: assessments.tnv` {
					log.Printf("assessment being written: %v\n\n", assessment)

					var ass Assessment

					q := `
						SELECT credits, description, internalexternal, level, number, points, purpose, schoolref, subfield, title, tnv, type, version, weighting
						FROM assessments
						WHERE tnv = ?
					`

					ct, ca := context.WithTimeout(context.Background(), 3*time.Second)
					defer ca()

					err := m.DB.QueryRowContext(ct, q, assessment.TNV).Scan(
						&ass.Credits,
						&ass.Description,
						&ass.Internalexternal,
						&ass.Level,
						&ass.Number,
						&ass.Points,
						&ass.Purpose,
						&ass.SchoolRef,
						&ass.Subfield,
						&ass.Title,
						&ass.TNV,
						&ass.Type,
						&ass.Version,
						&ass.Weighting,
					)
					log.Printf("assessment that already exists: %v\n\n", ass)

					return err
				} else {
					return err
				}
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

// Return one assessment that matches the provided assessment number
func (m *AssessmentModel) GetByAssessmentNumber(num string) (*Assessment, error) {
	var assessment Assessment

	if i, err := strconv.Atoi(num); i < 1 || err != nil {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT credits, description, internalexternal, level, number, points, purpose, schoolref, subfield, title, tnv, type, version, weighting
		FROM assessments
		WHERE number = ?
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, num).Scan(
		&assessment.Credits,
		&assessment.Description,
		&assessment.Internalexternal,
		&assessment.Level,
		&assessment.Number,
		&assessment.Points,
		&assessment.Purpose,
		&assessment.SchoolRef,
		&assessment.Subfield,
		&assessment.Title,
		&assessment.TNV,
		&assessment.Type,
		&assessment.Version,
		&assessment.Weighting,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &assessment, nil
}

func (m *AssessmentModel) GetAssessmentCount() (int, int, error) {
	today, total, err := QueryForRecordCounts("assessments", m.DB)

	return today, total, err
}
