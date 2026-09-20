package request

import (
	"context"

	"github.com/jzes/reqman/internal/domain/request"
)

type requestDoer interface {
	Do(context.Context, request.Request) (request.Response, error)
}

type RequestDoer = requestDoer
