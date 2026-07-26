package request

import "github.com/jzes/reqman/internal/domain/request"

type requestWriter interface {
	WriteToFile(request.Request) error
}

type RequestWriter = requestWriter
