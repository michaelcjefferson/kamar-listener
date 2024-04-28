package data

import "database/sql"

type Models struct {
	Results ResultModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Results: ResultModel{DB: db},
	}
}
