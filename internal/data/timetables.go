package data

import "database/sql"

type Timetable struct {
	Student   int    `json:"student,omitempty"`
	UUID      string `json:"uuid,omitempty"`
	Grid      string `json:"grid,omitempty"`
	Timetable string `json:"timetable,omitempty"`
}

type TimetableModel struct {
	DB *sql.DB
}
