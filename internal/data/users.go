package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"github.com/mjefferson-whs/listener/internal/validator"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
)

// For clients that have not provided an Authentication token as an Authorization header, allowing them to make user requests without being authenticated.
var AnonymousUser = &User{}

// Use the json:"-" tag to prevent these fields from appearing in any output when encoded to JSON.
type User struct {
	ID                  int64    `json:"id"`
	CreatedAt           string   `json:"created_at"`
	LastAuthenticatedAt string   `json:"last_authenticated_at"`
	Username            string   `json:"username"`
	Password            password `json:"-"`
}

// Any user object can call this function which will return true if the user object doesn't have a username, password, and ID associated with it.
func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

func ValidateUsername(v *validator.Validator, username string) {
	v.Check(username != "", "name", "must be provided")
	v.Check(len(username) <= 500, "name", "must not be more than 500 bytes long")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	ValidateUsername(v, user.Username)

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
		case err.Error() == `Error while executing SQL query on database 'sms': UNIQUE constraint failed: users.username`:
			return ErrUserAlreadyExists
		default:
			return err
		}
	}

	return nil
}

func (m UserModel) GetAll() ([]*User, error) {
	query := `
		SELECT id, created_at, last_authenticated_at, username FROM users;
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	// Make sure result from QueryContext is closed before returning from function
	defer rows.Close()

	users := []*User{}

	for rows.Next() {
		var user User

		err := rows.Scan(
			&user.ID,
			&user.CreatedAt,
			&user.LastAuthenticatedAt,
			&user.Username,
		)
		if err != nil {
			return nil, err
		}

		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
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

func (m UserModel) GetForToken(tokenPlaintext string) (*User, string, error) {
	// This returns an array ([32]byte, specified length) rather than a slice ([]byte, unspecified length)
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	// INNER JOIN returns only the rows in the inner (overlapping) section of the Venn diagram created when users and tokens are joined on id. i.e., only rows with a matching id/user_id in both tables will exist in the join table.
	query := `
		SELECT users.id, users.created_at, users.username, users.password_hash, tokens.expiry
		FROM users
		INNER JOIN tokens
		ON users.id = tokens.user_id
		WHERE tokens.hash = $1
		AND tokens.expiry > datetime('now')`

	// Use [:] to convert the tokenHash [32]byte to a []byte. This is to match with SQLite's blob type, which tokens are stored as.
	args := []interface{}{tokenHash[:]}

	var user User
	var tokenExpiry string

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Username,
		&user.Password.hash,
		&tokenExpiry,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, "", ErrRecordNotFound
		default:
			return nil, "", err
		}
	}

	return &user, tokenExpiry, nil
}

// Get number of users registered in database
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
		case err.Error() == `Error while executing SQL query on database 'sms': UNIQUE constraint failed: users.username`:
			return ErrUserAlreadyExists
		default:
			return err
		}
	}

	return nil
}
