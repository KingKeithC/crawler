// Crawler.go defines methods related to the crawler object.
package main

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Crawler given a set of seed pages crawls those pages searching for links.
type Crawler struct {
	ID             int
	log            *logrus.Logger
	db             *sql.DB
	numFetchers    int
	unvisitedURLs  chan string
	visitedResults chan *PageResults
	running        bool
}

// NewCrawler iniitializes and returns a crawler.
func NewCrawler(cID int, cLog *logrus.Logger, db *sql.DB) *Crawler {
	if cLog == nil {
		cLog = logrus.New()
	}

	// Create the channels
	unvisited := make(chan string, 500)
	visitedResults := make(chan *PageResults, 500)

	c := &Crawler{
		ID:             cID,
		log:            cLog,
		db:             db,
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

// AttemptScraping Makes an attempt to scrape
func (c *Crawler) AttemptScraping(i int, wg *sync.WaitGroup) {
	// Add ourselves to the waitgroup, and defer its closure
	wg.Add(1)
	defer wg.Done()

	f := NewFetcher(nil)

	// Loop through the unvisitedURLs channel
	for u := range c.unvisitedURLs {
		if !c.running {
			c.log.Warnf("Fetcher %d: We should no longer be running. Leaving...", i)
			return
		}

		c.log.Infof("Fetcher %d: Scraping Page %s", i, u)
		res, err := f.ScrapePage(u)
		if err != nil {
			c.log.Warnf("Fetcher %d: URL %s Received Error %v", i, u, err)
			continue
		}

		// Add the results
		c.log.Infof("Fetcher %d: Found %d URLs on %s", i, len(res.ValidURLs), res.URL)

		// Add the found URLs to the unvisited channel
		for _, URL := range res.ValidURLs {
			c.unvisitedURLs <- URL
		}

		// Add the results to the visitedresults channel
		c.visitedResults <- res
	}
}

// StoreQueued Inserts all urls from the visitedResults channel into the urls channel.
// It also Inserts all but numFetchers *2 URLs from the unvisitedURLs channel
func (c *Crawler) StoreQueued() error {
	numVisited := len(c.visitedResults)
	numUnvisited := len(c.unvisitedURLs)

	c.log.Infof("Storing Queued URLs. There are currently %d UnvisitedURLs, and %d VisitedURLs queued.",
		numUnvisited, numVisited)

	// Prepare the insert statement
	stmt, err := c.db.Prepare(`INSERT INTO "urls" (url, visited) VALUES ($1, $2)`)
	if err != nil {
		return fmt.Errorf("error %v while preparing statement", err)
	}
	defer stmt.Close()

	// Insert each visitedResult into the DB
	for i := 0; i < numVisited; i++ {
		v := <-c.visitedResults

		c.log.Debugf("Inserting Visited URL %s into DB.", v.URL)
		_, err := stmt.Exec(v.URL, true)
		if err != nil {
			return fmt.Errorf("error %v inserting into DB", err)
		}
	}

	// Insert each unvisitedURL into the DB
	for i := 0; i < numUnvisited; i++ {
		v := <-c.unvisitedURLs

		c.log.Debugf("Inserting Unvisited URL %s into DB.", v)
		_, err := stmt.Exec(v, false)
		if err != nil {
			return fmt.Errorf("error %v inserting into DB", err)
		}
	}

	return nil
}

// CrawlForever crawls the channel of unscraped URLs, and a channel of page results. It retrieves a URL from the
// unscraped URLs channel and scrapes the URL. If the scraped URL returns an error, the error is logged, otherwise the
// returned PageResults are sent to the visitedResults channel.
func (c *Crawler) CrawlForever() {
	c.log.Debugf("Crawler %d: CrawlingForever...", c.ID)
	// If we are already running, just return
	if c.running {
		c.log.Warnf("Crawler %d: CrawlForever called, but we're already running. Returning early...", c.ID)
		return
	}

	// Set running to true
	c.running = true

	// Make a wait group so we know when these guys are finshed
	wg := &sync.WaitGroup{}

	// Launch c.numFetchers goroutines to fetch a url from the channel, and place the results in the c.PageResults channel if there
	// are any. If there was an error, it is logged.
	for i := 0; i < c.numFetchers; i++ {
		go c.AttemptScraping(i, wg)
	}

	// Create a function to call StoreQueued every 30 seconds.
	timer := time.AfterFunc(time.Duration(30)*time.Second, func() {
		err := c.StoreQueued()
		if err != nil {
			c.log.Fatalf("error %v while storing queued", err)
		}
	})
	defer timer.Stop()

	// Wait for the fetchers to finish
	c.log.Infof("Crawler %d: %d Fetchers launched. Now waiting for them to exit.", c.ID, c.numFetchers)
	wg.Wait()
	c.log.Infof("Crawler %d: All Fetchers finished. Continuing...", c.ID)
}
