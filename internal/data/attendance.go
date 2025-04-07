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
