// Scraper is a package which can be used to scrape the contents of a webpage.
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
)

func Test_isValidURL(t *testing.T) {
	type args struct {
		u string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Empty String",
			args: args{""},
			want: false,
		},
		{
			name: "Valid URL",
			args: args{"https://www.kinglabs.ca/"},
			want: true,
		},
		{
			name: "URL Without Host",
			args: args{"https:///"},
			want: false,
		},
		{
			name: "URL Without Protocol",
			args: args{"www.kinglabs.ca/"},
			want: false,
		},
		{
			name: "Non HTTP Scheme",
			args: args{"ftp://ftp.kinglabs.ca/"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidURL(tt.args.u); got != tt.want {
				t.Errorf("isValidURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isValidContentType(t *testing.T) {
	type args struct {
		ct string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Test Invalid Content-Type",
			args: args{"content/octect-stream"},
			want: false,
		},
		{
			name: "Test Valid Content-Type",
			args: args{"text/html"},
			want: true,
		},
		{
			name: "Test Invalid Content-Type",
			args: args{"text/plain"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidContentType(tt.args.ct); got != tt.want {
				t.Errorf("isValidContentType() = %v, want %v", got, tt.want)
			}
		})
	}
}

// BadReader implements the Read method which always returns an error
type BadReader struct{}

// Read always returns an error
func (b *BadReader) Read(p []byte) (n int, err error) {
	return
}

// roundTripFunc .
type RoundTripFunc func(req *http.Request) (*http.Response, error)

// roundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// newTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {

	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

func TestScrape(t *testing.T) {
	h = NewTestClient(func(req *http.Request) (*http.Response, error) {
		header := &http.Header{}
		header.Set("content-type", "text/html")

		switch req.URL.String() {
		case "http://example.com/some/path":
			return &http.Response{
				StatusCode: 200,
				// Send response to be tested
				Body: ioutil.NopCloser(bytes.NewBufferString(`
				Gret Job, here are some URLs:
				<html><body>
				<a href="http://abc.123/something-cool/">One</a>
				<a href="ftp://abc.123">Two</a>
				<a href="abc.123/something-cool/#fragment">Three</a>
				</body></html>
				`)),
				// Must be set to non-nil value or it panics
				Header: *header,
			}, nil
		case "http://special.valid/url/that/simulates-a-connection/#error":
			return nil, fmt.Errorf("this is an artificial error")
		case "https://the.worst/content-type/on#theweb":
			header.Del("content-type")
			header.Set("content-type", "binary/octect-stream")
			return &http.Response{
				StatusCode: 200,
				// Send response to be tested
				Body: ioutil.NopCloser(bytes.NewBufferString(`Hungry for Octects?`)),
				// Must be set to non-nil value or it panics
				Header: *header,
			}, nil
		case "https://the.worst/body/on#theweb":
			return &http.Response{
				StatusCode: 200,
				// Send response to be tested
				Body: ioutil.NopCloser(&BadReader{}),
				// Must be set to non-nil value or it panics
				Header: *header,
			}, nil
		default:
			return &http.Response{
				StatusCode: 404,
				// Send response to be tested
				Body: ioutil.NopCloser(bytes.NewBufferString(`404 PAGE NOT FOUND`)),
				// Must be set to non-nil value or it panics
				Header: *header,
			}, nil
		}
	})

	type args struct {
		u string
	}
	tests := []struct {
		name    string
		args    args
		wantS   *Scraping
		wantErr bool
	}{
		{
			name: "Test Successful Scrape",
			args: args{u: "http://example.com/some/path"},
			wantS: &Scraping{
				URL: "http://example.com/some/path",
				RawHrefs: []string{
					"http://abc.123/something-cool/",
					"ftp://abc.123",
					"abc.123/something-cool/#fragment",
				},
				ValidHrefs: []string{"http://abc.123/something-cool/"},
			},
			wantErr: false,
		},
		{
			name:    "Test Invalid URL",
			args:    args{u: "example"},
			wantS:   nil,
			wantErr: true,
		},
		{
			name:    "Test Connection Error",
			args:    args{u: "http://special.valid/url/that/simulates-a-connection/#error"},
			wantS:   nil,
			wantErr: true,
		},
		{
			name:    "Test 404 Status Code",
			args:    args{u: "https://xyz.ca/"},
			wantS:   nil,
			wantErr: true,
		},
		{
			name:    "Test Bad Content-Type",
			args:    args{u: "https://the.worst/content-type/on#theweb"},
			wantS:   nil,
			wantErr: true,
		},
		{
			name:    "Test Bad Body Parse",
			args:    args{u: "https://the.worst/body/on#theweb"},
			wantS:   nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotS, err := Scrape(tt.args.u)
			if (err != nil) != tt.wantErr {
				t.Errorf("Scrape() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotS, tt.wantS) {
				t.Errorf("Scrape() = %v, want %v", gotS, tt.wantS)
			}
		})
	}
}
