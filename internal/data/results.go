package data

import (
	"database/sql"
)

type Result struct {
	Code            any      `json:"code,omitempty"`
	Comment         string   `json:"comment,omitempty"`
	Course          string   `json:"course,omitempty"`
	CurriculumLevel any      `json:"curriculumlevel,omitempty"`
	Date            string   `json:"date,omitempty"`
	Enrolled        bool     `json:"enrolled,omitempty"`
	ID              int      `json:"id,omitempty"`
	NSN             string   `json:"nsn,omitempty"`
	Number          string   `json:"number,omitempty"`
	Published       bool     `json:"published,omitempty"`
	Result          string   `json:"result,omitempty"`
	ResultData      []any    `json:"resultData,omitempty"`
	Results         []string `json:"results,omitempty"`
	Subject         string   `json:"subject,omitempty"`
	Type            string   `json:"type,omitempty"`
	Version         int      `json:"version,omitempty"`
	Year            int      `json:"year,omitempty"`
	YearLevel       int      `json:"yearlevel,omitempty"`
}

type ResultData struct {
}

type ResultModel struct {
	DB *sql.DB
}

func (r ResultModel) InsertMany(results []Result) error {
	// Start a transaction (tx)
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback transaction if there's an error

	stmt, err := tx.Prepare(`
	INSERT INTO results (code, comment, course, curriculumlevel, date, enrolled, id, nsn, number, published, result, subject, type, version, year, yearlevel) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`)
	// INSERT INTO results (code, comment, course, curriculumlevel, date, enrolled, id, nsn, number, published, result, resultData, results, subject, type, version, year, yearlevel) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Insert entries in batches
	batchSize := 100 // adjust as needed
	for i := 0; i < len(results); i += batchSize {
		// Create a slice of 100 results
		batch := results[i:min(i+batchSize, len(results))]

		// Insert each entry
		for _, result := range batch {
			_, err := stmt.Exec(result.Code, result.Comment, result.Course, result.CurriculumLevel, result.Date, result.Enrolled, result.ID, result.NSN, result.Number, result.Published, result.Result, result.Subject, result.Type, result.Version, result.Year, result.YearLevel)
			// _, err := stmt.Exec(result.Code, result.Comment, result.Course, result.CurriculumLevel, result.Date, result.Enrolled, result.ID, result.NSN, result.Number, result.Published, result.Result, result.ResultData, result.Results, result.Subject, result.Type, result.Version, result.Year, result.YearLevel)
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
