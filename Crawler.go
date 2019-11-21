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
		cLog = logrus.New()
	}

	// Create a default client if none were passed
	if cHTTPClient == nil {
		cHTTPClient = &http.Client{}
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
func (c *Crawler) CrawlWebpage(url string) ([]string, error) {
	c.logger.Infof("Crawler %d: Getting Page %v.", c.ID, url)

	// 1. GET the web page and if we cannot connect return an error.
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return []string{}, fmt.Errorf("error getting webpage %v, due to %w", url, err)
	}
	defer resp.Body.Close()

	// 2. If the status is not 2XX return an error.
	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return []string{}, fmt.Errorf("status Code Not Acceptable. expecting 2XX, got %d", resp.StatusCode)
	}

	// 2.5 Error if headers do not specify type as html
	if !strings.Contains(resp.Header.Get("content-type"), "text/html") {
		return []string{}, fmt.Errorf("content-type of %v did not contain text/html", url)
	}

	// 3. Attempt to parse the HTML, return an error if there are any issues.
	document, err := html.Parse(resp.Body)
	if err != nil {
		return []string{}, fmt.Errorf("could not parse HTML body from %v, got error %w", url, err)
	}

	// 4. Gather all the hrefs add them to the PagesFound for this Crawler,
	//    and finally return them.
	foundUrls := parseNode(document)
	c.PagesVisited[url] = foundUrls
	return foundUrls, nil
}

func main() {
	// Create and prime a crawler
	crawler := NewCrawler(0, nil, nil)
	links, err := crawler.CrawlWebpage("https://www.yourhtmlsource.com/myfirstsite/")
	if err != nil {
		fmt.Printf("Error Crawling webpage! %+v.", err)
		os.Exit(1)
	}
	fmt.Printf("Found %v links on %v pages.", len(links), len(crawler.PagesVisited))
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

	fmt.Printf("Found %d hrefs in node.\n", len(hrefs))
	return hrefs
}
