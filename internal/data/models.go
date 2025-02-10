package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

type Models struct {
	Results ResultModel
	Users   UserModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Results: ResultModel{DB: db},
		Users:   UserModel{DB: db},
	}
}
