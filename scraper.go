package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

var h = &http.Client{Timeout: time.Duration(10) * time.Second}

// Scraping is what was found after scraping a URL.
type Scraping struct {
	URL        string
	RawHrefs   []string
	ValidHrefs []string
}

// Scrape takes a URL and makes an HTTP GET request for the whatever is at the URL.
// it parses the results into a *Scraping struct which represents what was found on the page.
func Scrape(u string) (s *Scraping, err error) {
	// Validate the URL
	if !isValidURL(u) {
		err = fmt.Errorf("url %s is not valid", u)
		return
	}

	// GET the page, and return an error if there was one
	resp, err := h.Get(u)
	if err != nil {
		err = fmt.Errorf("received error %v getting url %s", err, u)
		return
	}
	defer resp.Body.Close()

	// Check that the status is in the 200s
	if resp.StatusCode < 200 && resp.StatusCode >= 300 {
		err = fmt.Errorf("received status code %d getting url %s", resp.StatusCode, u)
		return
	}

	// Ensure content type is text
	ct := resp.Header.Get("content-type")
	if !isValidContentType(ct) {
		err = fmt.Errorf("content-type %s of %s is not valid", ct, u)
		return
	}

	// Attempt to turn the body into an HTML Node
	rootNode, err := html.Parse(resp.Body)
	if err != nil {
		err = fmt.Errorf("received error %v parsing content of body", err)
		return
	}

	// Get all the hrefs from the body
	hrefs := getBodyhrefs(rootNode)
	if err != nil {
		err = fmt.Errorf("error %v getting hrefs from body", err)
		return
	}

	// Make a separate slice of valid hrefs
	validHrefs := []string{}
	for _, href := range hrefs {
		if isValidURL(href) {
			validHrefs = append(validHrefs, href)
		}
	}

	s = &Scraping{
		URL:        u,
		RawHrefs:   hrefs,
		ValidHrefs: validHrefs,
	}
	return
}

// getBodyhrefs ...
func getBodyhrefs(root *html.Node) (hrefs []string) {
	// Create and launch a recursive function to parse the node tree.
	// Create an anonymous function to recursively parse each node
	var getNodeHrefs func(node *html.Node)

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

		// Perform Depth-First recursive search, calling this function for each
		// element found. Calls this function on the first child of this node,
		// then when it eventually returns, it calls the function on the next
		// sibling of the first child, then on each sibling thereafter.
		for next := node.FirstChild; next != nil; next = next.NextSibling {
			getNodeHrefs(next)
		}
	}

	// Begin by parsing this node
	getNodeHrefs(root)
	return
}

// isValidURL takes a URL as a string, and returns a bool of whether it is valid.
func isValidURL(u string) bool {
	p, err := url.ParseRequestURI(u)
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

// isValidContentType takes a content-type header as a string, and returns a bool
// of whether it is valid.
func isValidContentType(ct string) bool {
	return strings.Contains(ct, "text/html") ||
		strings.Contains(ct, "text/plain")
}
