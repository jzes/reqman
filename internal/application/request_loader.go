// Package application provides use cases for loading domain requests.
package application

import (
	applicationrequest "github.com/jzes/reqman/internal/application/request"
	"github.com/jzes/reqman/internal/domain/request"
)

type RequestLoader struct {
	requestLister applicationrequest.Lister
}

func NewRequestLoader(requestLister applicationrequest.Lister) RequestLoader {
	return RequestLoader{requestLister: requestLister}
}

func (l RequestLoader) LoadFromDirectory(dir string) ([]request.Request, error) {
	return l.requestLister.List(dir)
}
