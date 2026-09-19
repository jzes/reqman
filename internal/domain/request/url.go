package request

import (
	"net/url"
	"strings"
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
