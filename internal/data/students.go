package data

import (
	"database/sql"
	"encoding/json"
)

// Without clear examples or documentation from KAMAR, a number of fields' data types are unclear. These data types are Unmarshalled as "any" type to the struct, and then written as TEXT values to the database, to ensure they still come through
type Student struct {
	ID                int             `json:"id,omitempty"`
	UUID              string          `json:"uuid,omitempty"`
	Role              string          `json:"role,omitempty"`
	Created           int64           `json:"created,omitempty"`
	Uniqueid          int             `json:"uniqueid,omitempty"`
	Nsn               string          `json:"nsn,omitempty"`
	Username          string          `json:"username,omitempty"`
	Firstname         string          `json:"firstname,omitempty"`
	FirstnameLegal    string          `json:"firstnamelegal,omitempty"`
	Lastname          string          `json:"lastname,omitempty"`
	LastnameLegal     string          `json:"lastnamelegal,omitempty"`
	Forenames         string          `json:"forenames,omitempty"`
	ForenamesLegal    string          `json:"forenameslegal,omitempty"`
	Gender            any             `json:"gender,omitempty"`
	GenderPreferred   any             `json:"genderpreffered,omitempty"`
	Gendercode        int             `json:"gendercode,omitempty"`
	SchoolIndex       int             `json:"schoolindex,omitempty"`
	Email             string          `json:"email,omitempty"`
	Mobile            string          `json:"mobile,omitempty"`
	House             string          `json:"house,omitempty"`
	Whanau            string          `json:"whanau,omitempty"`
	Boarder           any             `json:"boarder,omitempty"`
	BYODInfo          any             `json:"byodinfo,omitempty"`
	ECE               any             `json:"ece,omitempty"`
	ESOL              any             `json:"esol,omitempty"`
	ORS               any             `json:"ors,omitempty"`
	LanguageSpoken    string          `json:"languagespoken,omitempty"`
	Datebirth         int             `json:"datebirth,omitempty"`
	Startingdate      int             `json:"startingdate,omitempty"`
	StartSchoolDate   any             `json:"startschooldate,omitempty"`
	Leavingdate       int             `json:"leavingdate,omitempty"`
	LeavingReason     string          `json:"leavingreason,omitempty"`
	LeavingSchool     string          `json:"leavingschool,omitempty"`
	LeavingActivity   any             `json:"leavingactivity,omitempty"`
	MOEType           string          `json:"moetype,omitempty"`
	EthnicityL1       any             `json:"ethnicityL1,omitempty"`
	EthnicityL2       any             `json:"ethnicityL2,omitempty"`
	Ethnicity         any             `json:"ethnicity,omitempty"`
	Iwi               string          `json:"iwi,omitempty"`
	YearLevel         any             `json:"yearlevel,omitempty"`
	FundingLevel      any             `json:"fundinglevel,omitempty"`
	Tutor             string          `json:"tutor,omitempty"`
	TimetableBottom1  any             `json:"timetablebottom1,omitempty"`
	TimetableBottom2  any             `json:"timetablebottom2,omitempty"`
	TimetableBottom3  any             `json:"timetablebottom3,omitempty"`
	TimetableBottom4  any             `json:"timetablebottom4,omitempty"`
	TimetableTop1     any             `json:"timetabletop1,omitempty"`
	TimetableTop2     any             `json:"timetabletop2,omitempty"`
	TimetableTop3     any             `json:"timetabletop3,omitempty"`
	TimetableTop4     any             `json:"timetabletop4,omitempty"`
	MaoriLevel        any             `json:"maorilevel,omitempty"`
	PacificLanguage   string          `json:"pacificlanguage,omitempty"`
	PacificLevel      any             `json:"pacificlevel,omitempty"`
	SiblingLink       any             `json:"siblinglink,omitempty"`
	PhotocopierID     any             `json:"photocopierid,omitempty"`
	SignedAgreement   any             `json:"signedagreement,omitempty"`
	AccountDisabled   any             `json:"accountdisabled,omitempty"`
	NetworkAccess     any             `json:"networkaccess,omitempty"`
	AltDescription    string          `json:"altdescription,omitempty"`
	AltHomeDrive      string          `json:"althomedrive,omitempty"`
	Flags             Flags           `json:"flags,omitempty"`
	Res               []Residence     `json:"res,omitempty"`
	Caregivers        []Caregiver     `json:"caregivers,omitempty"`
	Emergency         []Emergency     `json:"emergency,omitempty"`
	Groups            []Group         `json:"groups,omitempty"`
	Awards            []Award         `json:"awards,omitempty"`
	Datasharing       Datasharing     `json:"datasharing,omitempty"`
	Custom            json.RawMessage `json:"custom,omitempty"`
	ListenerUpdatedAt string
}

