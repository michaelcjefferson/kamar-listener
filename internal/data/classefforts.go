package data

import (
	"database/sql"
	"strconv"
	"strings"
)

type ClassEffort struct {
	Count             *int    `json:"count"`
	ID                *int    `json:"id"`
	NSN               *string `json:"nsn,omitempty"`
	Date              *string `json:"date"`
	Slot              *int    `json:"slot"`
	Term              *int    `json:"term,omitempty"`
	Week              *int    `json:"week,omitempty"`
	Subject           *string `json:"subject,omitempty"`
	User              *string `json:"user,omitempty"`
	Efforts           []int   `json:"efforts,omitempty"`
	ListenerUpdatedAt string
}

type ClassEffortsModel struct {
	DB *sql.DB
}

func (m *ClassEffortsModel) InsertManyClassEfforts(efforts []ClassEffort) error {
	// Start a transaction (tx)
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback transaction if there's an error

	stmt, err := tx.Prepare(`
	INSERT INTO class_efforts (count, id, nsn, date, slot, term, week, subject, user, efforts)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	ON CONFLICT(id, date, slot) DO UPDATE SET
		count = excluded.count,
		nsn = excluded.nsn,
		term = excluded.term,
		week = excluded.week,
		subject = excluded.subject,
		user = excluded.user,
		efforts = excluded.efforts,
		listener_updated_at = (datetime('now'))
	;`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Insert entries in batches - extrapolate into own function
	batchSize := 100 // adjust as needed
	for i := 0; i < len(efforts); i += batchSize {
		// Create a slice of 100 results
		batch := efforts[i:min(i+batchSize, len(efforts))]

		// Insert each entry
		for _, effort := range batch {
			// TODO: TEST AND REFINE
			effs := []string{}
			for _, e := range effort.Efforts {
				effs = append(effs, strconv.Itoa(e))
			}

			_, err := stmt.Exec(effort.Count, effort.ID, effort.NSN, effort.Date, effort.Slot, effort.Term, effort.Week, effort.Subject, effort.User, strings.Join(effs, ","))
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

func (m *ClassEffortsModel) GetClassEffortsCount() (int, int, error) {
	today, total, err := QueryForRecordCounts("class_efforts", m.DB)

	return today, total, err
}
