// Crawler.go defines methods related to the crawler object.
package main

import (
	"github.com/sirupsen/logrus"
)

// Crawler given a set of seed pages crawls those pages searching for links.
type Crawler struct {
	ID          int
	logger      *logrus.Logger
	numFetchers int
	fetchers    []*Fetcher
}

// NewCrawler iniitializes and returns a crawler.
func NewCrawler(cID int, cLog *logrus.Logger) *Crawler {
	if cLog == nil {
		cLog = logrus.New()
	}

	c := &Crawler{
		ID:          cID,
		logger:      cLog,
		numFetchers: 10,
		fetchers:    []*Fetcher{},
	}

	// Create the fetchers
	for i := 0; i < c.numFetchers; i++ {
		c.fetchers[i] = NewFetcher()
	}

	return c
}
