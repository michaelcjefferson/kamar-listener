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
	DateBirth          string          `json:"datebirth,omitempty"`
	LeavingDate        string          `json:"leavingdate,omitempty"`
	StartingDate       string          `json:"startingdate,omitempty"`
	ESLGUID            any             `json:"eslguid,omitempty"`
	MOENumber          any             `json:"moenumber,omitempty"`
	PhotocopierID      any             `json:"photocopierid,omitempty"`
	RegistrationNumber any             `json:"registrationnumber,omitempty"`
	Custom             json.RawMessage `json:"custom,omitempty"`
}

type StaffModel struct {
	DB *sql.DB
}
