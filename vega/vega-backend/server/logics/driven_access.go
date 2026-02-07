// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

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
