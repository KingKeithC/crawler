// Crawler.go defines methods related to the crawler object.
package main

import (
	"github.com/sirupsen/logrus"
)

// Crawler given a set of seed pages crawls those pages searching for links.
type Crawler struct {
	ID            int
	logger        *logrus.Logger
	numFetchers   int
	fetchers      []*Fetcher
	unvisitedURLs *chan string
	visitedURLs   *chan string
	running       bool
}

// NewCrawler iniitializes and returns a crawler.
func NewCrawler(cID int, cLog *logrus.Logger) *Crawler {
	if cLog == nil {
		cLog = logrus.New()
	}

	// Create the channels
	unvisited := make(chan string, 100)
	visited := make(chan string, 100)

	c := &Crawler{
		ID:            cID,
		logger:        cLog,
		numFetchers:   10,
		fetchers:      []*Fetcher{},
		unvisitedURLs: &unvisited,
		visitedURLs:   &visited,
		running:       false,
	}

	// Create the fetchers and add them to the crawler
	for i := 0; i < c.numFetchers; i++ {
		c.fetchers[i] = NewFetcher(nil)
	}

	log.Debugf("Created Crawler %+v", c)
	return c
}

// AddURLs adds a slice of URLs to the unvisited slice
func (c *Crawler) AddURLs(URLsToAdd ...string) {
	for _, URLToAdd := range URLsToAdd {
		*c.unvisitedURLs <- URLToAdd
	}
}
