package data

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

type Staff struct {
	ID                 *string      `json:"id,omitempty"`
	UUID               *string      `json:"uuid,omitempty"`
	Role               *string      `json:"role,omitempty"`
	Created            int64        `json:"created,omitempty"`
	Uniqueid           *int         `json:"uniqueid,omitempty"`
	Username           *string      `json:"username,omitempty"`
	Firstname          *string      `json:"firstname,omitempty"`
	Lastname           *string      `json:"lastname,omitempty"`
	Gender             *string      `json:"gender,omitempty"`
	Groups             []Group      `json:"groups,omitempty"`
	SchoolsIndex       []int        `json:"schoolindex,omitempty"`
	Title              *string      `json:"title,omitempty"`
	Email              *string      `json:"email,omitempty"`
	Mobile             *string      `json:"mobile,omitempty"`
	Extension          *int         `json:"extension,omitempty"`
	Classification     *string      `json:"classification,omitempty"`
	Position           *string      `json:"position,omitempty"`
	House              *string      `json:"house,omitempty"`
	Tutor              *string      `json:"tutor,omitempty"`
	DateBirth          *int         `json:"datebirth,omitempty"`
	LeavingDate        *any         `json:"leavingdate,omitempty"`
	StartingDate       *any         `json:"startingdate,omitempty"`
	ESLGUID            *string      `json:"eslguid,omitempty"`
	MOENumber          *any         `json:"moenumber,omitempty"`
	PhotocopierID      *any         `json:"photocopierid,omitempty"`
	RegistrationNumber *any         `json:"registrationnumber,omitempty"`
	Custom             *CustomField `json:"custom,omitempty"`
	ListenerUpdatedAt  string
}

type SchoolIndex int

type StaffModel struct {
	DB *sql.DB
}

