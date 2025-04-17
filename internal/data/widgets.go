package data

import (
	"database/sql"
	"time"
)

type WidgetData struct {
	LastUpdated  time.Time
	TotalRecords int64
}

type WidgetModel struct {
	DB *sql.DB
}
