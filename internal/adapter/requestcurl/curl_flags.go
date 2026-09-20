package requestcurl

import (
	"fmt"
	"strings"

	"github.com/jzes/reqman/internal/domain/request"
)

var methodFlagPrefixes = []string{"--request=", "-X"}
var headerFlagPrefixes = []string{"--header=", "-H"}
var bodyFlagPrefixes = []string{"--json=", "--data=", "--data-raw=", "--data-binary=", "--data-ascii=", "-d"}

func isLocationFlag(arg string) bool {
	return arg == "-L" || arg == "--location"
}

func isMethodFlag(arg string) bool {
	return arg == "-X" || arg == "--request" || strings.HasPrefix(arg, "--request=") || compactFlagValue(arg, "-X") != ""
}

func isHeaderFlag(arg string) bool {
	return arg == "-H" || arg == "--header" || strings.HasPrefix(arg, "--header=") || compactFlagValue(arg, "-H") != ""
}

func isBodyFlag(arg string) bool {
	return arg == "--json" || isDataFlag(arg) || strings.HasPrefix(arg, "--json=") || hasAnyPrefix(arg, bodyFlagPrefixes)
}

func flagValueFromArg(args []string, index int, compactPrefixes []string) (string, int, error) {
	arg := args[index]
	if value, ok := strings.CutPrefix(arg, "--request="); ok {
		return value, index, nil
	}
	if value, ok := strings.CutPrefix(arg, "--header="); ok {
		return value, index, nil
	}
	for _, prefix := range compactPrefixes {
		if value := compactFlagValue(arg, prefix); value != "" {
			return value, index, nil
		}
	}

	return flagValue(args, index)
}

func compactFlagValue(arg, prefix string) string {
	if arg == prefix || !strings.HasPrefix(arg, prefix) {
		return ""
	}
	return strings.TrimPrefix(arg, prefix)
}

func hasAnyPrefix(arg string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(arg, prefix) {
			return true
		}
	}
	return false
}

func flagValue(args []string, index int) (string, int, error) {
	if index+1 >= len(args) {
		return "", index, fmt.Errorf("missing value for %s", args[index])
	}

	return args[index+1], index + 1, nil
}

func parseMethod(value string) (request.Method, error) {
	switch strings.ToUpper(value) {
	case string(request.MethodGet):
		return request.MethodGet, nil
	case string(request.MethodHead):
		return request.MethodHead, nil
	case string(request.MethodPost):
		return request.MethodPost, nil
	case string(request.MethodPut):
		return request.MethodPut, nil
	case string(request.MethodPatch):
		return request.MethodPatch, nil
	case string(request.MethodDelete):
		return request.MethodDelete, nil
	case string(request.MethodConnect):
		return request.MethodConnect, nil
	case string(request.MethodOptions):
		return request.MethodOptions, nil
	case string(request.MethodTrace):
		return request.MethodTrace, nil
	default:
		return "", fmt.Errorf("unsupported HTTP method %q", value)
	}
}

func addHeader(headers *request.Headers, value string) error {
	name, headerValue, ok := strings.Cut(value, ":")
	if !ok {
		return nil
	}

	name = strings.TrimSpace(name)
	return headers.Add(name, strings.TrimSpace(headerValue))
}

func addHeaderIfMissing(headers *request.Headers, name, value string) error {
	return headers.AddIfMissing(name, value)
}

func isDataFlag(arg string) bool {
	switch arg {
	case "-d", "--data", "--data-raw", "--data-binary", "--data-ascii":
		return true
	default:
		return false
	}
}

func skipUnknownFlag(args []string, index int) int {
	if flagHasSeparateValue(args[index]) && index+1 < len(args) {
		return index + 1
	}
	return index
}

func flagHasSeparateValue(arg string) bool {
	if strings.Contains(arg, "=") {
		return false
	}

	if strings.HasPrefix(arg, "--") {
		return true
	}

	switch arg {
	case "-A", "-b", "-e", "-F", "-m", "-o", "-u":
		return true
	default:
		return false
	}
}