func (m *StaffModel) InsertManyStaff(staff []Staff) error {
	// Start a transaction (tx)
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback transaction if there's an error

	staffStmt, err := tx.Prepare(`
	INSERT INTO staff (id, uuid, role, created, uniqueid, username, firstname, lastname, gender, schoolindex, title, email, mobile, extension, classification, position, house, tutor, datebirth, leavingdate, startingdate, eslguid, moenumber, photocopierid, registrationnumber, custom)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26)
	ON CONFLICT(uuid) DO UPDATE SET
		id = excluded.id,
		role = excluded.role,
		created = excluded.created,
		uniqueid = excluded.uniqueid,
		username = excluded.username,
		firstname = excluded.firstname,
		lastname = excluded.lastname,
		gender = excluded.gender,
		schoolindex = excluded.schoolindex,
		title = excluded.title,
		email = excluded.email,
		mobile = excluded.mobile,
		extension = excluded.extension,
		classification = excluded.classification,
		position = excluded.position,
		house = excluded.house,
		tutor = excluded.tutor,
		datebirth = excluded.datebirth,
		leavingdate = excluded.leavingdate,
		startingdate = excluded.startingdate,
		eslguid = excluded.eslguid,
		moenumber = excluded.moenumber,
		photocopierid = excluded.photocopierid,
		registrationnumber = excluded.registrationnumber,
		custom = excluded.custom,
		listener_updated_at = (datetime('now'))
	;`)
	if err != nil {
		return err
	}
	defer staffStmt.Close()

	staffClassGrpUpdateStmt, err := tx.Prepare(`
	UPDATE staff_groups
	SET
		staff_id = $1,
		type = $2,
		subject = $3,
		year = $4,
		name = $5,
		description = $6,
		teacher = $7,
		showreport = $8,
		listener_updated_at = (datetime('now'))
	WHERE staff_uuid = $9 AND coreoption = $10
	;`)
	if err != nil {
		return err
	}
	defer staffClassGrpUpdateStmt.Close()

	staffOtherGrpUpdateStmt, err := tx.Prepare(`
	UPDATE staff_groups
	SET
		staff_id = $1,
		type = $2,
		subject = $3,
		year = $4,
		name = $5,
		description = $6,
		teacher = $7,
		showreport = $8,
		listener_updated_at = (datetime('now'))
	WHERE staff_uuid = $9 AND ref = $10
	;`)
	if err != nil {
		return err
	}
	defer staffOtherGrpUpdateStmt.Close()

	staffGrpStmt, err := tx.Prepare(`
	INSERT INTO staff_groups (staff_uuid, staff_id, type, subject, coreoption, ref, year, name, description, teacher, showreport)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);`)
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
			var schoolIndexJSON, customJSON []byte
			var err error

			if len(s.SchoolsIndex) > 0 {
				schoolIndexJSON, err = json.Marshal(s.SchoolsIndex)
				if err != nil {
					return fmt.Errorf("marshal iwi: %w", err)
				}
			} else {
				schoolIndexJSON = nil
			}

			if s.Custom != nil {
				customJSON, err = json.Marshal(s.Custom)
				if err != nil {
					return fmt.Errorf("marshal custom: %w", err)
				}
			} else {
				customJSON = nil
			}

			_, err = staffStmt.Exec(s.ID, s.UUID, s.Role, s.Created, s.Uniqueid, s.Username, s.Firstname, s.Lastname, s.Gender, schoolIndexJSON, s.Title, s.Email, s.Mobile, s.Extension, s.Classification, s.Position, s.House, s.Tutor, s.DateBirth, s.LeavingDate, s.StartingDate, s.ESLGUID, s.MOENumber, s.PhotocopierID, s.RegistrationNumber, customJSON)
			if err != nil {
				return err
			}

			// Check for group type (class or group), and check for pre-existing rows with corresponding unique identifiers. If a match is found, run an update - if not, insert a new row
			for _, g := range s.Groups {
				switch *g.Type {
				case "class":
					var exists bool

					err := tx.QueryRow(`SELECT 1 FROM staff_groups WHERE staff_uuid = ? AND coreoption = ? LIMIT 1;`, s.UUID, g.Coreoption).Scan(&exists)
					if err != nil && err != sql.ErrNoRows {
						return err
					}
					if exists {
						_, err := staffClassGrpUpdateStmt.Exec(s.ID, g.Type, g.Subject, g.Year, g.Name, g.Description, g.Teacher, g.ShowReport, s.UUID, g.Coreoption)
						if err != nil {
							return err
						}
						continue
					}
				case "group":
					var exists bool

					err := tx.QueryRow(`SELECT 1 FROM staff_groups WHERE staff_uuid = ? AND ref = ? LIMIT 1;`, s.UUID, g.Ref).Scan(&exists)
					if err != nil && err != sql.ErrNoRows {
						return err
					}
					if exists {
						_, err := staffOtherGrpUpdateStmt.Exec(s.ID, g.Type, g.Subject, g.Year, g.Name, g.Description, g.Teacher, g.ShowReport, s.UUID, g.Ref)
						if err != nil {
							return err
						}
						continue
					}
				default:
					// TODO: Handle gracefully, rather than preventing writes?
					return ErrUnfoundGroupType
				}

				// Insert new row if a matching previous entry wasn't found
				_, err = staffGrpStmt.Exec(s.UUID, s.ID, g.Type, g.Subject, g.Coreoption, g.Ref, g.Year, g.Name, g.Description, g.Teacher, g.ShowReport)
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

func (m *StaffModel) GetStaffCount() (int, int, error) {
	today, total := 0, 0

	tod, tot, err := QueryForRecordCounts("staff", m.DB)
	if err != nil {
		return 0, 0, err
	}
	today += tod
	total += tot

	tod, tot, err = QueryForRecordCounts("staff_groups", m.DB)
	if err != nil {
		return 0, 0, err
	}
	today += tod
	total += tot

	return today, total, err
}
