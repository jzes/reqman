package application

import (
	"testing"

	"github.com/jzes/reqman/internal/domain/request"
)

type fakeLister struct {
	requests []request.Request
}

func (l fakeLister) List(string) ([]request.Request, error) {
	return l.requests, nil
}

func TestRequestLoaderUsesLister(t *testing.T) {
	parsedURL, err := request.NewURL("http://localhost/books")
	if err != nil {
		t.Fatalf("NewURL() error = %v", err)
	}

	loader := NewRequestLoader(fakeLister{requests: []request.Request{
		{
			Name:   "req.custom",
			Path:   "req.custom",
			URL:    parsedURL,
			Method: request.MethodGet,
		},
	}})

	requests, err := loader.LoadFromDirectory("unused")
	if err != nil {
		t.Fatalf("LoadFromDirectory() error = %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("len(requests) = %d, want 1", len(requests))
	}
	if requests[0].Name != "req.custom" {
		t.Fatalf("Name = %q, want %q", requests[0].Name, "req.custom")
	}
	if requests[0].URL.String() != "http://localhost/books" {
		t.Fatalf("URL = %q, want %q", requests[0].URL.String(), "http://localhost/books")
	}
}
