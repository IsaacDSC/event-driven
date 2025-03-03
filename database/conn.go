package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

func NewConnection(connStr string) (db *sql.DB, err error) {
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		err = fmt.Errorf("could not connect to database: %v", err)
		return
	}

	if err := db.Ping(); err != nil {
		return db, fmt.Errorf("could not ping database: %v", err)
	}

	return
}
