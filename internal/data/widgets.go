package data

import (
	"database/sql"
	"time"
)

type WidgetData struct {
	CountByType     map[string]int
	DBSize          float64
	IP              string
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
