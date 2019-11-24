// Crawler.go defines methods related to the crawler object.
package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

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
		cLog = logrus.New()
	}

	// Create a default client with a 10 second timeout if none were provided
	if cHTTPClient == nil {
		cHTTPClient = &http.Client{
			Timeout: time.Second * 10,
		}
	}

	// Return the final crawler
	cLog.Infof("Created Crawler %d.", cID)
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
func (c *Crawler) CrawlWebpage(urlToVisit string) ([]string, error) {
	c.logger.Debugf("Crawler %d: Getting Page %v.", c.ID, urlToVisit)

	// 1. GET the web page and if we cannot connect return an error.
	resp, err := c.httpClient.Get(urlToVisit)
	if err != nil {
		return []string{}, fmt.Errorf("error getting webpage %v, due to %w", urlToVisit, err)
	}
	defer resp.Body.Close()

	// 2. If the status is not 2XX return an error.
	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return []string{}, fmt.Errorf("status Code Not Acceptable. expecting 2XX, got %d", resp.StatusCode)
	}

	// 2.5 Error if headers do not specify type as html
	if !strings.Contains(resp.Header.Get("content-type"), "text/html") {
		return []string{}, fmt.Errorf("content-type of %v did not contain text/html", urlToVisit)
	}

	// 3. Attempt to parse the HTML, return an error if there are any issues.
	document, err := html.Parse(resp.Body)
	if err != nil {
		return []string{}, fmt.Errorf("could not parse HTML body from %v, got error %w", urlToVisit, err)
	}

	// 4. Gather all the hrefs add them to the PagesFound for this Crawler,
	//    and finally return them.
	foundURLs := parseNode(document)

	// Validate the URLs that were found, and add only the valid ones to a slice
	validURLs := make([]string, 0, len(foundURLs))
	for _, foundURL := range foundURLs {
		if validURL, ok := ValidateURL(foundURL); ok {
			// Append the URL, since there was no error
			validURLs = append(validURLs, validURL)
		}
	}

	c.logger.Infof("Crawler %d: Found %d/%d valid hrefs on %s.", c.ID, len(validURLs), len(foundURLs), urlToVisit)
	c.PagesVisited[urlToVisit] = validURLs
	return validURLs, nil
}

// ValidateURL tests a url against a set of rules, if the URL is found to be
// valid, the string of the URL and true are returned, otherwise an empty string and false are returned.
func ValidateURL(u string) (string, bool) {
	uParsed, err := url.Parse(u)
	if err != nil {
		return "", false
	}

	if !uParsed.IsAbs() {
		return "", false
	}

	if uParsed.Scheme != "http" && uParsed.Scheme != "https" {
		return "", false
	}

	return uParsed.String(), true
}

// CrawlWebpages is a variadic function which calls CrawlWebpage for each URL passed as a parameter.
// The function sums up the found hrefs and returns them.
func (c *Crawler) CrawlWebpages(urls ...string) []string {
	c.logger.Infof("Crawler %d: Crawling %d Urls.", c.ID, len(urls))
	totalFoundUrls := []string{}
	for _, url := range urls {
		foundUrls, err := c.CrawlWebpage(url)
		if err != nil {
			c.logger.Warnf("Crawler %d: Caught Error %w for page %s, ignoring...", c.ID, err, url)
		}
		totalFoundUrls = append(totalFoundUrls, foundUrls...)
	}
	return totalFoundUrls
}

// CrawlNRecursively crawls a slice of URLs and then crawls all hrefs it found.
// It repeats this pattern N times and returns a slice of hrefs visited and unvisited.
func (c *Crawler) CrawlNRecursively(n uint32, URLsToVisit ...string) ([]string, []string) {
	c.logger.Infof("Crawler %d: Crawling %d URLs recursively for %d iterations.", c.ID, len(URLsToVisit), n)

	// Make our return variables as 25 times the iterations, as on average we get about 25
	// hrefs when crawling any average page. **As if this writing I have no real evidence to back
	// this up, however I intend to benchmark this and confirm the real number.
	var visitedUrls, unvisitedUrls []string
	visitedUrls = make([]string, 0, 25*n)
	unvisitedUrls = make([]string, 0, 25*n)

	// Add the first URLs to the slice of unvisitedUrls
	unvisitedUrls = append(unvisitedUrls, URLsToVisit...)

	// For n iterations
	for i := uint32(0); i < n; i++ {
		// Take the top URL to visit off the stack then update the stack to be everything else
		nextURL := unvisitedUrls[i]
		unvisitedUrls = unvisitedUrls[1:]

		// Crawl the page and log but ignore any errors.
		found, err := c.CrawlWebpage(nextURL)
		if err != nil {
			c.logger.Warnf("Crawler %d: URL %s returned error %w. Ignoring...", c.ID, nextURL, err)
		}

		// Update the stack of visited URLs with this URL,
		// and the stack of unvisited URLs with the URLs found.
		visitedUrls = append(visitedUrls, nextURL)
		unvisitedUrls = append(unvisitedUrls, found...)

		// Return early if there are no URLs left to crawl.
		if len(unvisitedUrls) == 0 {
			c.logger.Infof("Crawler %d: No URLs left to visit. Exiting...", c.ID)
			return visitedUrls, unvisitedUrls
		}
	}

	// All iterations done, return
	c.logger.Infof("Crawler %d: All %d iterations complete. Exiting....", c.ID, n)
	return visitedUrls, unvisitedUrls
}

// parseNode Takes an HTML Node and returns a list
// of all hrefs found in a tags under that node
func parseNode(firstNode *html.Node) []string {
	hrefs := []string{}

	// If a nil node was passed, just return nothing
	if firstNode == nil {
		return hrefs
	}

	// Create an anonymous function to recursively parse each node
	var getNodeHrefs func(*html.Node)
	getNodeHrefs = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "a" {
			// Loop through the elements attributes, if one of them
			// if an href, append the value of the href to the hrefs slice
			// and break out of the for loop
			for _, attr := range node.Attr {
				if attr.Key == "href" {
					hrefs = append(hrefs, attr.Val)
					break
				}
			}
		}

		// Perform Depth-First recursive search, calling this function for each element found.
		// Calls this function on the first child of this node, then when it eventually returns,
		// it calls the function on the next sibling of the first child, then on each sibling thereafter.
		for next := node.FirstChild; next != nil; next = next.NextSibling {
			getNodeHrefs(next)
		}
	}

	// Begin by parsing this node
	getNodeHrefs(firstNode)
	return hrefs
}
