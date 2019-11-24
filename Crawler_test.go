package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"
)

// RoundTripFunc
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip, to implement the Transport.RoundTripFunc
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

//NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestCrawler(cID int, cLog *logrus.Logger, fn RoundTripFunc) *Crawler {
	// Define a custom client with a round tripper which executes on the passed in function
	client := &http.Client{Transport: RoundTripFunc(fn)}

	return NewCrawler(cID, cLog, client)
}

// TestCrawlWebpage
func TestCrawlWebpage(t *testing.T) {
	crawler := NewTestCrawler(0, nil, func(req *http.Request) *http.Response {
		// If URL is incorrect
		if req.URL.String() != "https://website.com/webpage/" {
			t.Fatalf("Request URL was not expected")
		}

		header := make(http.Header)
		header.Set("content-type", "text/html")

		pageHTML := string(`<html><body><a href="https://www.google.ca">somelink</a><a href="#some-fragment">hi</a></body></html>`)
		return &http.Response{
			StatusCode: 200,
			// Send response to be tested
			Body: ioutil.NopCloser(bytes.NewBufferString(pageHTML)),
			// Must be set to non-nil value or it panics
			Header: header,
		}
	})

	foundUrls, err := crawler.CrawlWebpage("https://website.com/webpage/")
	if err != nil {
		t.Errorf("an error was not expected to e returned from the call to https://website.com/webpage/")
	}

	if len(foundUrls) != 1 {
		t.Errorf("length of foundUrls should be 1, was %d", len(foundUrls))
	}

	if foundUrls[0] != "https://www.google.ca" {
		t.Errorf("the first found URL was not https://www.google.ca, got %s", foundUrls[0])
	}

}
