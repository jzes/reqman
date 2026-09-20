package request

import (
	domainrequest "github.com/jzes/reqman/internal/domain/request"
)

type resultTarget interface {
	SetRequestInFlight(bool)
	SetResponseError(string)
	SetResponse(domainrequest.Response)
	SetHasResponse(bool)
}

type RequestResultMessage struct {
	RequestID int
	Response  domainrequest.Response
	Err       error
	Fatal     bool
}

func (msg RequestResultMessage) UpdateTarget(target resultTarget) {
	target.SetRequestInFlight(false)
	if msg.Err != nil {
		target.SetResponseError(msg.Err.Error())
		target.SetHasResponse(false)
		target.SetResponse(domainrequest.Response{})
		return
	}

	target.SetResponseError("")
	target.SetResponse(msg.Response)
	target.SetHasResponse(true)
}
