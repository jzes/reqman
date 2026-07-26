package application

import (
	"path/filepath"

	applicationrequest "github.com/jzes/reqman/internal/application/request"
	"github.com/jzes/reqman/internal/domain/request"
)

type RequestWriter struct {
	requestWriter    applicationrequest.Writer
	requestDirectory string
}

func NewRequestWriter(requestWriter applicationrequest.Writer, requestDirectory string) RequestWriter {
	if requestDirectory == "" {
		requestDirectory = "."
	}
	return RequestWriter{requestWriter: requestWriter, requestDirectory: requestDirectory}
}

func (w RequestWriter) WriteToFile(request request.Request) error {
	if request.Path == "" && request.Name != "" {
		request.Path = filepath.Join(w.requestDirectory, request.Name)
	}
	return w.requestWriter.Write(request)
}
