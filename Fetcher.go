// Fetcher is an object which can be used to fetch the contents of a webpage.
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// DefaultFetcherClient is the http Client that newly created fetchers use
var DefaultFetcherClient = &http.Client{Timeout: time.Duration(30) * time.Second}

// Fetcher fetches webpages.
type Fetcher struct {
	client *http.Client
}

// NewFetcher initializes and returns a Fetcher. If a value for fClient is provided, the
// Fetcher will use that as a client. Otherwise, it will use the DefaultFetcherClient.
func NewFetcher(fClient *http.Client) *Fetcher {
	if fClient == nil {
		fClient = DefaultFetcherClient
	}

	return &Fetcher{
		client: fClient,
	}
}

// FetchWebpage takes the URL to a web page, and returns a pointer to the body of the given page,
// and an error if it could not.
func (f *Fetcher) FetchWebpage(u string) (*io.ReadCloser, error) {
	// Verify the url is an actual URL
	if !isValidURL(u) {
		return nil, fmt.Errorf("url %s is not valid", u)
	}

	// GET the page, and return an error if there was one
	resp, err := f.client.Get(u)
	if err != nil {
		return nil, fmt.Errorf("received error %v getting url %s", err, u)
	}

	// Check that the status is in the 200s
	if resp.StatusCode < 200 && resp.StatusCode >= 300 {
		return nil, fmt.Errorf("received status code %d getting url %s", resp.StatusCode, u)
	}

	// Ensurre content type is text
	ct := resp.Header.Get("content-type")
	validCT := isValidContentType(ct)
	if !validCT {
		return nil, fmt.Errorf("content-type %s of %s is not valid", ct, u)
	}

	return &resp.Body, nil
}

// PageResults represents what was scraped from a web page
type PageResults struct {
	URL       string
	ValidURLs []string
}

// ScrapePage Fetches a url and then scrapes the contents of the page into a PageResults struct.
// It also returns an optional error if there was an issue.
func (f *Fetcher) ScrapePage(u string) (*PageResults, error) {
	// Fetch the webpage first.
	body, err := f.FetchWebpage(u)
	if err != nil {
		return nil, fmt.Errorf("error %v while fetching webpage %s", err, u)
	}

	// Get all the hrefs from the body
	hrefs, err := getBodyhrefs(body)
	if err != nil {
		return nil, fmt.Errorf("error %v getting hrefs from body", err)
	}

	// Make a slice for valid URLs, and append any valid hrefs to it
	validURLs := []string{}
	for _, testURL := range hrefs {
		if isValidURL(testURL) {
			validURLs = append(validURLs, testURL)
		}
	}

	// Return the PageResultd
	return &PageResults{
		URL:       u,
		ValidURLs: validURLs,
	}, nil
}

// getBodyhrefs parses the body of a web page and returns all the links found on the page, and an error if it could not.
// It always closes the body it is provided, whether it could be read or not.
func getBodyhrefs(body *io.ReadCloser) ([]string, error) {
	defer (*body).Close()

	// Attempt to turn the body into an HTML Node
	rootNode, err := html.Parse(*body)
	if err != nil {
		return nil, fmt.Errorf("received error %v parsing content of body", err)
	}

	// Make a slice for the hrefs
	hrefs := []string{}

	// Create and launch a recursive function to parse the node tree.
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
	getNodeHrefs(rootNode)

	return hrefs, nil
}

// isValidURL takes a URL as a string, and returns a bool of whether it is valid.
func isValidURL(u string) bool {
	p, err := url.Parse(u)
	if err != nil {
		return false
	}
	if !(p.Scheme == "http" ||
		p.Scheme == "https" ||
		p.Scheme == "HTTP" ||
		p.Scheme == "HTTPS") ||
		p.Host == "" ||
		p.Path == "" {
		return false
	}

	return true
}

// isValidContentType takes a content-type header as a string, and returns a bool of whether it is valid.
func isValidContentType(ct string) bool {
	return strings.Contains(ct, "text/html") ||
		strings.Contains(ct, "text/plain")
}
