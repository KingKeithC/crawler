package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/kingkeithc/crawler/dal"

	flags "github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
)

// Args are the command line arguments passed to the program.
var Args struct {
	DBHost   string   `long:"dbHost" env:"CRAWLER_DB_HOST" default:"localhost" description:"The hostname of the DB server."`
	DBUser   string   `long:"dbUser" env:"CRAWLER_DB_USER" default:"crawler" description:"The username to connect to the DB with."`
	DBPass   string   `long:"dbPass" env:"CRAWLER_DB_PASS" default:"" description:"The password to connect to the DB with."`
	DBName   string   `long:"dbName" env:"CRAWLER_DB_NAME" default:"crawler" description:"The DB name to connect to."`
	SeedURLs []string `long:"seedURLs" short:"u" description:"The initial seed URLs to crawl." required:"true"`
	LogLevel string   `long:"loglevel" env:"CRAWLER_LOGLEVEL" short:"l" description:"The log level of the program." choice:"FATAL" choice:"ERROR" choice:"WARN" choice:"INFO" choice:"DEBUG" default:"INFO"`
}

// log is the logger
var log = logrus.New()

func init() {
	log.Infoln("Initializing the program.")

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

	log.Debugf("The arguments are: %+v", Args)
	log.Infoln("Main Initialized.")
}

func main() {
	// Initialize the Data Access Layer with the connection string and logger
	constr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		Args.DBHost, Args.DBUser, Args.DBPass, Args.DBName)
	dal.Init(constr, log)
	defer dal.DB.Close()

	// Open the null file for writing
	null, err := os.OpenFile(os.DevNull, os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Fatalf("could not open null for for writing")
	}
	defer null.Close()

	f := NewFetcher(nil)
	for _, seedURL := range Args.SeedURLs {
		body, err := f.FetchWebpage(seedURL)
		if err != nil {
			log.Warnf("%v", err)
			continue
		}
		defer (*body).Close()

		// Create a reader for the body
		read := bufio.NewReader(*body)

		// Write the body to null
		written, err := read.WriteTo(null)
		if err != nil {
			log.Fatalf("could not write to null file")
		}

		// Log the amount of bytes written
		log.Infof("Wrote %d bytes to null from url %s", written, seedURL)
	}
}
