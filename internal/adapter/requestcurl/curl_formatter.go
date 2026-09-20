package requestcurl

import (
	"fmt"
	"strings"

	"github.com/jzes/reqman/internal/domain/request"
)

func Format(request request.Request) ([]byte, error) {
	args := []string{"curl"}
	if request.Method != "" {
		args = append(args, "-X", string(request.Method))
	}

	for _, header := range request.Headers.List() {
		args = append(args, "-H", fmt.Sprintf("%s: %s", header.Key, header.Value))
	}

	if request.Body != "" {
		args = append(args, "-d", request.Body)
	}

	if request.URL.String() != "" {
		args = append(args, request.URL.String())
	}

	return []byte(strings.Join(quoteArgs(args), " ")), nil
}

func quoteArgs(args []string) []string {
	quoted := make([]string, len(args))
	for i, arg := range args {
		quoted[i] = shellQuote(arg)
	}
	return quoted
}

func shellQuote(arg string) string {
	return "'" + strings.ReplaceAll(arg, "'", "'\\''") + "'"
}
