package requestfile

import "github.com/jzes/reqman/internal/domain/request"

type Parser func(name, path string, content []byte) (request.Request, error)

type Formatter func(request.Request) ([]byte, error)

type Format struct {
	Parse  Parser
	Format Formatter
}
