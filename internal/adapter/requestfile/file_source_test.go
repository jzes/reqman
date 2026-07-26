package requestfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jzes/reqman/internal/domain/request"
)

func TestFileSourceWritesRequestToFile(t *testing.T) {
	parsedURL, err := request.NewURL("http://localhost/books")
	if err != nil {
		t.Fatalf("NewURL() error = %v", err)
	}

	path := filepath.Join(t.TempDir(), "req.curl")
	source := NewFileSource(nil)
	err = source.Write(request.Request{
		Path:   path,
		URL:    parsedURL,
		Method: request.MethodPost,
		Headers: map[string]string{
			"Content-Type": "application/json",
			"X-Test":       "yes",
		},
		Body: `{"name":"x"}`,
	})
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	want := `'curl' '-X' 'POST' '-H' 'Content-Type: application/json' '-H' 'X-Test: yes' '-d' '{"name":"x"}' 'http://localhost/books'`
	if string(content) != want {
		t.Fatalf("content = %q, want %q", string(content), want)
	}
}

func TestFileSourceWritesMultilineJSONBodyRoundTrip(t *testing.T) {
	parsedURL, err := request.NewURL("http://localhost/books")
	if err != nil {
		t.Fatalf("NewURL() error = %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "req.curl")
	source := NewFileSource(map[string]Parser{CurlFileExtension: ParseCurlRequest})
	body := "{\n  \"name\": \"x\",\n  \"active\": true\n}"
	err = source.Write(request.Request{
		Path:   path,
		URL:    parsedURL,
		Method: request.MethodPost,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: body,
	})
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if strings.Contains(string(content), `\n`) {
		t.Fatalf("content contains escaped newlines: %q", string(content))
	}

	requests, err := source.List(dir)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("requests len = %d, want 1", len(requests))
	}
	if got := requests[0].Body; got != body {
		t.Fatalf("Body = %q, want %q", got, body)
	}
}

func TestFileSourceWritesEmptyRequestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.curl")
	source := NewFileSource(map[string]Parser{CurlFileExtension: ParseCurlRequest})
	err := source.Write(request.Request{
		Name:    "empty.curl",
		Path:    path,
		Method:  request.MethodGet,
		Headers: map[string]string{},
	})
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	requests, err := source.List(dir)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("requests len = %d, want 1", len(requests))
	}
	if got := requests[0].Name; got != "empty.curl" {
		t.Fatalf("Name = %q, want empty.curl", got)
	}
	if got := requests[0].URL.String(); got != "" {
		t.Fatalf("URL = %q, want empty", got)
	}
	if got := requests[0].Method; got != request.MethodGet {
		t.Fatalf("Method = %q, want GET", got)
	}
}

func TestFileSourceWriteReturnsError(t *testing.T) {
	source := NewFileSource(nil)
	err := source.Write(request.Request{Path: filepath.Join(t.TempDir(), "missing", "req.curl")})
	if err == nil {
		t.Fatal("Write() error = nil, want error")
	}
}
