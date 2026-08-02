package application

import (
	"testing"

	"github.com/jzes/reqman/internal/domain/request"
)

type fakeLister struct {
	paths    []string
	requests []request.Request
}

func (l fakeLister) List(string) ([]string, error) {
	return l.paths, nil
}

func (l fakeLister) Load(string) (request.Request, error) {
	if len(l.requests) == 0 {
		return request.Request{}, nil
	}
	return l.requests[0], nil
}

func TestRequestLoaderUsesLister(t *testing.T) {
	loader := NewRequestLoader(fakeLister{paths: []string{"req.custom"}})

	paths, err := loader.List("unused")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(paths) != 1 {
		t.Fatalf("len(paths) = %d, want 1", len(paths))
	}
	if paths[0] != "req.custom" {
		t.Fatalf("path = %q, want req.custom", paths[0])
	}
}

func TestRequestLoaderUsesLoader(t *testing.T) {
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

	loaded, err := loader.Load("req.custom")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.Name != "req.custom" {
		t.Fatalf("Name = %q, want %q", loaded.Name, "req.custom")
	}
	if loaded.URL.String() != "http://localhost/books" {
		t.Fatalf("URL = %q, want %q", loaded.URL.String(), "http://localhost/books")
	}
}