type Award struct {
	Type              string `json:"type,omitempty"`
	Name              string `json:"name,omitempty"`
	Year              int    `json:"year,omitempty"`
	Date              string `json:"date,omitempty"`
	ListenerUpdatedAt string
}

type Caregiver struct {
	Ref               int    `json:"ref,omitempty"`
	Role              string `json:"role,omitempty"`
	Name              string `json:"name,omitempty"`
	Email             string `json:"email,omitempty"`
	Mobile            string `json:"mobile,omitempty"`
	Relationship      string `json:"relationship,omitempty"`
	Status            any    `json:"status,omitempty"`
	ListenerUpdatedAt string
}

type Datasharing struct {
	Details           int `json:"details,omitempty"`
	Photo             int `json:"photo,omitempty"`
	Other             int `json:"other,omitempty"`
	ListenerUpdatedAt string
}

type Emergency struct {
	Name              string `json:"name,omitempty"`
	Relationship      string `json:"relationship,omitempty"`
	Mobile            string `json:"mobile,omitempty"`
	ListenerUpdatedAt string
}

type Flags struct {
	General string `json:"general,omitempty"`
	// TODO: Notes is a string, but its value (if it exists) is either 0 or 1 - better translated as a bool
	Notes             string `json:"notes,omitempty"`
	Alert             any    `json:"alert,omitempty"`
	Conditions        any    `json:"conditions,omitempty"`
	Dietary           any    `json:"dietary,omitempty"`
	Ibuprofen         any    `json:"ibuprofen,omitempty"`
	Medical           any    `json:"medical,omitempty"`
	Paracetamol       any    `json:"paracetamol,omitempty"`
	Pastoral          any    `json:"pastoral,omitempty"`
	Reactions         any    `json:"reactions,omitempty"`
	SpecialNeeds      any    `json:"specialneeds,omitempty"`
	Vaccinations      any    `json:"vaccinations,omitempty"`
	EOTCConsent       any    `json:"eotcconsent,omitempty"`
	EOTCForm          any    `json:"eotcform,omitempty"`
	ListenerUpdatedAt string
}

type Group struct {
	Type              string `json:"type,omitempty"`
	Subject           string `json:"subject,omitempty"`
	Coreoption        string `json:"coreoption,omitempty"`
	ListenerUpdatedAt string
}

type Residence struct {
	Title             string `json:"title,omitempty"`
	Salutation        string `json:"salutation,omitempty"`
	Email             string `json:"email,omitempty"`
	NumFlatUnit       any    `json:"numFlatUnit,omitempty"`
	NumStreet         any    `json:"numStreet,omitempty"`
	RuralDelivery     any    `json:"ruralDelivery,omitempty"`
	Suburb            string `json:"suburb,omitempty"`
	Town              string `json:"town,omitempty"`
	Postcode          any    `json:"postcode,omitempty"`
	ListenerUpdatedAt string
}

type StudentModel struct {
	DB *sql.DB
}

