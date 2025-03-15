package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Log struct {
	ID         int                    `json:"id,omitempty"`
	Level      string                 `json:"level"`
	Time       string                 `json:"time"`
	Message    string                 `json:"message"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Trace      string                 `json:"trace,omitempty"`
	UserID     int                    `json:"user_id,omitempty"`
}

type LogModel struct {
	DB *sql.DB
}

// Insert an individual log into the logs table, as well as a reference to the log message for text search into the logs_fts table
func (m *LogModel) Insert(log *Log) error {
	// Get user_id if it exists to add to logs_metadata table
	var userID int
	if i, ok := ToInt(log.Properties["user_id"]); ok {
		userID = i
	}

	query := `
		INSERT INTO logs (level, time, message, properties, trace)
		VALUES ($1, $2, $3, $4, $5);

		INSERT INTO logs_fts (rowid, message)
		VALUES (last_insert_rowid(), $6);

		INSERT INTO logs_metadata (type, level, count)
		VALUES ("level", $7, 1)
		ON CONFLICT(level) DO UPDATE
		SET count=count+1;
	`

	jsonProps, err := json.Marshal(log.Properties)
	if err != nil {
		fmt.Println("error marshalling json when attempting to write a log to database:", err)
	}

	args := []interface{}{log.Level, log.Time, log.Message, jsonProps, log.Trace, log.Message, log.Level}

	if userID > 0 {
		query += `
			INSERT INTO logs_metadata (type, user_id, count)
			VALUES ("user_id", $8, 1)
			ON CONFLICT(user_id) DO UPDATE
			SET count=count+1;
		`

		args = append(args, userID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err = m.DB.ExecContext(ctx, query, args...)

	return err
}

func (m *LogModel) GetForID(id int) (*Log, error) {
	// Seeing as all logs that exist in the database will have a positive integer as their id, check that the request id is valid before querying database to prevent wasted queries
	// IMPORTANT: Though it may seem like a good idea to use an unsigned int here (seeing as id will never be negative), it is more important that the types we use in our code align with the types available in our database. SQLite doesn't have unsigned ints, so use a standard int which is a reflection of SQLite's integer type
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, level, time, message, properties, trace, user_id
		FROM logs
		WHERE id = ?
	`

	// Declare Log struct to hold data returned by query
	var log Log
	// Decalre propertiesJSON string to hold the properties value returned by the query, so that it can be unmarshalled before being attached to the Log struct
	var propertiesJSON string

	// Create an empty context (Background()) with a 3 second timeout. The timeout countdown will begin as soon as it is created
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	// IMPORTANT: defer the cancel() function returned by context.WithTimeout(), so that in case of a successful request, the context is cancelled and resources are freed up before the request returns
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&log.ID,
		&log.Level,
		&log.Time,
		&log.Message,
		&propertiesJSON,
		&log.Trace,
		&log.UserID,
	)

	// Unmarshal properties into log struct
	if propertiesJSON != "" {
		err = json.Unmarshal([]byte(propertiesJSON), &log.Properties)
	}

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &log, nil
}

