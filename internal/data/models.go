package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

type Models struct {
	Config  ConfigModel
	Logs    LogModel
	Results ResultModel
	Tokens  TokenModel
	Users   UserModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Config:  ConfigModel{DB: db},
		Logs:    LogModel{DB: db},
		Results: ResultModel{DB: db},
		Tokens:  TokenModel{DB: db},
		Users:   UserModel{DB: db},
	}
}