// TODO: Check if ID already exists, and update instead of insert in those cases - change query to ON CONFLICT? It shouldn't happen often
func (m *StudentModel) InsertManyStudents(students []Student) error {
	// Start a transaction (tx)
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback transaction if there's an error

	studentStmt, err := tx.Prepare(`
	INSERT INTO students (id, uuid, role, created, uniqueid, nsn, username, firstname, firstnamelegal, lastname, lastnamelegal, forenames, forenameslegal, gender, genderpreferred, gendercode, schoolindex, email, mobile, house, whanau, boarder, byodinfo, ece, esol, ors, languagespoken, datebirth, startingdate, startschooldate, leavingdate, leavingreason, leavingschool, leavingactivity, moetype, ethnicityL1, ethnicityL2, ethnicity, iwi, yearlevel, fundinglevel, tutor, timetablebottom1, timetablebottom2, timetablebottom3, timetablebottom4, timetabletop1, timetabletop2, timetabletop3, timetabletop4, maorilevel, pacificlanguage, pacificlevel, siblinglink, photocopierid, signedagreement, accountdisabled, networkaccess, altdescription, althomedrive, custom) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41, $42, $43, $44, $45, $46, $47, $48, $49, $50, $51, $52, $53, $54, $55, $56, $57, $58, $59, $60, $61)`)
	if err != nil {
		return err
	}
	defer studentStmt.Close()

	studentAwardStmt, err := tx.Prepare(`
	INSERT INTO student_awards (student_uuid, student_id, type, name, year, date) VALUES ($1, $2, $3, $4, $5, $6)
	`)
	if err != nil {
		return err
	}
	defer studentAwardStmt.Close()

	studentCareStmt, err := tx.Prepare(`
	INSERT INTO student_caregivers (student_uuid, student_id, ref, role, name, email, mobile, relationship, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`)
	if err != nil {
		return err
	}
	defer studentCareStmt.Close()

	studentDataStmt, err := tx.Prepare(`
	INSERT INTO student_datasharing (student_uuid, student_id, details, photo, other) VALUES ($1, $2, $3, $4, $5)
	`)
	if err != nil {
		return err
	}
	defer studentDataStmt.Close()

	studentEmgyStmt, err := tx.Prepare(`
	INSERT INTO student_emergency (student_uuid, student_id, name, relationship, mobile) VALUES ($1, $2, $3, $4, $5)
	`)
	if err != nil {
		return err
	}
	defer studentEmgyStmt.Close()

	studentFlagStmt, err := tx.Prepare(`
	INSERT INTO student_flags (student_uuid, student_id, general, notes, alert, conditions, dietary, ibuprofen, medical, paracetamol, pastoral, reactions, specialneeds, vaccinations, eotcconsent, eotcform) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`)
	if err != nil {
		return err
	}
	defer studentFlagStmt.Close()

	studentGrpStmt, err := tx.Prepare(`
	INSERT INTO student_groups (student_uuid, student_id, type, subject, coreoption) VALUES ($1, $2, $3, $4, $5)
	`)
	if err != nil {
		return err
	}
	defer studentGrpStmt.Close()

	studentResStmt, err := tx.Prepare(`
	INSERT INTO student_residences (student_uuid, student_id, title, salutation, email, numFlatUnit, numStreet, ruralDelivery, suburb, town, postcode) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`)
	if err != nil {
		return err
	}
	defer studentResStmt.Close()

	// Insert entries in batches - extrapolate into own function
	batchSize := 100 // adjust as needed
	for i := 0; i < len(students); i += batchSize {
		// Create a slice of 100 results
		batch := students[i:min(i+batchSize, len(students))]

		// Insert each entry
		for _, s := range batch {
			_, err := studentStmt.Exec(s.ID, s.UUID, s.Role, s.Created, s.Uniqueid, s.Nsn, s.Username, s.Firstname, s.FirstnameLegal, s.Lastname, s.LastnameLegal, s.Forenames, s.ForenamesLegal, s.Gender, s.GenderPreferred, s.Gendercode, s.SchoolIndex, s.Email, s.Mobile, s.House, s.Whanau, s.Boarder, s.BYODInfo, s.ECE, s.ESOL, s.ORS, s.LanguageSpoken, s.Datebirth, s.Startingdate, s.StartSchoolDate, s.Leavingdate, s.LeavingReason, s.LeavingSchool, s.LeavingActivity, s.MOEType, s.EthnicityL1, s.EthnicityL2, s.Ethnicity, s.Iwi, s.YearLevel, s.FundingLevel, s.Tutor, s.TimetableBottom1, s.TimetableBottom2, s.TimetableBottom3, s.TimetableBottom4, s.TimetableTop1, s.TimetableTop2, s.TimetableBottom3, s.TimetableBottom4, s.MaoriLevel, s.PacificLanguage, s.PacificLevel, s.SiblingLink, s.PhotocopierID, s.SignedAgreement, s.AccountDisabled, s.NetworkAccess, s.AltDescription, s.AltHomeDrive, s.Custom)
			if err != nil {
				return err
			}

			for _, a := range s.Awards {
				_, err = studentAwardStmt.Exec(s.UUID, s.ID, a.Type, a.Name, a.Year, a.Date)
				if err != nil {
					return err
				}
			}

			for _, c := range s.Caregivers {
				_, err = studentCareStmt.Exec(s.UUID, s.ID, c.Ref, c.Role, c.Name, c.Email, c.Mobile, c.Relationship, c.Status)
				if err != nil {
					return err
				}
			}

			_, err = studentDataStmt.Exec(s.UUID, s.ID, s.Datasharing.Details, s.Datasharing.Photo, s.Datasharing.Other)
			if err != nil {
				return err
			}

			for _, e := range s.Emergency {
				_, err = studentEmgyStmt.Exec(s.UUID, s.ID, e.Name, e.Relationship, e.Mobile)
				if err != nil {
					return err
				}
			}

			_, err = studentFlagStmt.Exec(s.UUID, s.ID, s.Flags.General, s.Flags.Notes, s.Flags.Alert, s.Flags.Conditions, s.Flags.Dietary, s.Flags.Ibuprofen, s.Flags.Medical, s.Flags.Paracetamol, s.Flags.Pastoral, s.Flags.Reactions, s.Flags.SpecialNeeds, s.Flags.Vaccinations, s.Flags.EOTCConsent, s.Flags.EOTCForm)
			if err != nil {
				return err
			}

			for _, g := range s.Groups {
				_, err = studentGrpStmt.Exec(s.UUID, s.ID, g.Type, g.Subject, g.Coreoption)
				if err != nil {
					return err
				}
			}

			for _, r := range s.Res {
				_, err = studentResStmt.Exec(s.UUID, s.ID, r.Title, r.Salutation, r.Email, r.NumFlatUnit, r.NumStreet, r.RuralDelivery, r.Suburb, r.Town, r.Postcode)
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

func (m *StudentModel) GetStudentsCount() (int, int, error) {
	today, total := 0, 0

	tod, tot, err := QueryForRecordCounts("students", m.DB)
	if err != nil {
		return 0, 0, err
	}
	today += tod
	total += tot

	tod, tot, err = QueryForRecordCounts("student_awards", m.DB)
	if err != nil {
		return 0, 0, err
	}
	today += tod
	total += tot

	tod, tot, err = QueryForRecordCounts("student_caregivers", m.DB)
	if err != nil {
		return 0, 0, err
	}
	today += tod
	total += tot

	tod, tot, err = QueryForRecordCounts("student_datasharing", m.DB)
	if err != nil {
		return 0, 0, err
	}
	today += tod
	total += tot

	tod, tot, err = QueryForRecordCounts("student_emergency", m.DB)
	if err != nil {
		return 0, 0, err
	}
	today += tod
	total += tot

	tod, tot, err = QueryForRecordCounts("student_flags", m.DB)
	if err != nil {
		return 0, 0, err
	}
	today += tod
	total += tot

	tod, tot, err = QueryForRecordCounts("student_groups", m.DB)
	if err != nil {
		return 0, 0, err
	}
	today += tod
	total += tot

	tod, tot, err = QueryForRecordCounts("student_residences", m.DB)
	if err != nil {
		return 0, 0, err
	}
	today += tod
	total += tot

	return today, total, err
}
