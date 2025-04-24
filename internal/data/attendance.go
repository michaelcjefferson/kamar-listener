package data

import "database/sql"

type Attendance struct {
	ID     int    `json:"id,omitempty"`
	Nsn    string `json:"nsn,omitempty"`
	Values []struct {
		Date  string `json:"date,omitempty"`
		Codes string `json:"codes,omitempty"`
		Alt   string `json:"alt,omitempty"`
		Hdu   int    `json:"hdu,omitempty"`
		Hdj   int    `json:"hdj,omitempty"`
		Hdp   int    `json:"hdp,omitempty"`
	} `json:"values,omitempty"`
}

type AttendanceModel struct {
	DB *sql.DB
}

func (m *AttendanceModel) InsertManyAttendance(attendance []Attendance) error {
	// Start a transaction (tx)
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback transaction if there's an error

	// INSERT OR IGNORE prevents conflict errors from being thrown if the provided student_id already exists in the database
	attStmt, err := tx.Prepare(`
	INSERT OR IGNORE INTO attendance (student_id, nsn) VALUES ($1, $2);`)
	if err != nil {
		return err
	}
	defer attStmt.Close()

	attValStmt, err := tx.Prepare(`
	INSERT INTO attendance_values (student_id, date, codes, alt, hdu, hdj, hdp)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT(student_id, date) DO UPDATE SET
		codes = excluded.codes,
		alt = excluded.alt,
		hdu = excluded.hdu,
		hdj = excluded.hdj,
		hdp = excluded.hdp
	;`)
	if err != nil {
		return err
	}
	defer attValStmt.Close()

	// Insert entries in batches - extrapolate into own function
	batchSize := 100 // adjust as needed
	for i := 0; i < len(attendance); i += batchSize {
		// Create a slice of 100 results
		batch := attendance[i:min(i+batchSize, len(attendance))]

		// Insert each entry
		for _, att := range batch {
			_, err := attStmt.Exec(att.ID, att.Nsn)
			if err != nil {
				return err
			}

			for _, val := range att.Values {
				_, err = attValStmt.Exec(att.ID, val.Date, val.Codes, val.Alt, val.Hdu, val.Hdj, val.Hdp)
				if err != nil {
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

func (m *AttendanceModel) GetAttendanceCount() (int, int, error) {
	today, total := 0, 0

	tod, tot, err := QueryForRecordCounts("attendance", m.DB)
	if err != nil {
		return 0, 0, err
	}
	today += tod
	total += tot

	tod, tot, err = QueryForRecordCounts("attendance_values", m.DB)
	if err != nil {
		return 0, 0, err
	}
	today += tod
	total += tot

	return today, total, err
}
