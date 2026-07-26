package request

import "github.com/jzes/reqman/internal/domain/request"

type requestDoer interface {
	Do(request.Request) (request.Response, error)
}

type RequestDoer = requestDoer
