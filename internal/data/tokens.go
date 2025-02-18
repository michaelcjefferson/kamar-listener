package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"

	"github.com/mjefferson-whs/listener/internal/validator"
)

// JSON tags dictate which fields will be encoded into JSON for the client, and the names  of their corresponding keys ("token" is more meaningful for the client than "plaintext")
type Token struct {
	Plaintext string `json:"token"`
	Hash      []byte `json:"-"`
	UserID    int64  `json:"-"`
	Expiry    string `json:"expiry"`
}

// ttl (time-to-live) is added to time.Now to create a token expiry
func generateToken(userID int64, ttl time.Duration) (*Token, error) {
	t := time.Now().Add(ttl).UTC().Format(time.RFC3339)

	token := &Token{
		UserID: userID,
		Expiry: t,
	}

	// crypto/rand.Read fills a byte slice with random bytes from CSPRNG. This token will have an entropy (randomness) of 16 bytes. Base32 encoding means the plaintext token itself will be 26 bytes long.
	randomBytes := make([]byte, 16)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	// Omit the possible = sign at the end by using WithPadding(base32.NoPadding)
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	hash := sha256.Sum256([]byte(token.Plaintext))
	// Convert the array returned by Sum256() to a slice using [:], so that it's easier to work with.
	token.Hash = hash[:]

	return token, nil
}

func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}

type TokenModel struct {
	DB *sql.DB
}

// Whenever a token is created, the next step will be for it to be stored in the tokens table on the database. So, call m.Insert() as part of the token creation process.
func (m TokenModel) New(userID int64, ttl time.Duration) (*Token, error) {
	token, err := generateToken(userID, ttl)
	if err != nil {
		return nil, err
	}

	err = m.Insert(token)
	return token, err
}

func (m TokenModel) Insert(token *Token) error {
	query := `
		INSERT INTO tokens (hash, user_id, expiry)
		VALUES ($1, $2, $3)`

	args := []interface{}{token.Hash, token.UserID, token.Expiry}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}

func (m TokenModel) DeleteAllForUser(userID int64) (int64, error) {
	query := `
		DELETE FROM tokens
		WHERE user_id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, userID)
	r, _ := result.RowsAffected()
	return r, err
}

func (m TokenModel) DeleteExpiredTokens() (int64, error) {
	query := `
		DELETE FROM tokens
		WHERE expiry < strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query)
	r, _ := result.RowsAffected()
	return r, err
}
