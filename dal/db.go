package dal

import (
	"database/sql"

	"github.com/sirupsen/logrus"

	// Necessary for sql
	_ "github.com/lib/pq"
)

var (
	// DB is the actual Database object
	DB         *sql.DB
	initalized = false
	log        *logrus.Logger
)

// Init creates and tests a connection to the SQL DB
func Init(conStr string, tlog *logrus.Logger) {
	if tlog == nil {
		panic("dal log can't be null!")
	}
	log = tlog

	db, err := sql.Open("postgres", conStr)
	if err != nil {
		log.Fatalf("error %v opening connection to %v", err, conStr)
	}
	DB = db

	// Test that the DB connection works
	err = db.Ping()
	if err != nil {
		log.Fatalf("ping to DB failed with error %v", err)
	}

	log.Printf("DAL Initialized.")
}
