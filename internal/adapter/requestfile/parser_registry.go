package requestfile

import "github.com/jzes/reqman/internal/domain/request"

type Parser func(Item) (request.Request, error)
