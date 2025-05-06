package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var ErrMissingReqType = errors.New("missing request type")

type ListenerEvent struct {
	ID            int    `json:"id,omitempty"`
	ReqType       string `json:"req_type"`
	WasSuccessful bool   `json:"was_successful"`
	Time          string `json:"time"`
}

type ListenerEventsModel struct {
	DB *sql.DB
}

func (m *ListenerEventsModel) Insert(event *ListenerEvent) error {
	if event.ReqType == "" {
		return ErrMissingReqType
	}

	query := `
		INSERT INTO listener_events (req_type, was_successful)
		VALUES ($1, $2)
		RETURNING id
	`

	args := []any{event.ReqType, event.WasSuccessful}

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
		SELECT id, req_type, was_successful, time FROM listener_events
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

		err := rows.Scan(
			&event.ID,
			&event.ReqType,
			&wasSuccessful,
			&event.Time,
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

		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}
