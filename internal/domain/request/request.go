package request

import "time"

type Request struct {
	Name    string
	Path    string
	URL     URL
	Method  Method
	Headers Headers
	Body    string
}

type Response struct {
	URL        URL
	Status     string
	StatusCode int
	Headers    Headers
	Body       string
	Duration   time.Duration
}
