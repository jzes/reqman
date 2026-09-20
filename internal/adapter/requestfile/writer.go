package requestfile

import (
	"path/filepath"

	"github.com/jzes/reqman/internal/domain/request"
)

const defaultDir = "."

type FileWriter struct {
	fileSource FileSource
	dir        string
}

func NewFileWriter(fileSource FileSource, dir string) FileWriter {
	if dir == "" {
		return FileWriter{fileSource: fileSource, dir: defaultDir}
	}
	return FileWriter{fileSource: fileSource, dir: dir}
}

func (w FileWriter) WriteToFile(req request.Request) error {
	if req.Path == "" && req.Name != "" {
		req.Path = filepath.Join(w.dir, req.Name)
	}
	return w.fileSource.Write(req)
}
