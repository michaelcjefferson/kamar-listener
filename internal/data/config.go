package data

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/michaelcjefferson/kamar-listener/internal/validator"
)

// type ListenerConfig struct {
// 	ServiceName      string   `json:"service_name,omitempty"`
// 	InfoURL          string   `json:"info_url,omitempty"`
// 	PrivacyStatement string   `json:"privacy_statement,omitempty"`
// 	ListenerUsername string   `json:"listener_username,omitempty"`
// 	ListenerPassword password `json:"-"`
// 	Details          bool     `json:"details,omitempty"`
// 	Passwords        bool     `json:"passwords,omitempty"`
// 	Photos           bool     `json:"photos,omitempty"`
// 	Groups           bool     `json:"groups,omitempty"`
// 	Awards           bool     `json:"awards,omitempty"`
// 	Timetables       bool     `json:"timetables,omitempty"`
// 	Attendance       bool     `json:"attendance,omitempty"`
// 	Assessments      bool     `json:"assessments,omitempty"`
// 	Pastoral         bool     `json:"pastoral,omitempty"`
// 	LearningSupport  bool     `json:"learning_support,omitempty"`
// 	Subjects         bool     `json:"subjects,omitempty"`
// 	Notices          bool     `json:"notices,omitempty"`
// 	Calendar         bool     `json:"calendar,omitempty"`
// 	Bookings         bool     `json:"bookings,omitempty"`
// }

// TODO: Add port
// "calendar" is an option from KAMAR, but it isn't particularly useful and its data structure is messy - to allow calendars to be received from KAMAR, a new data structure needs to be built and implemented before adding "calendar" to this list
// UPDATE: calendars should be fine - it's just a long string - add later
var ConfigKeySafeList = []string{"service_name", "info_url", "privacy_statement", "listener_username", "listener_password", "details", "passwords", "photos", "groups", "awards", "timetables", "attendance", "assessments", "pastoral", "learningsupport", "recognitions", "classefforts", "subjects", "notices", "bookings", "calendar"}

type ConfigEntry struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

// type ConfigUpdateRequest struct {
// 	Key   string `json:"key"`
// 	Value string `json:"value"`
// 	Type  string `json:"type"`
// }

type ListenerConfig map[string]any

type ConfigModel struct {
	DB *sql.DB
}

func ValidateConfigKey(v *validator.Validator, key string) {
	v.Check(validator.In(key, ConfigKeySafeList...), "key", "invalid key value")
}

func ValidateConfigValue(v *validator.Validator, key, value, valueType string) {
	v.Check(key != "", key, "must not be empty")
	switch valueType {
	case "int":
		v.Check(validator.StringIsInt(value), key, "must be convertable to integer type")
	case "bool":
		v.Check(validator.StringIsBool(value), key, "must be convertable to boolean type")
	}
}

func ValidateConfigUpdate(v *validator.Validator, config ConfigEntry) {
	ValidateConfigKey(v, config.Key)
	ValidateConfigValue(v, config.Key, config.Value, config.Type)
}

// NewConfig creates a Config from ConfigEntries
func NewConfig(entries []ConfigEntry) *ListenerConfig {
	c := make(ListenerConfig)

	for _, entry := range entries {
		switch entry.Type {
		case "bool":
			c[entry.Key], _ = strconv.ParseBool(entry.Value)
		case "int":
			c[entry.Key], _ = strconv.Atoi(entry.Value)
		case "password":
			p := Password{}
			p.hash = []byte(entry.Value)
			c[entry.Key] = p
		default:
			c[entry.Key] = entry.Value
		}
	}

	return &c
}

// Get returns a typed value from ListenerConfig with proper type assertion, and returns its zero value on failure
// TODO: These are likely unnecessary (actually likely useful for getting app config)
func (c *ListenerConfig) GetString(key string) (string, bool) {
	if val, ok := (*c)[key]; ok {
		if str, ok := val.(string); ok {
			return str, true
		}
	}
	return "", false
}

func (c *ListenerConfig) GetBool(key string) bool {
	if val, ok := (*c)[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func (c *ListenerConfig) GetInt(key string) (int, bool) {
	if val, ok := (*c)[key]; ok {
		if i, ok := val.(int); ok {
			return i, true
		}
	}
	return 0, false
}

func (c *ListenerConfig) GetPassword(key string) (Password, bool) {
	if val, ok := (*c)[key]; ok {
		if i, ok := val.(Password); ok {
			return i, true
		}
	}
	return Password{}, false
}

func (m *ConfigModel) GetAll() ([]ConfigEntry, error) {
	query := `SELECT key, value, type, description, updated_at FROM config;`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []ConfigEntry
	for rows.Next() {
		var entry ConfigEntry
		err := rows.Scan(&entry.Key, &entry.Value, &entry.Type, &entry.Description, &entry.UpdatedAt)
		if err != nil {
			return nil, err
		}
		// if entry.Key == "listener_password" {
		// 	entry.Value = ""
		// }
		entries = append(entries, entry)
	}

	return entries, nil
}

// GetAuth returns the username and (hashed) password stored for KAMAR directory service
func (m *ConfigModel) GetAuth() (*User, error) {
	var KAMARUser User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, "SELECT value FROM config WHERE key = 'listener_username';").Scan(&KAMARUser.Username)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	err = m.DB.QueryRowContext(ctx, "SELECT value FROM config WHERE key = 'listener_password';").Scan(&KAMARUser.Password.hash)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &KAMARUser, nil
}

// GetByKey retrieves a single configuration item
func (m *ConfigModel) GetByKey(key string) (ConfigEntry, error) {
	var entry ConfigEntry

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, "SELECT key, value, type, description, updated_at FROM config WHERE key = ?", key).
		Scan(&entry.Key, &entry.Value, &entry.Type, &entry.Description, &entry.UpdatedAt)
	return entry, err
}

// LoadConfig loads all configs as a typed ListenerConfig object
func (m *ConfigModel) LoadConfig() (*ListenerConfig, error) {
	entries, err := m.GetAll()
	if err != nil {
		return nil, err
	}

	return NewConfig(entries), nil
}

// Set stores or updates a configuration item
func (m *ConfigModel) Set(entry ConfigEntry) error {
	// query := `
	// 	INSERT INTO config (key, value, type) VALUES (?, ?, ?)
	// 	ON CONFLICT(key) DO UPDATE SET value = ?, type = ?, updated_at = CURRENT_TIMESTAMP
	// `
	query := `
		INSERT INTO config (key, value, type, description) VALUES (?, ?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = ?, type = ?, description = ?, updated_at = CURRENT_TIMESTAMP
	`

	args := []any{
		entry.Key,
		entry.Value,
		entry.Type,
		entry.Description,
		entry.Value,
		entry.Type,
		entry.Description,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}
