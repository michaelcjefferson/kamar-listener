package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

type Models struct {
	Assessments AssessmentModel
	Attendance  AttendanceModel
	Config      ConfigModel
	Logs        LogModel
	Pastoral    PastoralModel
	Results     ResultModel
	Timetables  TimetableModel
	Tokens      TokenModel
	Users       UserModel
}

func NewModels(db *sql.DB, background func(fn func())) Models {
	return Models{
		Assessments: AssessmentModel{DB: db},
		Attendance:  AttendanceModel{DB: db},
		Config:      ConfigModel{DB: db},
		Logs:        LogModel{DB: db, background: background},
		Pastoral:    PastoralModel{DB: db},
		Results:     ResultModel{DB: db},
		Timetables:  TimetableModel{DB: db},
		Tokens:      TokenModel{DB: db},
		Users:       UserModel{DB: db},
	}
}
