package requestfile

import (
	"fmt"

	"github.com/jzes/reqman/internal/domain/request"
	"mvdan.cc/sh/v3/shell"
)

const CurlFileExtension = ".curl"

func ParseCurlRequest(item Item) (request.Request, error) {
	args, err := curlArgs(item.Content)
	if err != nil {
		return request.Request{}, fmt.Errorf("parse curl file %q: %w", item.Path, err)
	}

	parsed, err := requestFromArgs(args)
	if err != nil {
		return request.Request{}, fmt.Errorf("parse curl file %q: %w", item.Path, err)
	}

	parsed.Name = item.Name
	parsed.Path = item.Path

	return parsed, nil
}

func curlArgs(content []byte) ([]string, error) {
	args, err := shell.Fields(string(content), func(string) string { return "" })
	if err != nil {
		return nil, err
	}

	if len(args) == 0 {
		return nil, fmt.Errorf("missing curl command")
	}

	if args[0] != "curl" {
		return nil, fmt.Errorf("expected curl command")
	}

	return args[1:], nil
}

func requestFromArgs(args []string) (request.Request, error) {
	parsed := request.Request{
		Method:  request.MethodGet,
		Headers: make(map[string][]string),
	}

	state := curlArgParser{request: &parsed}
	for i := 0; i < len(args); i++ {
		next, err := state.handleArg(args, i)
		if err != nil {
			return request.Request{}, err
		}
		i = next
	}

	if state.rawURL != "" {
		parsedURL, err := request.NewURL(state.rawURL)
		if err != nil {
			return request.Request{}, err
		}

		parsed.URL = parsedURL
	}
	if parsed.Body != "" && parsed.Method == request.MethodGet {
		parsed.Method = request.MethodPost
	}

	return parsed, nil
}
