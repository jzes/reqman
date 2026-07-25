package requestfile

import (
	"os"
	"path/filepath"
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
	want := `"curl" "-X" "POST" "-H" "Content-Type: application/json" "-H" "X-Test: yes" "-d" "{\"name\":\"x\"}" "http://localhost/books"`
	if string(content) != want {
		t.Fatalf("content = %q, want %q", string(content), want)
	}
}

func TestFileSourceWriteReturnsError(t *testing.T) {
	source := NewFileSource(nil)
	err := source.Write(request.Request{Path: filepath.Join(t.TempDir(), "missing", "req.curl")})
	if err == nil {
		t.Fatal("Write() error = nil, want error")
	}
}
