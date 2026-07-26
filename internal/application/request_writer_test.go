package application

import (
	"path/filepath"
	"testing"

	"github.com/jzes/reqman/internal/domain/request"
)

type fakeWriter struct {
	request request.Request
}

func (w *fakeWriter) Write(request request.Request) error {
	w.request = request
	return nil
}

func TestRequestWriterUsesWriter(t *testing.T) {
	parsedURL, err := request.NewURL("http://localhost/books")
	if err != nil {
		t.Fatalf("NewURL() error = %v", err)
	}

	writer := &fakeWriter{}
	usecase := NewRequestWriter(writer, t.TempDir())
	input := request.Request{
		Name:   "req.curl",
		Path:   "req.curl",
		URL:    parsedURL,
		Method: request.MethodPost,
		Body:   `{"name":"x"}`,
	}

	if err := usecase.WriteToFile(input); err != nil {
		t.Fatalf("WriteToFile() error = %v", err)
	}
	if writer.request.Name != "req.curl" {
		t.Fatalf("Name = %q, want %q", writer.request.Name, "req.curl")
	}
	if writer.request.URL.String() != "http://localhost/books" {
		t.Fatalf("URL = %q, want %q", writer.request.URL.String(), "http://localhost/books")
	}
}

func TestRequestWriterSetsPathFromDirectoryWhenRequestHasNoPath(t *testing.T) {
	writer := &fakeWriter{}
	dir := t.TempDir()
	usecase := NewRequestWriter(writer, dir)

	if err := usecase.WriteToFile(request.Request{Name: "req.curl"}); err != nil {
		t.Fatalf("WriteToFile() error = %v", err)
	}

	want := filepath.Join(dir, "req.curl")
	if writer.request.Path != want {
		t.Fatalf("Path = %q, want %q", writer.request.Path, want)
	}
}
