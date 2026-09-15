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
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
			"X-Test":       {"yes"},
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

func TestFileSourceWriteDoesNotLeaveTempFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "req.curl")
	source := NewFileSource(nil)

	if err := source.Write(request.Request{Path: path, Method: request.MethodGet}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("entries len = %d, want 1", len(entries))
	}
	if got := entries[0].Name(); got != "req.curl" {
		t.Fatalf("entry name = %q, want req.curl", got)
	}
}

func TestFileSourceWritePreservesExistingFilePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "req.curl")
	if err := os.WriteFile(path, []byte("old"), 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	source := NewFileSource(nil)
	if err := source.Write(request.Request{Path: path, Method: request.MethodGet}); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("file permissions = %v, want %v", got, os.FileMode(0600))
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
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
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

	paths, err := source.List(dir)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(paths) != 1 {
		t.Fatalf("paths len = %d, want 1", len(paths))
	}
	loaded, err := source.Load(paths[0])
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got := loaded.Body; got != body {
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
		Headers: map[string][]string{},
	})
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	paths, err := source.List(dir)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(paths) != 1 {
		t.Fatalf("paths len = %d, want 1", len(paths))
	}
	if got := paths[0]; got != path {
		t.Fatalf("path = %q, want %q", got, path)
	}
	loaded, err := source.Load(paths[0])
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got := loaded.Name; got != "empty.curl" {
		t.Fatalf("Name = %q, want empty.curl", got)
	}
	if got := loaded.URL.String(); got != "" {
		t.Fatalf("URL = %q, want empty", got)
	}
	if got := loaded.Method; got != request.MethodGet {
		t.Fatalf("Method = %q, want GET", got)
	}
}

func TestFileSourceWritesRepeatedHeaders(t *testing.T) {
	parsedURL, err := request.NewURL("http://localhost/books")
	if err != nil {
		t.Fatalf("NewURL() error = %v", err)
	}

	path := filepath.Join(t.TempDir(), "req.curl")
	source := NewFileSource(nil)
	err = source.Write(request.Request{
		Path:   path,
		URL:    parsedURL,
		Method: request.MethodGet,
		Headers: map[string][]string{
			"Accept": {"application/json", "text/plain"},
		},
	})
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	want := `'curl' '-X' 'GET' '-H' 'Accept: application/json' '-H' 'Accept: text/plain' 'http://localhost/books'`
	if string(content) != want {
		t.Fatalf("content = %q, want %q", string(content), want)
	}
}

func TestFileSourceListSkipsInvalidCurlContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.curl")
	if err := os.WriteFile(path, []byte("curl -X"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	source := NewFileSource(map[string]Parser{CurlFileExtension: ParseCurlRequest})
	paths, err := source.List(dir)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(paths) != 1 || paths[0] != path {
		t.Fatalf("paths = %v, want [%q]", paths, path)
	}

	if _, err := source.Load(path); err == nil {
		t.Fatal("Load() error = nil, want error")
	}
}

func TestFileSourceWriteReturnsError(t *testing.T) {
	source := NewFileSource(nil)
	err := source.Write(request.Request{Path: filepath.Join(t.TempDir(), "missing", "req.curl")})
	if err == nil {
		t.Fatal("Write() error = nil, want error")
	}
}
