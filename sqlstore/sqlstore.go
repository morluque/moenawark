package sqlstore

import (
	"database/sql"
	sqlite3 "github.com/mattn/go-sqlite3"
)

// Open returns a new database connection.
// Currently uses sqlite.
func Open(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// IsConstraintError returns true if err is a constraint violation error.
func IsConstraintError(err error) bool {
	if err == nil {
		return false
	}

	if e, ok := err.(sqlite3.Error); ok {
		switch e.Code {
		case sqlite3.ErrConstraint:
			return true
		default:
			return false
		}
	}
	return false
}
