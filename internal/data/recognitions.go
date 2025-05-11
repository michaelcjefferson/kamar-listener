package data

import (
	"database/sql"
	"strconv"
	"strings"
)

type Recognition struct {
	Count             *int    `json:"count"`
	ID                *int    `json:"id"`
	NSN               *string `json:"nsn,omitempty"`
	UUID              *string `json:"uuid,omitempty"`
	Date              *string `json:"date"`
	Slot              *int    `json:"slot"`
	Term              *int    `json:"term,omitempty"`
	Week              *int    `json:"week,omitempty"`
	Subject           *string `json:"subject,omitempty"`
	User              *string `json:"user,omitempty"`
	Points            *int    `json:"points,omitempty"`
	Comment           *string `json:"comment,omitempty"`
	Values            []int   `json:"values,omitempty"`
	ListenerUpdatedAt string
}

type RecognitionsModel struct {
	DB *sql.DB
}

func (m *RecognitionsModel) InsertManyRecognitions(recognitions []Recognition) error {
	// Start a transaction (tx)
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback transaction if there's an error

	stmt, err := tx.Prepare(`
	INSERT INTO recognitions (count, id, nsn, uuid, date, slot, term, week, subject, user, points, comment, values)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	ON CONFLICT(id, date, slot) DO UPDATE SET
		count = excluded.count,
		nsn = excluded.nsn,
		uuid = excluded.uuid,
		term = excluded.term,
		week = excluded.week,
		subject = excluded.subject,
		user = excluded.user,
		points = excluded.points,
		comment = excluded.comment,
		values = excluded.values,
		listener_updated_at = (datetime('now'))
	;`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Insert entries in batches - extrapolate into own function
	batchSize := 100 // adjust as needed
	for i := 0; i < len(recognitions); i += batchSize {
		// Create a slice of 100 results
		batch := recognitions[i:min(i+batchSize, len(recognitions))]

		// Insert each entry
		for _, recognition := range batch {
			// TODO: TEST AND REFINE
			vals := []string{}
			for _, v := range recognition.Values {
				vals = append(vals, strconv.Itoa(v))
			}

			_, err := stmt.Exec(recognition.Count, recognition.ID, recognition.NSN, recognition.UUID, recognition.Date, recognition.Slot, recognition.Term, recognition.Week, recognition.Subject, recognition.User, recognition.Points, recognition.Comment, strings.Join(vals, ","))
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

func (m *RecognitionsModel) GetRecognitionsCount() (int, int, error) {
	today, total, err := QueryForRecordCounts("recognitions", m.DB)

	return today, total, err
}
