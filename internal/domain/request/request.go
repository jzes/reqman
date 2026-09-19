package request

import (
	"errors"
	"net/http"
	"net/url"
	"sort"
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

type Header struct {
	Key   string
	Value string
}

type Headers struct {
	values map[string][]string
}

func NewHeaders() Headers {
	return Headers{values: make(map[string][]string)}
}

func NewHeadersFrom(values map[string][]string) Headers {
	headers := NewHeaders()
	for key, values := range values {
		for _, value := range values {
			_ = headers.Add(key, value)
		}
	}
	return headers
}

func (h *Headers) Add(key, value string) error {
	if h.values == nil {
		return errors.New("domain: headers: Headers must be initialized")
	}
	h.values[key] = append(h.values[key], value)
	return nil
}

func (h *Headers) AddIfMissing(key, value string) error {
	for existingKey := range h.values {
		if strings.EqualFold(existingKey, key) {
			return nil
		}
	}
	return h.Add(key, value)
}

func (h Headers) Values(key string) []string {
	return append([]string(nil), h.values[key]...)
}

func (h Headers) Len() int {
	return len(h.values)
}

func (h Headers) Names() []string {
	keys := make([]string, 0, len(h.values))
	for key := range h.values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (h Headers) Rows() []Header {
	keys := h.Names()

	rows := make([]Header, 0, len(h.values))
	for _, key := range keys {
		for _, value := range h.values[key] {
			rows = append(rows, Header{Key: key, Value: value})
		}
	}
	return rows
}
