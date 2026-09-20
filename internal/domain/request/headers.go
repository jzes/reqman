package request

import (
	"errors"
	"sort"
	"strings"
)

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

func (h Headers) List() []Header {
	keys := h.Names()

	headers := make([]Header, 0, len(h.values))
	for _, key := range keys {
		for _, value := range h.values[key] {
			headers = append(headers, Header{Key: key, Value: value})
		}
	}
	return headers
}
