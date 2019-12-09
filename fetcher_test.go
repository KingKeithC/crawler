// Fetcher is an object which can be used to fetch the contents of a webpage.
package main

import (
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
