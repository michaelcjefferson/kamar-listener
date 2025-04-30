package data

import "database/sql"

type Subject struct {
	ID                string `json:"id,omitempty"`
	Created           int64  `json:"created,omitempty"`
	Name              string `json:"name,omitempty"`
	Department        string `json:"department,omitempty"`
	Subdepartment     any    `json:"subdepartment,omitempty"`
	Qualification     string `json:"qualification,omitempty"`
	Level             int    `json:"level,omitempty"`
	ListenerUpdatedAt string
}

type SubjectModel struct {
	DB *sql.DB
}

// TODO: Possibly check if ID already exists, and update instead of insert in those cases - change query to ON CONFLICT? It shouldn't happen often
func (m *SubjectModel) InsertManySubjects(subjects []Subject) error {
	// Start a transaction (tx)
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback transaction if there's an error

	stmt, err := tx.Prepare(`
	INSERT INTO subjects (id, created, name, department, subdepartment, qualification, level)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT(id) DO UPDATE SET
		created = excluded.created,
		name = excluded.name,
		department = excluded.department,
		subdepartment = excluded.subdepartment,
		qualification = excluded.qualification,
		level = excluded.level,
		listener_updated_at = (datetime('now'))
	;`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Insert entries in batches - extrapolate into own function
	batchSize := 100 // adjust as needed
	for i := 0; i < len(subjects); i += batchSize {
		// Create a slice of 100 results
		batch := subjects[i:min(i+batchSize, len(subjects))]

		// Insert each entry
		for _, s := range batch {
			_, err := stmt.Exec(s.ID, s.Created, s.Name, s.Department, s.Subdepartment, s.Qualification, s.Level)
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

func (m *SubjectModel) GetSubjectsCount() (int, int, error) {
	today, total, err := QueryForRecordCounts("subjects", m.DB)

	return today, total, err
}
