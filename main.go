package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/jessevdk/go-flags"

	"github.com/sirupsen/logrus"
)

var opts struct {
	Iterations int `short:"i" long:"iterations" default:"250"`
	Positional struct {
		PagesToVisit []string
	} `positional-args:"yes" required:"yes"`
}

func init() {
	// Parse the command line arguments
	_, err := flags.Parse(&opts)

	// If there was an error, but no Positional Arguments were Provided, log a warning and continue.
	// Otherwise, log the error and exit.
	if err != nil {
		if len(opts.Positional.PagesToVisit) == 0 {
			logrus.Errorf("No Pages to visit were provided!")
		} else {
			logrus.Fatalf("Error %w was logged when parsong the arguments.", err)
		}
	}
}

func main() {
	log := logrus.New()
	log.Debugf("Running with these options %+v", opts)

	crawler := NewCrawler(1, log, nil)
	visited, unvisited := crawler.CrawlNRecursively(uint32(opts.Iterations), opts.Positional.PagesToVisit...)

	visitedBytes := strSliceToByteSlice(visited)
	unvisitedBytes := strSliceToByteSlice(unvisited)

	err := ioutil.WriteFile("visited.txt", visitedBytes, os.ModePerm)
	if err != nil {
		log.Fatalf("Error writing visited.txt")
	}

	err = ioutil.WriteFile("unvisited.txt", unvisitedBytes, os.ModePerm)
	if err != nil {
		log.Fatalf("Error writing unvisited.txt")
	}

	fmt.Printf("***********************************\n\n")
	fmt.Printf("The Original Urls were:\n")
	fmt.Printf("%v\n\n", opts.Positional.PagesToVisit)
	fmt.Printf("After %d iterations:\n", opts.Iterations)
	fmt.Printf("%8d URLs were Visited, and\n", len(visited))
	fmt.Printf("%8d URLs are Unvisited.\n\n", len(unvisited))
	fmt.Printf("***********************************\n")
}

func strSliceToByteSlice(strSlice []string) []byte {
	// Create a buffer to store our bytes in
	buf := &bytes.Buffer{}

	// For each string, add it to the buffer
	for i, str := range strSlice {
		// If its not the first string, append a newline to the beginning
		if i != 0 {
			str = fmt.Sprintf("\n%s", str)
		}

		// Write the string to the buffer, and check for errors
		_, err := buf.WriteString(str)
		if err != nil {
			panic("Error writing string to buffer!")
		}
	}

	// Return the contents of the buffer as bytes
	return buf.Bytes()
}
