package data

import (
	"database/sql"
	"time"
)

type WidgetData struct {
	DBSize          float64
	LastCheckTime   time.Time
	LastInsertTime  time.Time
	RecentErrorLogs []*Log
	RecentLogs      []*Log
	RecordsToday    int
	TotalErrors     int
	TotalLogs       int
	TotalRecords    int
}

type WidgetModel struct {
	DB *sql.DB
}
