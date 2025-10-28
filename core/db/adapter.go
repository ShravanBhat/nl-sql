package db

import (
	"database/sql"
	"nlsql/models"
)

type DBAdapter interface {
	GetConnectionString(config models.DBConfig) string
	GetSchema(db *sql.DB) (string, error)
}

