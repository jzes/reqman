package request

import "github.com/jzes/reqman/internal/domain/request"

type requestLoader interface {
	Load(path string) (request.Request, error)
}

type RequestLoader = requestLoader
