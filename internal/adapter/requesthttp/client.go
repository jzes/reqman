package requesthttp

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jzes/reqman/internal/domain/request"
)

const defaultRequestTimeout = 30 * time.Second

type Client struct {
	httpClient *http.Client
}

func NewClient() Client {
	return Client{httpClient: &http.Client{Timeout: defaultRequestTimeout}}
}

func (c Client) Do(ctx context.Context, req request.Request) (request.Response, error) {
	start := time.Now()
	body := strings.NewReader(req.Body)
	httpRequest, err := http.NewRequestWithContext(ctx, string(req.Method), requestURL(req), body)
	if err != nil {
		return request.Response{}, err
	}

	for _, header := range req.Headers.List() {
		httpRequest.Header.Add(header.Key, header.Value)
	}

	httpClient := c.httpClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultRequestTimeout}
	}

	response, err := httpClient.Do(httpRequest)
	if err != nil {
		return request.Response{}, err
	}
	defer response.Body.Close()

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return request.Response{}, err
	}

	effectiveURL := req.URL
	if response.Request != nil && response.Request.URL != nil {
		if parsedURL, err := request.NewURL(response.Request.URL.String()); err == nil {
			effectiveURL = parsedURL
		}
	}

	return request.Response{
		URL:        effectiveURL,
		Status:     response.Status,
		StatusCode: response.StatusCode,
		Headers:    cloneResponseHeaders(response.Header),
		Body:       string(bodyBytes),
		Duration:   time.Since(start),
	}, nil
}

func cloneResponseHeaders(headers http.Header) request.Headers {
	cloned := make(map[string][]string, len(headers))
	for key, values := range headers {
		cloned[key] = append([]string(nil), values...)
	}
	return request.NewHeadersFrom(cloned)
}

func requestURL(request request.Request) string {
	rawURL := request.URL.String()
	if strings.Contains(rawURL, "://") {
		return rawURL
	}
	return "http://" + rawURL
}
