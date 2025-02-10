package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/mjefferson-whs/listener/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
)

// Use the json:"-" tag to prevent these fields from appearing in any output when encoded to JSON.
type User struct {
	ID        int64    `json:"id"`
	CreatedAt string   `json:"created_at"`
	Username  string   `json:"username"`
	Password  password `json:"-"`
}

// Using a pointer to plaintext allows differentiation between a password that hasn't been provided and a password that is an empty string, because nil value of a string is "" whereas nil value of a pointer is nil.
type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

func (m UserModel) Insert(user *User) error {
	query := `
		INSERT INTO users (username, password_hash)
		VALUES ($1, $2)
		RETURNING id
	`

	args := []interface{}{user.Username, user.Password.hash}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID)
	if err != nil {
		switch {
		case err.Error() == `Error while executing SQL query on database 'sms': UNIQUE constraint failed: users.email`:
			return ErrUserAlreadyExists
		default:
			return err
		}
	}

	return nil
}

func (m UserModel) GetByUsername(username string) (*User, error) {
	query := `
		SELECT id, created_at, username, password_hash
		FROM users
		WHERE username = $1
	`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Username,
		&user.Password.hash,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) GetUserCount() (int, error) {
	query := `
		SELECT COUNT(*) FROM users
	`

	var count int

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query).Scan(&count)

	return count, err
}

func (m UserModel) Update(user User) error {
	query := `
		UPDATE users
		SET username = $1, password_hash = $2
		WHERE id = $3
		RETURNING username
	`

	args := []interface{}{
		user.Username,
		user.Password.hash,
		user.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Username)
	if err != nil {
		switch {
		case err.Error() == `Error while executing SQL query on database 'sms': UNIQUE constraint failed: users.email`:
			return ErrUserAlreadyExists
		default:
			return err
		}
	}

	return nil
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Username != "", "name", "must be provided")
	v.Check(len(user.Username) <= 500, "name", "must not be more than 500 bytes long")

	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}

	// If user's password hash is ever nil, it is an issue with our codebase rather than the user, so raise a panic rather than creating a validation error message.
	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

type UserModel struct {
	DB *sql.DB
}
