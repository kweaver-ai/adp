// Package logics contains business logic implementations.
package logics

import (
	"database/sql"
)

var (
	DB *sql.DB
)

func SetDB(db *sql.DB) {
	DB = db
}
