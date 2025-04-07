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
	Demerits       any    `json:"demerits,omitempty"`
	Dateevent      string `json:"dateevent,omitempty"`
	Timeevent      string `json:"timeevent,omitempty"`
	Datedue        string `json:"datedue,omitempty"`
	Duestatus      string `json:"duestatus,omitempty"`
}

type PastoralModel struct {
	DB *sql.DB
}
