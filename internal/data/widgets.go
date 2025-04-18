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
	RecordsToday    int64
	TotalErrors     int64
	TotalLogs       int64
	TotalRecords    int64
}

type WidgetModel struct {
	DB *sql.DB
}
