package data

import (
	"errors"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// PASSWORDS
// Using a pointer to plaintext allows differentiation between a password that hasn't been provided and a password that is an empty string, because nil value of a string is "" whereas nil value of a pointer is nil.
type password struct {
	plaintext *string
	hash      []byte
}

// Hash plaintext password from form, and set both plaintext and hashed passwords as values on User struct
func (p *password) Set(plaintextPassword string) error {
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
func (p *password) Matches(plaintextPassword string) (bool, error) {
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

// OTHER HELPERS
// Tries to read and convert an interface{} value (eg. the ones in log.Properties) to an int value - returns an int (or 0 on failure) and a bool (ok)
func ToInt(value interface{}) (int, bool) {
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
