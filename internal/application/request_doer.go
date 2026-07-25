package application

import (
	applicationrequest "github.com/jzes/reqman/internal/application/request"
	domainrequest "github.com/jzes/reqman/internal/domain/request"
)

type RequestDoer struct {
	httpRequestDoer applicationrequest.Doer
}

func NewRequestDoer(httpRequestDoer applicationrequest.Doer) RequestDoer {
	return RequestDoer{httpRequestDoer: httpRequestDoer}
}

func (d RequestDoer) Do(request domainrequest.Request) (domainrequest.Response, error) {
	return d.httpRequestDoer.Do(request)
}
