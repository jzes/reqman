package request

import "github.com/jzes/reqman/internal/domain/request"

type Writer interface {
	Write(request request.Request) error
}
