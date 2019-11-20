// Crawler.go defines methods related to the crawler object.
package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"

	"github.com/sirupsen/logrus"
)

// Crawler is a web crawler object which when given a starting page.
// ID is a unique int representing the Crawler, for logging purposes.
// PagesVisited is a map of page urls to a slice of links found on the page.
type Crawler struct {
	ID           int
	PagesVisited map[string][]string
	logger       *logrus.Logger
	httpClient   *http.Client
}

// NewCrawler initializes and returns a new Crawler.
// Pass in an optional http.Client to override the client the
// crawler will use for its http requests.
func NewCrawler(cID int, cLog *logrus.Logger, cHTTPClient *http.Client) *Crawler {
	// Create a default logger if one was not provided
	if cLog == nil {
		cLog := logrus.New()
	}

	// Create a default client if none were passed
	if cHTTPClient == nil {
		cHTTPClient := &http.Client{}
	}

	// Return the final crawler
	cLog.Infof("Created Logger $%d.\n", cID)
	return &Crawler{
		ID:           cID,
		PagesVisited: map[string][]string{},
		logger:       cLog,
		httpClient:   cHTTPClient,
	}
}

// CrawlWebpage gets the content of a web page and then returns a slice of,
// urls it found on the page. It optionally returns an error if something
// unexpected happens.
func (c *Crawler) CrawlWebpage(url string) ([]string, error) {
	c.logger.Infof("Getting Page %v.\n", url)

	// 1. GET the web page and if we cannot connect return an error.
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Error getting webpage %s, due to %w.\n", url, err)
	}

	// 2. If the status is not 2XX return an error.
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return []string{}, fmt.Errorf("Status Code Not Acceptable. Expecting 2XX, got %d.\n", resp.StatusCode)
	}

	// 2.5 Error if headers do not specify type as html
	if !strings.Contains(resp.Header.Get("content-type"), "text/html") {
		return []string{}, fmt.Errorf("content-type of %v did not contain text/html.\n", url)
	}

	// 3. Attempt to parse the HTML, return an error if there are any issues.
	tokenizer := html.NewTokenizer(resp.Body)

	// TODO: Write tokenizer.

	foundUrls = []string{}
	return nil, foundUrls
}

func main() {
	// Create and prime a crawler
	crawler := NewCrawler(0, nil, nil)
	links, err := crawler.CrawlWebpage("https://en.wikipedia.org/wiki/Main_Page")
	if err != nil {
		fmt.Printf("Error Crawling webpage! %+v.\n", err)
		os.Exit(1)
	}
	fmt.Printf("Found %v links on %v pages.\n", len(links), len(crawler.PagesVisited))
}
