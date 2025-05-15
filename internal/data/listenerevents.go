package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var ErrMissingReqType = errors.New("missing request type")

type ListenerEvent struct {
	ID            int       `json:"id,omitempty"`
	ReqType       string    `json:"req_type"`
	WasSuccessful bool      `json:"was_successful"`
	Message       string    `json:"message"`
	Time          time.Time `json:"time"`
}

type ListenerEventsModel struct {
	DB *sql.DB
}

func (m *ListenerEventsModel) Insert(event *ListenerEvent) error {
	if event.ReqType == "" {
		return ErrMissingReqType
	}

	query := `
		INSERT INTO listener_events (req_type, was_successful, message)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	args := []any{event.ReqType, event.WasSuccessful, event.Message}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&event.ID)
	if err != nil {
		return err
	}

	return nil
}

func (m *ListenerEventsModel) GetAll() ([]ListenerEvent, error) {
	query := `
		SELECT id, req_type, was_successful, message, time FROM listener_events
		ORDER BY time DESC;
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	// Make sure result from QueryContext is closed before returning from function
	defer rows.Close()

	var events []ListenerEvent

	for rows.Next() {
		var event ListenerEvent
		var wasSuccessful int
		var timeStr string

		err := rows.Scan(
			&event.ID,
			&event.ReqType,
			&wasSuccessful,
			&event.Message,
			&timeStr,
		)
		if err != nil {
			return nil, err
		}

		// Attach value of user_id if it isn't nil
		if wasSuccessful > 0 {
			event.WasSuccessful = true
		} else {
			event.WasSuccessful = false
		}

		t, err := time.Parse("2006-01-02 15:04:05", timeStr)
		if err != nil {
			return nil, err
		}
		event.Time = t

		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (m *ListenerEventsModel) GetMostRecentCheckAndInsert() ([]ListenerEvent, error) {
	query := `
		WITH latest_check AS (
			SELECT * FROM listener_events
			WHERE req_type = 'check' AND was_successful = 1
			ORDER BY time DESC
			LIMIT 1
		),
		latest_insert AS (
			SELECT * FROM listener_events
			WHERE req_type = 'insert' AND was_successful = 1
			ORDER BY time DESC
			LIMIT 1
		)
		SELECT * FROM latest_check
		UNION ALL
		SELECT * FROM latest_insert;
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	// Make sure result from QueryContext is closed before returning from function
	defer rows.Close()

	var events []ListenerEvent

	for rows.Next() {
		var event ListenerEvent
		var wasSuccessful int
		var timeStr string

		err := rows.Scan(
			&event.ID,
			&event.ReqType,
			&wasSuccessful,
			&event.Message,
			&timeStr,
		)
		if err != nil {
			return nil, err
		}

		// Attach value of user_id if it isn't nil
		if wasSuccessful > 0 {
			event.WasSuccessful = true
		} else {
			event.WasSuccessful = false
		}

		t, err := time.Parse("2006-01-02 15:04:05", timeStr)
		if err != nil {
			return nil, err
		}
		event.Time = t

		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}
