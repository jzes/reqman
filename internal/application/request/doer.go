package request

import (
	"context"

	domainrequest "github.com/jzes/reqman/internal/domain/request"
)

type Doer interface {
	Do(context.Context, domainrequest.Request) (domainrequest.Response, error)
}
