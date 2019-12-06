package main

import (
	"database/sql"
	"fmt"

	// Necessary for SQL
	_ "github.com/lib/pq"
)

// InitDB creates and tests a connection to the SQL DB
func InitDB(conStr string) (*sql.DB, error) {
	log.Infof("Initializing Database...")

	// Create the DB Object
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

	// Create the Schema
	err = CreateSchema(db)
	if err != nil {
		return nil, fmt.Errorf("error %v creating schema", err)
	}

	log.Infof("Database Initialized.")
	return db, nil
}

// CreateSchema connects to the DB and ensures that the schema exists
func CreateSchema(db *sql.DB) error {
	// Start a transaction to keep the modifications
	tx, err := db.Begin()
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			panic("could not rollback DB transaction!")
		}
		return fmt.Errorf("error %v while beginning DB transaction", err)
	}

	// Create the table schema
	tableSchema := `
	CREATE TABLE IF NOT EXISTS urls (
		id serial PRIMARY KEY,
		url TEXT,
		visited BOOLEAN
	);`
	_, err = tx.Exec(tableSchema)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			panic("could not rollback DB transaction!")
		}
		return fmt.Errorf("error %v crating DB schema %s", err, tableSchema)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error %v while commiting DB transaction", err)
	}

	log.Infof("Schema Initialized.")
	return nil
}
