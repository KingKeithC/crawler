// Crawler.go defines methods related to the crawler object.
package main

import (
	"github.com/sirupsen/logrus"
)

// Crawler given a set of seed pages crawls those pages searching for links.
type Crawler struct {
	ID            int
	log           *logrus.Logger
	numFetchers   int
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
		log:           cLog,
		numFetchers:   10,
		unvisitedURLs: &unvisited,
		visitedURLs:   &visited,
		running:       false,
	}

	c.log.Debugf("Created Crawler %+v", c)
	return c
}

// AddURLs adds a slice of URLs to the unvisited slice
func (c *Crawler) AddURLs(URLsToAdd ...string) {
	for _, URLToAdd := range URLsToAdd {
		c.log.Debugf("Adding URL %s to Crawler %d", URLToAdd, c.ID)
		*c.unvisitedURLs <- URLToAdd
	}
}

// Crawl is the main method of the Crawler.
func (c *Crawler) Crawl() {
	c.log.Infof("Crawler %d is now Crawling...", c.ID)

}
