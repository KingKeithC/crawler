package main

import (
	"database/sql"
	"fmt"

	// Necessary for SQL
	_ "github.com/lib/pq"
)

// InitDB creates and tests a connection to the SQL DB
func InitDB(conStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", conStr)
	if err != nil {
		return nil, fmt.Errorf("error %v creating sql.DB object with"+
			" connection string %s", err, conStr)
	}

	// Test that the DB connection works
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error %v pinging DB %v", err, db)
	}

	return db, nil
}

// CreateSchema connects to the DB and ensures that the schema exists
func CreateSchema(db *sql.DB) error {
	tableSchema := `
	CREATE TABLE urls (
		id serial PRIMARY KEY,
		url TEXT,
		visited BOOLEAN
	);`

	tx, err := db.Begin()
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			panic("could not rollback DB transaction!")
		}
		return fmt.Errorf("error %v while beginning DB transaction", err)
	}

	_, err = tx.Exec(tableSchema)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			panic("could not rollback DB transaction!")
		}
		return fmt.Errorf("error %v crating DB schema %s", err, tableSchema)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error %v while commiting DB transaction", err)
	}

	log.Infof("Initializing DB")
	return nil
}
