package request

import domainrequest "github.com/jzes/reqman/internal/domain/request"

type Doer interface {
	Do(request domainrequest.Request) (domainrequest.Response, error)
}
