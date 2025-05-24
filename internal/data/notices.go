package data

import (
	"database/sql"
)

type Notice struct {
	UUID              *string `json:"uuid"`
	DateStart         *string `json:"DateStart,omitempty"`
	DateFinish        *string `json:"DateFinish,omitempty"`
	PublishWeb        *bool   `json:"Publishweb,omitempty"`
	Level             *string `json:"Level,omitempty"`
	Subject           *string `json:"Subject,omitempty"`
	Body              *string `json:"Body,omitempty"`
	Teacher           *string `json:"Teacher,omitempty"`
	MeetingDate       *string `json:"MeetingDate,omitempty"`
	MeetingTime       *string `json:"MeetingTime,omitempty"`
	MeetingPlace      *string `json:"MeetingPlace,omitempty"`
	ListenerUpdatedAt string
}

type NoticesModel struct {
	DB *sql.DB
}

func (m *NoticesModel) InsertManyNotices(notices []Notice) error {
	// Start a transaction (tx)
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback transaction if there's an error

	stmt, err := tx.Prepare(`
	INSERT INTO notices (uuid, datestart, datefinish, publishweb, level, subject, body, teacher, meetingdate, meetingtime, meetingplace)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	ON CONFLICT(uuid) DO UPDATE SET
		datestart = excluded.datestart,
		datefinish = excluded.datefinish,
		publishweb = excluded.publishweb,
		level = excluded.level,
		subject = excluded.subject,
		body = excluded.body,
		teacher = excluded.teacher,
		meetingdate = excluded.meetingdate,
		meetingtime = excluded.meetingtime,
		meetingplace = excluded.meetingplace,
		listener_updated_at = (datetime('now'))
	;`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Insert entries in batches - extrapolate into own function
	batchSize := 100 // adjust as needed
	for i := 0; i < len(notices); i += batchSize {
		// Create a slice of 100 results
		batch := notices[i:min(i+batchSize, len(notices))]

		// Insert each entry
		for _, notice := range batch {
			_, err := stmt.Exec(notice.UUID, notice.DateStart, notice.DateFinish, notice.PublishWeb, notice.Level, notice.Subject, notice.Body, notice.Teacher, notice.MeetingDate, notice.MeetingTime, notice.MeetingPlace)
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

func (m *NoticesModel) GetNoticesCount() (int, int, error) {
	today, total, err := QueryForRecordCounts("notices", m.DB)

	return today, total, err
}