func (m *LogModel) GetAll(filters Filters) ([]*Log, Metadata, *LogsMetadata, error) {
	// It's not possible to interpolate ORDER BY column or direction into an SQL query using $ values, so use Sprintf to create the query.
	// Subquery SELECT COUNT(*) FROM logs_fts provides the total number of rows returned by the query, and appends it to each row in the location specified (in this case, it is the last column of each row, i.e. after trace)
	// The JOIN also uses the logs_fts table to perform a search for messages that contain the provided searchTerm
	// query := fmt.Sprintf(`
	// 	SELECT logs.id, logs.level, logs.time, logs.message, logs.properties, logs.user_id,
	// 		(SELECT COUNT(*) FROM logs_fts WHERE logs_fts MATCH ?)
	// 		AS total_count
	// 	FROM logs
	// 	JOIN logs_fts ON logs.id = logs_fts.rowid
	// 	WHERE logs_fts MATCH ?
	// 	AND level = ?
	// 	AND user_id = ?
	// 	ORDER BY %s %s, id ASC
	// 	LIMIT ? OFFSET ?
	// `, filters.sortColumn(), filters.sortDirection())

	var queryBuilder strings.Builder
	args := []interface{}{}

	queryBuilder.WriteString(`
		SELECT logs.id, logs.level, logs.time, logs.message, logs.properties, logs.user_id,
			(SELECT COUNT(*) FROM logs JOIN logs_fts ON logs.id = logs_fts.rowid WHERE 1=1
	`)

	getAllLogsFilterQueryHelper(&queryBuilder, &args, filters)

	queryBuilder.WriteString(") AS total_count FROM logs JOIN logs_fts ON logs.id = logs_fts.rowid WHERE 1=1")

	getAllLogsFilterQueryHelper(&queryBuilder, &args, filters)

	queryBuilder.WriteString(fmt.Sprintf(" ORDER BY %s %s, id DESC LIMIT ? OFFSET ?", filters.sortColumn(), filters.sortDirection()))
	args = append(args, filters.limit(), filters.offset())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, Metadata{}, nil, err
	}

	// Make sure result from QueryContext is closed before returning from function
	defer rows.Close()

	totalRecords := 0
	logs := []*Log{}

	for rows.Next() {
		var log Log
		var propertiesJSON string
		var userID *int

		err := rows.Scan(
			&log.ID,
			&log.Level,
			&log.Time,
			&log.Message,
			&propertiesJSON,
			&userID,
			&totalRecords,
		)
		if err != nil {
			return nil, Metadata{}, nil, err
		}

		// Unmarshal properties into log struct
		if propertiesJSON != "" {
			err = json.Unmarshal([]byte(propertiesJSON), &log.Properties)
		}
		if err != nil {
			return nil, Metadata{}, nil, err
		}

		// Attach value of user_id if it isn't nil
		if userID != nil {
			log.UserID = *userID
		}

		logs = append(logs, &log)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, nil, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	logsMetadata, err := GetLogsMetadata(m)
	if err != nil {
		return logs, metadata, nil, err
	}

	return logs, metadata, logsMetadata, nil
}

func getAllLogsFilterQueryHelper(q *strings.Builder, args *[]interface{}, filters Filters) {
	if filters.LogFilters.Message != "" {
		q.WriteString(" AND logs_fts MATCH ?")
		*args = append(*args, filters.LogFilters.Message)
	}
	if len(filters.LogFilters.Level) > 0 {
		qp := fmt.Sprintf(" AND level IN (%s)", placeholders(len(filters.LogFilters.Level)))
		q.WriteString(qp)
		// q.WriteString(" AND level = ?")
		for _, val := range filters.LogFilters.Level {
			*args = append(*args, val)
		}
	}
	if len(filters.LogFilters.UserID) > 0 {
		// q.WriteString(" AND user_id = ?")
		qp := fmt.Sprintf(" AND user_id IN (%s)", placeholders(len(filters.LogFilters.UserID)))
		q.WriteString(qp)
		for _, val := range filters.LogFilters.UserID {
			*args = append(*args, val)
		}
	}
}

// DELETE
// tx, err := db.Begin()
// if err != nil {
//     log.Fatal(err)
// }
// defer tx.Rollback()

// _, err = tx.Exec(`
//     UPDATE logs_metadata
//     SET count = count - 1
//     WHERE (level = ? OR user_id = ?)
//     AND count > 0;
// `, level, user_id)
// if err != nil {
//     log.Fatal(err)
// }

// _, err = tx.Exec(`
//     DELETE FROM logs_metadata
//     WHERE count = 0;
// `)
// if err != nil {
//     log.Fatal(err)
// }

// err = tx.Commit()
// if err != nil {
//     log.Fatal(err)
// }

func GetLogsMetadata(m *LogModel) (*LogsMetadata, error) {
	query := `
		SELECT type, level, user_id, count FROM logs_metadata;
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return &LogsMetadata{}, err
	}

	logsMetadata := NewLogsMetadata()

	for rows.Next() {
		var logType string
		var level *string
		var userID *int
		var count int

		err := rows.Scan(
			&logType,
			&level,
			&userID,
			&count,
		)

		if err != nil {
			return &LogsMetadata{}, err
		}

		switch {
		case logType == "level" && level != nil:
			logsMetadata.Levels[*level] = count
		case logType == "user_id" && userID != nil:
			logsMetadata.Users[*userID] = count
		default:
			return &LogsMetadata{}, errors.New(fmt.Sprintf("error adding logtype %v to database", logType))
		}
	}

	if err = rows.Err(); err != nil {
		return &LogsMetadata{}, err
	}

	return &logsMetadata, nil
}
