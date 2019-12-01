// Fetcher is an object which can be used to fetch the contents of a webpage.
package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultFetcherClient is the http Client that newly created fetchers use
var DefaultFetcherClient = &http.Client{Timeout: time.Duration(10) * time.Second}

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

// FetchWebpage takes the URL to a web page, and returns the HTML on the given page,
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

// isValidURL takes a URL as a string, and returns a bool of whether it is valid.
func isValidURL(u string) bool {
	parsed, err := url.Parse(u)
	if err != nil {
		return false
	}
	return parsed.IsAbs()
}

// isValidContentType takes a content-type header as a string, and returns a bool of whether it is valid.
func isValidContentType(ct string) bool {
	return strings.Contains(ct, "text/html") ||
		strings.Contains(ct, "text/plain")
}
