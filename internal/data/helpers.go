package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// PASSWORDS
// Using a pointer to plaintext allows differentiation between a password that hasn't been provided and a password that is an empty string, because nil value of a string is "" whereas nil value of a pointer is nil.
type Password struct {
	plaintext *string
	hash      []byte
}

// Hash plaintext password from form, and set both plaintext and hashed passwords as values on User struct
func (p *Password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	// Set p.plaintext to be the value of the provided plaintextPassword, rather than the actual plaintextPassword in memory
	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

// Check whether the provided plaintextPassword, once hashed, matches the hashed password attached to the user struct. eg. on sign-in, GetByUsername() is called to retrieve a user struct matching the provided username, and their associated password.hash is compared below with the plaintext password provided
func (p *Password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		// If passwords don't match, return false but no error
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

// Provide access to password hash
func (p *Password) Hash() []byte {
	return p.hash
}

// OTHER HELPERS
// Attempts to insert a large set of data in batches (for improved performance)
// func BatchInsert(statement string, ) error {

// }

func QueryForRecordCounts(tableName string, db *sql.DB) (int, int, error) {
	var today, total int
	query := fmt.Sprintf(`
		SELECT
		COUNT(*) AS total_count,
		COUNT(CASE WHEN listener_updated_at >= datetime('now', '-1 day') THEN 1 END) AS todays_count
		FROM %s;
	`, tableName)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := db.QueryRowContext(ctx, query).Scan(&today, &total)

	return today, total, err
}

// Tries to read and convert an any value (eg. the ones in log.Properties) to an int value - returns an int (or 0 on failure) and a bool (ok)
func ToInt(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int8, int16, int32, int64:
		return int(reflect.ValueOf(v).Int()), true
	case uint, uint8, uint16, uint32, uint64:
		return int(reflect.ValueOf(v).Uint()), true
	case float32, float64:
		return int(reflect.ValueOf(v).Float()), true
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i, true
		}
	}
	return 0, false
}

// Creates a string of comma separated ? placeholders based on the int provided - good for dynamically adding placeholders to a query based on the number of parameters
func placeholders(n int) string {
	return strings.TrimSuffix(strings.Repeat("?,", n), ",")
}

// Utility function for the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
