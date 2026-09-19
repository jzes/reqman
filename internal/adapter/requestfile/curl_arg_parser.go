package requestfile

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jzes/reqman/internal/domain/request"
)

type curlArgParser struct {
	request *request.Request
	rawURL  string
}

func (s *curlArgParser) handleArg(args []string, index int) (int, error) {
	arg := args[index]

	switch {
	case isLocationFlag(arg):
		return index, nil
	case isMethodFlag(arg):
		return s.handleMethodFlag(args, index)
	case isHeaderFlag(arg):
		return s.handleHeaderFlag(args, index)
	case isBodyFlag(arg):
		return s.handleBodyFlag(args, index)
	case strings.HasPrefix(arg, "-"):
		return skipUnknownFlag(args, index), nil
	case s.rawURL == "":
		s.rawURL = arg
	}

	return index, nil
}

func (s *curlArgParser) handleMethodFlag(args []string, index int) (int, error) {
	value, next, err := flagValueFromArg(args, index, methodFlagPrefixes)
	if err != nil {
		return index, err
	}

	method, err := parseMethod(value)
	if err != nil {
		return index, err
	}

	s.request.Method = method
	return next, nil
}

func (s *curlArgParser) handleHeaderFlag(args []string, index int) (int, error) {
	value, next, err := flagValueFromArg(args, index, headerFlagPrefixes)
	if err != nil {
		return index, err
	}

	if err := addHeader(&s.request.Headers, value); err != nil {
		return index, err
	}
	return next, nil
}

func (s *curlArgParser) handleBodyFlag(args []string, index int) (int, error) {
	value, next, err := flagValueFromArg(args, index, bodyFlagPrefixes)
	if err != nil {
		return index, err
	}
	if s.request.Body != "" {
		return index, fmt.Errorf("multiple request bodies are not supported")
	}
	value = normalizeJSONBodyNewlines(value)
	if !json.Valid([]byte(value)) {
		return index, fmt.Errorf("request body must be valid JSON")
	}

	s.request.Body = value
	if strings.HasPrefix(args[index], "--json") {
		if err := addHeaderIfMissing(&s.request.Headers, "Content-Type", "application/json"); err != nil {
			return index, err
		}
	}

	return next, nil
}

func normalizeJSONBodyNewlines(value string) string {
	if json.Valid([]byte(value)) {
		return value
	}

	withNewlines := strings.ReplaceAll(value, `\n`, "\n")
	if json.Valid([]byte(withNewlines)) {
		return withNewlines
	}

	return value
}
