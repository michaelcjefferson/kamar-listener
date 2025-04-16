package data

import (
	"database/sql"
	"time"
)

type WidgetData struct {
	LastUpdated time.Time
}

type WidgetModel struct {
	DB *sql.DB
}
