package request

import (
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Method string

const (
	MethodGet     Method = http.MethodGet
	MethodHead    Method = http.MethodHead
	MethodPost    Method = http.MethodPost
	MethodPut     Method = http.MethodPut
	MethodPatch   Method = http.MethodPatch
	MethodDelete  Method = http.MethodDelete
	MethodConnect Method = http.MethodConnect
	MethodOptions Method = http.MethodOptions
	MethodTrace   Method = http.MethodTrace
)

type URL struct {
	raw   string
	value *url.URL
}

func NewURL(rawURL string) (URL, error) {
	original := rawURL
	if rawURL != "" && !strings.Contains(rawURL, "://") {
		rawURL = "http://" + rawURL
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return URL{}, err
	}

	return URL{raw: original, value: parsedURL}, nil
}

func (u URL) String() string {
	if u.raw != "" {
		return u.raw
	}

	if u.value == nil {
		return ""
	}

	return u.value.String()
}

type Request struct {
	Name    string
	Path    string
	URL     URL
	Method  Method
	Headers map[string]string
	Body    string
}

type Response struct {
	URL        URL
	Status     string
	StatusCode int
	Headers    map[string][]string
	Body       string
	Duration   time.Duration
}
