package requestfile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jzes/reqman/internal/domain/request"
)

func TestWriterSetsPathFromDirectoryWhenRequestHasNoPath(t *testing.T) {
	dir := t.TempDir()
	writer := NewFileWriter(NewFileSource(nil), dir)

	if err := writer.WriteToFile(request.Request{Name: "req.curl"}); err != nil {
		t.Fatalf("WriteToFile() error = %v", err)
	}

	wantPath := filepath.Join(dir, "req.curl")
	if _, err := os.Stat(wantPath); err != nil {
		t.Fatalf("os.Stat(%q) error = %v", wantPath, err)
	}
}
