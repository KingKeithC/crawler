// Crawler.go defines methods related to the crawler object.
package main

import (
	"sync"

	"github.com/sirupsen/logrus"
)

// Crawler given a set of seed pages crawls those pages searching for links.
type Crawler struct {
	ID             int
	log            *logrus.Logger
	numFetchers    int
	unvisitedURLs  chan string
	visitedResults chan *PageResults
	running        bool
}

// NewCrawler iniitializes and returns a crawler.
func NewCrawler(cID int, cLog *logrus.Logger) *Crawler {
	if cLog == nil {
		cLog = logrus.New()
	}

	// Create the channels
	unvisited := make(chan string, 100)
	visitedResults := make(chan *PageResults, 100)

	c := &Crawler{
		ID:             cID,
		log:            cLog,
		numFetchers:    10,
		unvisitedURLs:  unvisited,
		visitedResults: visitedResults,
		running:        false,
	}

	c.log.Debugf("Created Crawler %+v", c)
	return c
}

// AddURLs adds a slice of URLs to the unvisited slice
func (c *Crawler) AddURLs(URLsToAdd ...string) {
	for _, URLToAdd := range URLsToAdd {
		c.log.Debugf("Adding URL %s to Crawler %d", URLToAdd, c.ID)
		c.unvisitedURLs <- URLToAdd
	}
}

// Stop stops a running crawler. It has no affect on a stopped crawler.
func (c *Crawler) Stop() {
	c.running = false
}

// CrawlForever crawls the channel of unscraped URLs, and a channel of page results. It retrieves a URL from the
// unscraped URLs channel and scrapes the URL. If the scraped URL returns an error, the error is logged, otherwise the
// returned PageResults are sent to the visitedResults channel.
//
func (c *Crawler) CrawlForever() {
	c.log.Debugf("Crawler %d: CrawlingForever...", c.ID)
	// If we are already running, just return
	if c.running {
		c.log.Warnf("Crawler %d: CrawlForwver called, but we're already running. Returning early...", c.ID)
		return
	}

	// Set running to true
	c.running = true

	// Make a wait group so we know when these guys are finshed
	wg := &sync.WaitGroup{}

	// Launch c.numFetchers goroutines to fetch a url from the channel, and place the results in the c.PageResults channel if there
	// are any. If there was an error, it is logged.
	for i := 0; i < c.numFetchers; i++ {
		f := NewFetcher(nil)
		go func() {
			// Add ourselves to the waitgroup, and defer its closure
			wg.Add(1)
			defer wg.Done()

			// Loop through the unvisitedURLs channel
			for u := range c.unvisitedURLs {
				if !c.running {
					c.log.Warnf("Crawler %d: We should no longer be running. Leaving...", c.ID)
					return
				}

				c.log.Infof("Crawler %d: Scraping Page %s", c.ID, u)
				res, err := f.ScrapePage(u)
				if err != nil {
					c.log.Warnf("Crawler %d: URL %s Received Error %v", c.ID, u, err)
					// If it was an error, just reiterate the loop
					continue
				}
				err = nil

				// Just in case
				if res == nil {
					c.log.Errorf("Crawler %d: URL %s received a nil PageResults???", c.ID, u)
				}

				// Add the results
				c.log.Infof("Crawler %d: Found %d URLs on %s", c.ID, len(res.ValidURLs), res.URL)
				c.visitedResults <- res
			}
		}()

		// Wait for the fetchers to finish
		c.log.Infof("Crawler %d: %d Fetchers launched. Now waiting for them to exit.", c.ID, c.numFetchers)
		wg.Wait()
		c.log.Infof("Crawler %d: All Fetchers finished. Continuing...", c.ID)

		c.running = false
	}
}
