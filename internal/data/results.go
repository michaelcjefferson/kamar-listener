package data

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"
)

// Use json.RawMessage to allow the result and resultdata fields, which are arrays with varying values and datatypes within, to still be written to SQLite. JSON can be written to SQLite as TEXT
type Result struct {
	Code              *any            `json:"code,omitempty"`
	Comment           *string         `json:"comment,omitempty"`
	Course            *string         `json:"course,omitempty"`
	CurriculumLevel   *any            `json:"curriculumlevel,omitempty"`
	Date              *string         `json:"date,omitempty"`
	Enrolled          *bool           `json:"enrolled,omitempty"`
	ID                *int            `json:"id,omitempty"`
	NSN               *string         `json:"nsn,omitempty"`
	Number            *string         `json:"number,omitempty"`
	Published         *string         `json:"published,omitempty"`
	Result            *string         `json:"result,omitempty"`
	ResultData        json.RawMessage `json:"resultData,omitempty"`
	Results           json.RawMessage `json:"results,omitempty"`
	Subject           *string         `json:"subject,omitempty"`
	TNV               *string
	Type              *string `json:"type,omitempty"`
	Version           *int    `json:"version,omitempty"`
	Year              *int    `json:"year,omitempty"`
	YearLevel         *int    `json:"yearlevel,omitempty"`
	ListenerUpdatedAt string
}

type ResultModel struct {
	DB *sql.DB
}

func (r *Result) CreateTNV() {
	tnv := strings.Join([]string{*r.Type, *r.Number, strconv.Itoa(*r.Version)}, "_")
	r.TNV = &tnv
}

func (m *ResultModel) InsertManyResults(results []Result) error {
	// Start a transaction (tx)
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback transaction if there's an error

	stmt, err := tx.Prepare(`
	INSERT INTO results (code, comment, course, curriculumlevel, date, enrolled, id, nsn, number, published, result, resultData, results, subject, tnv, type, version, year, yearlevel)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
	ON CONFLICT(id, tnv, subject) DO UPDATE SET
		code = excluded.code,
		comment = excluded.comment,
		course = excluded.course,
		curriculumlevel = excluded.curriculumlevel,
		date = excluded.date,
		enrolled = excluded.enrolled,
		id = excluded.id,
		nsn = excluded.nsn,
		published = excluded.published,
		result = excluded.result,
		resultData = excluded.resultData,
		results = excluded.results,
		year = excluded.year,
		yearlevel = excluded.yearlevel,
		listener_updated_at = (datetime('now'))
	;`)
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
			_, err := stmt.Exec(result.Code, result.Comment, result.Course, result.CurriculumLevel, result.Date, result.Enrolled, result.ID, result.NSN, result.Number, result.Published, result.Result, result.ResultData, result.Results, result.Subject, result.TNV, result.Type, result.Version, result.Year, result.YearLevel)
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

func (m *ResultModel) GetResultsCount() (int, int, error) {
	today, total, err := QueryForRecordCounts("results", m.DB)

	return today, total, err
}
