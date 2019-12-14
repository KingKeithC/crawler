// Crawler.go defines methods related to the crawler object.
package main

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/lib/pq"
)

const (
	// StateReady is when the crawler is created, but hasnt begun running yet.
	StateReady = iota
	// StateRunning is when the crawler is running.
	StateRunning
	// StateStopped is when the crawler has finished.
	StateStopped
)

// Crawler given a set of seed pages crawls those pages searching for links.
type Crawler struct {
	db        *sql.DB
	workers   int
	unvisited chan string
	visited   chan *Scraping
	stop      *sync.Once
	state     int
	delay     int
}

// NewCrawler iniitializes and returns a crawler.
func NewCrawler(db *sql.DB, workers int, delay int) (c *Crawler) {
	c = &Crawler{
		db:        db,
		workers:   workers,
		unvisited: make(chan string, 1000),
		visited:   make(chan *Scraping, 100),
		stop:      &sync.Once{},
		state:     StateReady,
		delay:     delay,
	}
	return
}

// State returns the state of the crawler.
func (c *Crawler) State() int {
	return c.state
}

// AddURLs adds the URLs to the unvisited channel
func (c *Crawler) AddURLs(urls ...string) error {
	if c.State() == StateStopped {
		return fmt.Errorf("crawler is already stopped")
	}
	go func() {
		for _, u := range urls {
			if isValidURL(u) {
				c.unvisited <- u
			}
		}
	}()
	return nil
}

// Run ...
func (c *Crawler) Run() {
	if c.State() != StateReady {
		return
	}

	// Set the new state
	c.state = StateRunning

	// Create a waitgroup
	wg := &sync.WaitGroup{}

	// Launch the workers
	for i := 0; i < c.workers; i++ {
		wg.Add(1)
		go c.worker(wg)
	}

	// Launch the queue monitor
	go c.monitor()

	// Wait for them to complete
	log.Infof("%d workers launched, now waiting while they crawl", c.workers)
	wg.Wait()

	c.flushQueuesToDB()

	c.state = StateStopped
	log.Infof("crawlers finished, goodbye :)")
}

func (c *Crawler) monitor() {
	log.Infof("monitor launched")
	for {
		// Wait 5 seconds
		time.Sleep(time.Duration(5) * time.Second)

		if c.State() == StateStopped {
			log.Infof("moitor returning due to StateStopped")
			return
		}

		c.flushQueuesToDB()
	}
}

func (c *Crawler) flushQueuesToDB() {
	unvisitedSz := len(c.unvisited)
	visitedSz := len(c.visited)

	if unvisitedSz < 100 || visitedSz < 50 {
		return
	}

	log.Warnf("inserting %d: unvisited, %d: visited URLs to DB", unvisitedSz, visitedSz)

	// Create a transaction
	tx, err := c.db.Begin()
	if err != nil {
		abortTx(tx)
		log.Fatalf("error %v while beginning transaction", err)
	}

	// Prepare a statement for the transaction
	stmt, err := tx.Prepare(pq.CopyIn("urls", "url", "visited"))
	if err != nil {
		abortTx(tx)
		log.Fatalf("error %v while preparing statement", err)
	}

	// Insert the unvisited URLs into the DB
	for i := 0; i < unvisitedSz; i++ {
		u := <-c.unvisited
		_, err = stmt.Exec(u, false)
		if err != nil {
			abortTx(tx)
			log.Fatalf("error %v while executing insert statement", err)
		}
	}

	// Insert the visited URLs into the DB
	for i := 0; i < visitedSz; i++ {
		s := <-c.visited
		_, err = stmt.Exec(s.URL, true)
		if err != nil {
			abortTx(tx)
			log.Fatalf("error %v while executing insert statement", err)
		}
	}

	// Add a final execute to the statement
	_, err = stmt.Exec()
	if err != nil {
		abortTx(tx)
		log.Fatalf("error %v while executing final statement", err)
	}

	// Close the statement
	err = stmt.Close()
	if err != nil {
		abortTx(tx)
		log.Fatalf("error %v while closing statement", err)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		abortTx(tx)
		log.Fatalf("error %v while committing transaction", err)
	}
}

func (c *Crawler) worker(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		// Sleep for a bit so we're not DOSing
		if c.delay > 0 {
			time.Sleep(time.Duration(c.delay) * time.Millisecond)
		}

		if c.State() == StateStopped {
			log.Infof("worker returning due to StateStopped")
			return
		}
		u, ok := <-c.unvisited
		if !ok {
			log.Infof("worker returning due to not ok receive")
			return
		}

		s, err := Scrape(u)
		if err != nil {
			log.Infof("URL: %s, Returned: %v", u, err)
		} else {
			// Put the results into the channels
			c.visited <- s
			c.AddURLs(s.ValidHrefs...)
			log.Infof("URL: %s, Returned: %d/%d Valid Hrefs", u, len(s.ValidHrefs), len(s.RawHrefs))
		}
	}
}

// Stop stops a running crawler. It has no affect on a stopped crawler.
func (c *Crawler) Stop() {
	c.stop.Do(func() {
		c.state = StateStopped
	})
}
