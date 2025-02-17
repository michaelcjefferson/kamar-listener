package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

type Models struct {
	Logs    LogModel
	Results ResultModel
	Tokens  TokenModel
	Users   UserModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Logs:    LogModel{DB: db},
		Results: ResultModel{DB: db},
		Tokens:  TokenModel{DB: db},
		Users:   UserModel{DB: db},
	}
}
