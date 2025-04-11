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
	Staff       StaffModel
	Students    StudentModel
	Subjects    SubjectModel
	Timetables  TimetableModel
	Tokens      TokenModel
	Users       UserModel
}

func NewModels(appdb, kamardb *sql.DB, background func(fn func())) Models {
	return Models{
		Assessments: AssessmentModel{DB: kamardb},
		Attendance:  AttendanceModel{DB: kamardb},
		Config:      ConfigModel{DB: appdb},
		Logs:        LogModel{DB: appdb, background: background},
		Pastoral:    PastoralModel{DB: kamardb},
		Results:     ResultModel{DB: kamardb},
		Staff:       StaffModel{DB: kamardb},
		Students:    StudentModel{DB: kamardb},
		Subjects:    SubjectModel{DB: kamardb},
		Timetables:  TimetableModel{DB: kamardb},
		Tokens:      TokenModel{DB: appdb},
		Users:       UserModel{DB: appdb},
	}
}
