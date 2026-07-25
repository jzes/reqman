package application

import (
	applicationrequest "github.com/jzes/reqman/internal/application/request"
	"github.com/jzes/reqman/internal/domain/request"
)

type RequestWriter struct {
	requestWriter applicationrequest.Writer
}

func NewRequestWriter(requestWriter applicationrequest.Writer) RequestWriter {
	return RequestWriter{requestWriter: requestWriter}
}

func (w RequestWriter) WriteToFile(request request.Request) error {
	return w.requestWriter.Write(request)
}
