package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type Log struct {
	ID         int                    `json:"id,omitempty"`
	Level      string                 `json:"level"`
	Time       string                 `json:"time"`
	Message    string                 `json:"message"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Trace      string                 `json:"trace,omitempty"`
	UserID     int                    `json:"userid,omitempty"`
}

type LogModel struct {
	DB *sql.DB
}

// Insert an individual log into the logs table, as well as a reference to the log message for text search into the logs_fts table
func (m *LogModel) Insert(log *Log) error {
	query := `
		INSERT INTO logs (level, time, message, properties, trace)
		VALUES ($1, $2, $3, $4, $5);

		INSERT INTO logs_fts (rowid, message)
		VALUES (last_insert_rowid(), $6);
	`

	jsonProps, err := json.Marshal(log.Properties)
	if err != nil {
		fmt.Println("error marshalling json when attempting to write a log to database:", err)
	}

	args := []interface{}{log.Level, log.Time, log.Message, jsonProps, log.Trace, log.Message}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err = m.DB.ExecContext(ctx, query, args...)

	return err
}

func (m *LogModel) GetAll(searchTerm string, filters Filters) ([]*Log, Metadata, error) {
	// It's not possible to interpolate ORDER BY column or direction into an SQL query using $ values, so use Sprintf to create the query.
	// Subquery SELECT COUNT(*) FROM logs_fts provides the total number of rows returned by the query, and appends it to each row in the location specified (in this case, it is the last column of each row, i.e. after trace)
	// The JOIN also uses the logs_fts table to perform a search for messages that contain the provided searchTerm
	query := fmt.Sprintf(`
		SELECT logs.id, logs.time, logs.level, logs.message, logs.properties, logs.userID,
			(SELECT COUNT(*) FROM logs_fts WHERE logs_fts MATCH ?)
			AS total_count
		FROM logs
		JOIN logs_fts ON logs.id = logs_fts.rowid
		WHERE logs_fts MATCH ?
		ORDER BY %s %s, id ASC
		LIMIT ? OFFSET ?
	`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{searchTerm, searchTerm, filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	// Make sure result from QueryContext is closed before returning from function
	defer rows.Close()

	totalRecords := 0
	logs := []*Log{}

	for rows.Next() {
		var log Log

		err := rows.Scan(
			&log.ID,
			&log.Time,
			&log.Level,
			&log.Message,
			&log.Properties,
			&log.UserID,
			&totalRecords,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		logs = append(logs, &log)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return logs, metadata, nil
}
