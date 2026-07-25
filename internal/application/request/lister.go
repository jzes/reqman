// Package request provides interfaces for request operations.
package request

import "github.com/jzes/reqman/internal/domain/request"

type Lister interface {
	List(dir string) ([]request.Request, error)
}
