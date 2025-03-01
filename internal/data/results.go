package data

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"
)

type Assessment struct {
	Count            any    `json:"count",omitempty`
	Credits          any    `json:"credits",omitempty`
	Description      string `json:"description",omitempty`
	ID               int    `json:"id",omitempty`
	InternalExternal string `json:"internalexternal",omitempty`
	Level            int    `json:"level",omitempty`
	Number           string `json:"number",omitempty`
	Points           any    `json:"points",omitempty`
	Purpose          any    `json:"purpose",omitempty`
	Subfield         any    `json:"subfield",omitempty`
	TNV              string
	Type             string `json:"type",omitempty`
	Version          int    `json:"version",omitempty`
	Weighting        any    `json:"weighting",omitempty`
}

// Use json.RawMessage to allow the result and resultdata fields, which are arrays with varying values and datatypes within, to still be written to SQLite. JSON can be written to SQLite as TEXT
type Result struct {
	Code            any             `json:"code,omitempty"`
	Comment         string          `json:"comment,omitempty"`
	Course          string          `json:"course,omitempty"`
	CurriculumLevel any             `json:"curriculumlevel,omitempty"`
	Date            string          `json:"date,omitempty"`
	Enrolled        bool            `json:"enrolled,omitempty"`
	ID              int             `json:"id,omitempty"`
	NSN             string          `json:"nsn,omitempty"`
	Number          string          `json:"number,omitempty"`
	Published       bool            `json:"published,omitempty"`
	Result          string          `json:"result,omitempty"`
	ResultData      json.RawMessage `json:"resultData,omitempty"`
	Results         json.RawMessage `json:"results,omitempty"`
	Subject         string          `json:"subject,omitempty"`
	TNV             string
	Type            string `json:"type,omitempty"`
	Version         int    `json:"version,omitempty"`
	Year            int    `json:"year,omitempty"`
	YearLevel       int    `json:"yearlevel,omitempty"`
}

type ResultModel struct {
	DB *sql.DB
}

func (a *Assessment) CreateTNV() {
	tnv := strings.Join([]string{a.Type, a.Number, strconv.Itoa(a.Version)}, "_")
	a.TNV = tnv
}

func (r *Result) CreateTNV() {
	tnv := strings.Join([]string{r.Type, r.Number, strconv.Itoa(r.Version)}, "_")
	r.TNV = tnv
}

func (m ResultModel) InsertManyAssessments(assessments []Assessment) error {
	// Start a transaction (tx)
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback transaction if there's an error

	stmt, err := tx.Prepare(`
	INSERT INTO assessments (count, credits, description, id, internalexternal, level, number, points, purpose, subfield, tnv, type, version, weighting) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`)
	// INSERT INTO results (code, comment, course, curriculumlevel, date, enrolled, id, nsn, number, published, result, subject, tnv, type, version, year, yearlevel) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`)
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
			_, err := stmt.Exec(assessment.Count, assessment.Credits, assessment.Description, assessment.ID, assessment.InternalExternal, assessment.Level, assessment.Number, assessment.Points, assessment.Purpose, assessment.Subfield, assessment.TNV, assessment.Type, assessment.Version, assessment.Weighting)
			// _, err := stmt.Exec(result.Code, result.Comment, result.Course, result.CurriculumLevel, result.Date, result.Enrolled, result.ID, result.NSN, result.Number, result.Published, result.Result, result.Subject, result.TNV, result.Type, result.Version, result.Year, result.YearLevel)
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

func (m ResultModel) InsertManyResults(results []Result) error {
	// Start a transaction (tx)
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback transaction if there's an error

	stmt, err := tx.Prepare(`
	INSERT INTO results (code, comment, course, curriculumlevel, date, enrolled, id, nsn, number, published, result, resultData, results, subject, type, version, year, yearlevel) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`)
	// INSERT INTO results (code, comment, course, curriculumlevel, date, enrolled, id, nsn, number, published, result, subject, tnv, type, version, year, yearlevel) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Insert entries in batches - extrapolate into own function
	batchSize := 100 // adjust as needed
	for i := 0; i < len(results); i += batchSize {
		// Create a slice of 100 results
		batch := results[i:min(i+batchSize, len(results))]

		// Insert each entry
		for _, result := range batch {
			result.CreateTNV()
			_, err := stmt.Exec(result.Code, result.Comment, result.Course, result.CurriculumLevel, result.Date, result.Enrolled, result.ID, result.NSN, result.Number, result.Published, result.Result, result.ResultData, result.Results, result.Subject, result.Type, result.Version, result.Year, result.YearLevel)
			// _, err := stmt.Exec(result.Code, result.Comment, result.Course, result.CurriculumLevel, result.Date, result.Enrolled, result.ID, result.NSN, result.Number, result.Published, result.Result, result.Subject, result.TNV, result.Type, result.Version, result.Year, result.YearLevel)
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

// Utility function for the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
