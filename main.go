package main

import (
	"fmt"
	"os"

	flags "github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"

	// Necessary for SQL
	_ "github.com/lib/pq"
)

// Args are the command line arguments passed to the program.
var Args struct {
	DBHost     string   `long:"db-host" env:"CRAWLER_DB_HOST" default:"localhost" description:"The hostname of the DB server."`
	DBPort     string   `long:"db-port" env:"CRAWLER_DB_PORT" default:"5432" description:"The port of the DB server."`
	DBUser     string   `long:"db-bser" env:"CRAWLER_DB_USER" default:"crawler" description:"The username to connect to the DB with."`
	DBPass     string   `long:"db-pass" env:"CRAWLER_DB_PASS" default:"" description:"The password to connect to the DB with."`
	DBName     string   `long:"db-name" env:"CRAWLER_DB_NAME" default:"crawler" description:"The DB name to connect to."`
	SeedURLs   []string `long:"seed-urls" short:"u" description:"The initial seed URLs to crawl." required:"true"`
	LogLevel   string   `long:"log-level" env:"CRAWLER_LOGLEVEL" short:"l" description:"The log level of the program." choice:"FATAL" choice:"ERROR" choice:"WARN" choice:"INFO" choice:"DEBUG" default:"INFO"`
	DelayMills int      `long:"delay-mills" env:"CRAWLER_DELAY_MILLS" description:"The milliseconds between scrapes."`
}

// log is the logger
var log = logrus.New()

func main() {
	// Parse the arguments
	if _, err := flags.Parse(&Args); err != nil {
		os.Exit(1)
	}

	// Set the log level
	switch Args.LogLevel {
	case "FATAL":
		log.SetLevel(logrus.FatalLevel)
	case "ERROR":
		log.SetLevel(logrus.ErrorLevel)
	case "WARN":
		log.SetLevel(logrus.WarnLevel)
	case "INFO":
		log.SetLevel(logrus.InfoLevel)
	case "DEBUG":
		log.SetLevel(logrus.DebugLevel)
	default:
		log.Fatalf("log level %s is not supported", Args.LogLevel)
	}
	fmt.Printf("The log level is: %s", log.GetLevel().String())
	log.Debugf("The arguments are: %+v", Args)

	// Initialize the DB
	db, err := InitDB(fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		Args.DBHost, Args.DBPort, Args.DBUser, Args.DBPass, Args.DBName))
	if err != nil {
		log.Fatalf("could not initialize DB due to %v", err)
	}
	defer db.Close()

	// Run a Crawler
	c := NewCrawler(db, 10, Args.DelayMills)
	c.AddURLs(Args.SeedURLs...)
	c.Run()
}
