package database

import (
	"database/sql"
	"fmt"
)

func NewConnection() (db *sql.DB, err error) {
	db, err = sql.Open("postgres", "user=root password=root dbname=event-driven sslmode=disable")
	if err != nil {
		err = fmt.Errorf("could not connect to database: %v", err)
		return
	}

	if err := db.Ping(); err != nil {
		return db, fmt.Errorf("could not ping database: %v", err)
	}

	return
}
