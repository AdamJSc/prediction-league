package adapters

import (
	"net/http"
	"time"
)

// HTTPClient abstracts the execution of a HTTP request
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// RealHTTPClient implements HTTPClient by wrapping the net/http client
type RealHTTPClient struct {
	cl *http.Client
}

// Do implements HTTPClient
func (r *RealHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return r.cl.Do(req)
}

// NewRealHTTPClient returns a new RealHTTPClient with the provided timeout in seconds
func NewRealHTTPClient(tSecs int) *RealHTTPClient {
	cl := &http.Client{Timeout: time.Duration(tSecs) * time.Second}
	return &RealHTTPClient{cl}
}
