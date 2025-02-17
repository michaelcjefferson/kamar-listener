package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type Log struct {
	Level      string                 `json:"level"`
	Time       string                 `json:"time"`
	Message    string                 `json:"message"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Trace      string                 `json:"trace,omitempty"`
}

type LogModel struct {
	DB *sql.DB
}

func (m *LogModel) Insert(log *Log) error {
	query := `
		INSERT INTO logs (level, time, message, properties, trace)
		VALUES ($1, $2, $3, $4, $5)
	`

	jsonProps, err := json.Marshal(log.Properties)
	if err != nil {
		fmt.Println("error marshalling json when attempting to write a log to database:", err)
	}

	args := []interface{}{log.Level, log.Time, log.Message, jsonProps, log.Trace}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err = m.DB.ExecContext(ctx, query, args...)

	return err
}
