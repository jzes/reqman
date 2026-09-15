package requesthttp

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	applicationrequest "github.com/jzes/reqman/internal/application/request"
	"github.com/jzes/reqman/internal/domain/request"
)

const defaultRequestTimeout = 30 * time.Second

type Client struct {
	httpClient *http.Client
}

var _ applicationrequest.Doer = Client{}

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

	for key, values := range req.Headers {
		for _, value := range values {
			httpRequest.Header.Add(key, value)
		}
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

func cloneResponseHeaders(headers http.Header) map[string][]string {
	cloned := make(map[string][]string, len(headers))
	for key, values := range headers {
		cloned[key] = append([]string(nil), values...)
	}
	return cloned
}

func requestURL(request request.Request) string {
	rawURL := request.URL.String()
	if strings.Contains(rawURL, "://") {
		return rawURL
	}
	return "http://" + rawURL
}
