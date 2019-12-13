package main

import (
	"database/sql"
	"fmt"

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

	// Start a transaction to keep the modifications
	tx, err := db.Begin()
	if err != nil {
		abortTx(tx)
		return nil, fmt.Errorf("error %v while beginning DB transaction", err)
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
		abortTx(tx)
		return nil, fmt.Errorf("error %v crating DB schema %s", err, tableSchema)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		abortTx(tx)
		return nil, fmt.Errorf("error %v while commiting DB transaction", err)
	}

	log.Infof("Database Initialized.")
	return db, nil
}

// abortTx aborts the transaction, and kills the program
func abortTx(tx *sql.Tx) {
	if rollbackErr := tx.Rollback(); rollbackErr != nil {
		log.Fatalf("failed to rollback transaction %v", tx)
	}
}
