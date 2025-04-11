package data

import (
	"database/sql"
	"encoding/json"
)

type Staff struct {
	ID                 string          `json:"id,omitempty"`
	UUID               string          `json:"uuid,omitempty"`
	Role               string          `json:"role,omitempty"`
	Created            int64           `json:"created,omitempty"`
	Uniqueid           int             `json:"uniqueid,omitempty"`
	Username           string          `json:"username,omitempty"`
	Firstname          string          `json:"firstname,omitempty"`
	Lastname           string          `json:"lastname,omitempty"`
	Gender             string          `json:"gender,omitempty"`
	Groups             []Group         `json:"groups,omitempty"`
	SchoolIndex        int             `json:"schoolindex,omitempty"`
	Title              string          `json:"title,omitempty"`
	Email              string          `json:"email,omitempty"`
	Mobile             string          `json:"mobile,omitempty"`
	Extension          string          `json:"extension,omitempty"`
	Classification     string          `json:"classification,omitempty"`
	Position           string          `json:"position,omitempty"`
	House              string          `json:"house,omitempty"`
	Tutor              string          `json:"tutor,omitempty"`
	DateBirth          any             `json:"datebirth,omitempty"`
	LeavingDate        any             `json:"leavingdate,omitempty"`
	StartingDate       any             `json:"startingdate,omitempty"`
	ESLGUID            any             `json:"eslguid,omitempty"`
	MOENumber          any             `json:"moenumber,omitempty"`
	PhotocopierID      any             `json:"photocopierid,omitempty"`
	RegistrationNumber any             `json:"registrationnumber,omitempty"`
	Custom             json.RawMessage `json:"custom,omitempty"`
}

type StaffModel struct {
	DB *sql.DB
}

// TODO: Check if ID already exists, and update instead of insert in those cases - change query to ON CONFLICT? It shouldn't happen often
func (m *StaffModel) InsertManyStaff(staff []Staff) error {
	// Start a transaction (tx)
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback transaction if there's an error

	staffStmt, err := tx.Prepare(`
	INSERT INTO staff (id, uuid, role, created, uniqueid, username, firstname, lastname, gender, schoolindex, title, email, mobile, extension, classification, position, house, tutor, datebirth, leavingdate, startingdate, eslguid, moenumber, photocopierid, registrationnumber, custom) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26)`)
	if err != nil {
		return err
	}
	defer staffStmt.Close()

	staffGrpStmt, err := tx.Prepare(`
	INSERT INTO staff_groups (staff_uniqueid, type, subject, coreoption) VALUES ($1, $2, $3, $4)
	`)
	if err != nil {
		return err
	}
	defer staffGrpStmt.Close()

	// Insert entries in batches - extrapolate into own function
	batchSize := 100 // adjust as needed
	for i := 0; i < len(staff); i += batchSize {
		// Create a slice of 100 results
		batch := staff[i:min(i+batchSize, len(staff))]

		// Insert each entry
		for _, s := range batch {
			_, err := staffStmt.Exec(s.ID, s.UUID, s.Role, s.Created, s.Uniqueid, s.Username, s.Firstname, s.Lastname, s.Gender, s.SchoolIndex, s.Title, s.Email, s.Mobile, s.Extension, s.Classification, s.Position, s.House, s.Tutor, s.DateBirth, s.LeavingDate, s.StartingDate, s.ESLGUID, s.MOENumber, s.PhotocopierID, s.RegistrationNumber, s.Custom)
			if err != nil {
				return err
			}

			for _, g := range s.Groups {
				_, err = staffGrpStmt.Exec(s.Uniqueid, g.Type, g.Subject, g.Coreoption)
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
