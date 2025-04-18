package data

import "database/sql"

type Pastoral struct {
	ID             int    `json:"id,omitempty"`
	Nsn            string `json:"nsn,omitempty"`
	Type           string `json:"type,omitempty"`
	Ref            int    `json:"ref,omitempty"`
	Reason         string `json:"reason,omitempty"`
	ReasonPB       any    `json:"reasonPB,omitempty"`
	Motivation     any    `json:"motivation,omitempty"`
	MotivationPB   any    `json:"motivationPB,omitempty"`
	Location       any    `json:"location,omitempty"`
	LocationPB     any    `json:"locationPB,omitempty"`
	Othersinvolved any    `json:"othersinvolved,omitempty"`
	Action1        string `json:"action1,omitempty"`
	Action2        any    `json:"action2,omitempty"`
	Action3        any    `json:"action3,omitempty"`
	ActionPB1      any    `json:"actionPB1,omitempty"`
	ActionPB2      any    `json:"actionPB2,omitempty"`
	ActionPB3      any    `json:"actionPB3,omitempty"`
	Teacher        string `json:"teacher,omitempty"`
	Points         int    `json:"points,omitempty"`
	Demerits       int    `json:"demerits,omitempty"`
	Dateevent      string `json:"dateevent,omitempty"`
	Timeevent      string `json:"timeevent,omitempty"`
	Datedue        string `json:"datedue,omitempty"`
	Duestatus      string `json:"duestatus,omitempty"`
}

type PastoralModel struct {
	DB *sql.DB
}

func (m *PastoralModel) InsertManyPastoral(pastoral []Pastoral) error {
	// Start a transaction (tx)
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback transaction if there's an error

	stmt, err := tx.Prepare(`
	INSERT INTO pastoral (id, nsn, type, ref, reason, reason_pb, motivation, motivation_pb, location, location_pb, others_involved, action1, action2, action3, action_pb1, action_pb2, action_pb3, teacher, points, demerits, dateevent, timeevent, datedue, duestatus) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Insert entries in batches - extrapolate into own function
	batchSize := 100 // adjust as needed
	for i := 0; i < len(pastoral); i += batchSize {
		// Create a slice of 100 results
		batch := pastoral[i:min(i+batchSize, len(pastoral))]

		// Insert each entry
		for _, p := range batch {
			_, err := stmt.Exec(p.ID, p.Nsn, p.Type, p.Ref, p.Reason, p.ReasonPB, p.Motivation, p.MotivationPB, p.Location, p.LocationPB, p.Othersinvolved, p.Action1, p.Action2, p.Action3, p.ActionPB1, p.ActionPB2, p.ActionPB3, p.Teacher, p.Points, p.Demerits, p.Dateevent, p.Timeevent, p.Datedue, p.Duestatus)
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

func (m *PastoralModel) GetPastoralCount() (int, int, error) {
	today, total, err := QueryForRecordCounts("pastoral", m.DB)

	return today, total, err
}
