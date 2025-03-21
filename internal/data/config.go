package data

import (
	"context"
	"database/sql"
	"strconv"
	"time"
)

// type ListenerConfig struct {
// 	ServiceName      string   `json:"service_name,omitempty"`
// 	InfoURL          string   `json:"info_url,omitempty"`
// 	PrivacyStatement string   `json:"privacy_statement,omitempty"`
// 	ListenerUsername string   `json:"listener_username,omitempty"`
// 	ListenerPassword password `json:"-"`
// 	KamarIP          string   `json:"kamar_ip,omitempty"`
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

type ConfigEntry struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Type        string `json:"type"`
	Description string `json:"description"`
	UpdatedAt   string `json:"updated_at"`
}

type ListenerConfig map[string]interface{}

type ConfigModel struct {
	DB *sql.DB
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
			p := password{}
			p.hash = []byte(entry.Value)
			c[entry.Key] = p
		default:
			c[entry.Key] = entry.Value
		}
	}

	return &c
}

// Get returns a typed value from ListenerConfig with proper type assertion
func (c *ListenerConfig) GetString(key string) (string, bool) {
	if val, ok := (*c)[key]; ok {
		if str, ok := val.(string); ok {
			return str, true
		}
	}
	return "", false
}

func (c *ListenerConfig) GetBool(key string) (bool, bool) {
	if val, ok := (*c)[key]; ok {
		if b, ok := val.(bool); ok {
			return b, true
		}
	}
	return false, false
}

func (c *ListenerConfig) GetInt(key string) (int, bool) {
	if val, ok := (*c)[key]; ok {
		if i, ok := val.(int); ok {
			return i, true
		}
	}
	return 0, false
}

func (c *ListenerConfig) GetPassword(key string) (password, bool) {
	if val, ok := (*c)[key]; ok {
		if i, ok := val.(password); ok {
			return i, true
		}
	}
	return password{}, false
}

func (c *ConfigModel) GetAll() ([]ConfigEntry, error) {
	query := `SELECT key, value, type, description, updated_at FROM config;`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query)
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
		entries = append(entries, entry)
	}

	return entries, nil
}

// TODO: Add context with timeout
// GetByKey retrieves a single configuration item
func (c *ConfigModel) GetByKey(key string) (ConfigEntry, error) {
	var entry ConfigEntry
	err := c.DB.QueryRow("SELECT key, value, type, description, updated_at FROM config WHERE key = ?", key).
		Scan(&entry.Key, &entry.Value, &entry.Type, &entry.Description, &entry.UpdatedAt)
	return entry, err
}

// LoadConfig loads all configs as a typed ListenerConfig object
func (c *ConfigModel) LoadConfig() (*ListenerConfig, error) {
	entries, err := c.GetAll()
	if err != nil {
		return nil, err
	}

	return NewConfig(entries), nil
}

// TODO: Add context with timeout
// Set stores or updates a configuration item
func (c *ConfigModel) Set(entry ConfigEntry) error {
	_, err := c.DB.Exec(
		"INSERT INTO configs (key, value, type, description) VALUES (?, ?, ?, ?) "+
			"ON CONFLICT(key) DO UPDATE SET value = ?, type = ?, description = ?, updated_at = CURRENT_TIMESTAMP",
		entry.Key, entry.Value, entry.Type, entry.Description,
		entry.Value, entry.Type, entry.Description,
	)
	return err
}
