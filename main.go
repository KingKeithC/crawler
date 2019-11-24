package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
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
	if err != nil {
		panic(err)
	}
}

func main() {
	log := logrus.New()
	log.Infof("Running with these options %+v", opts)

	originalUrL := opts.Positional.PagesToVisit[0]
	iterations := uint32(opts.Iterations)

	crawler := NewCrawler(1, log, nil)
	visited, unvisited := crawler.CrawlNRecursively(originalUrL, iterations)

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
	fmt.Printf("The Original Url was:\n")
	fmt.Printf("%s\n\n", originalUrL)
	fmt.Printf("After %d iterations:\n", iterations)
	fmt.Printf("%8d: URLs were Visited, and\n", len(visited))
	fmt.Printf("%8d: URLs are Unvisited.\n\n", len(unvisited))
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
			log.Fatalf("Error writing string %s to buffer %v", str, *buf)
		}
	}

	// Return the contents of the buffer as bytes
	return buf.Bytes()
}
