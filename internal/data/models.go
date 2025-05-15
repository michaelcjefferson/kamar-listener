package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

type Models struct {
	Assessments    AssessmentModel
	Attendance     AttendanceModel
	ClassEfforts   ClassEffortsModel
	Config         ConfigModel
	ListenerEvents ListenerEventsModel
	Logs           LogModel
	Pastoral       PastoralModel
	Recognitions   RecognitionsModel
	Results        ResultModel
	Staff          StaffModel
	Students       StudentModel
	Subjects       SubjectModel
	Timetables     TimetableModel
	Tokens         TokenModel
	Users          UserModel
	Widgets        WidgetModel
}

func NewModels(appdb, kamardb *sql.DB, background func(fn func())) Models {
	return Models{
		Assessments:    AssessmentModel{DB: kamardb},
		Attendance:     AttendanceModel{DB: kamardb},
		ClassEfforts:   ClassEffortsModel{DB: kamardb},
		Config:         ConfigModel{DB: appdb},
		ListenerEvents: ListenerEventsModel{DB: appdb},
		Logs:           LogModel{DB: appdb, background: background},
		Pastoral:       PastoralModel{DB: kamardb},
		Recognitions:   RecognitionsModel{DB: kamardb},
		Results:        ResultModel{DB: kamardb},
		Staff:          StaffModel{DB: kamardb},
		Students:       StudentModel{DB: kamardb},
		Subjects:       SubjectModel{DB: kamardb},
		Timetables:     TimetableModel{DB: kamardb},
		Tokens:         TokenModel{DB: appdb},
		Users:          UserModel{DB: appdb},
		Widgets:        WidgetModel{DB: appdb},
	}
}
