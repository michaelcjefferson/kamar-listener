package data

import "database/sql"

// TODO: Find out whether or not teacher timetables might be sent with a similar format - if so, define StudentTimetable and TeacherTimetable structs
type Timetable struct {
	Student   int    `json:"student,omitempty"`
	UUID      string `json:"uuid,omitempty"`
	Grid      string `json:"grid,omitempty"`
	Timetable string `json:"timetable,omitempty"`
}

type TimetableModel struct {
	DB *sql.DB
}

func (m TimetableModel) InsertManyTimetables(timetables []Timetable) error {
	// Start a transaction (tx)
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback transaction if there's an error

	stmt, err := tx.Prepare(`
	INSERT INTO timetables (student, uuid, grid, timetable) VALUES ($1, $2, $3, $4)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Insert entries in batches - extrapolate into own function
	batchSize := 100 // adjust as needed
	for i := 0; i < len(timetables); i += batchSize {
		// Create a slice of 100 results
		batch := timetables[i:min(i+batchSize, len(timetables))]

		// Insert each entry
		for _, t := range batch {
			_, err := stmt.Exec(t.Student, t.UUID, t.Grid, t.Timetable)
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
