// Package requestfile is a package that provides functionality
// to load request items from files in a specified directory.
// It defines a FileSource struct that implements the requestload.
// Source interface, allowing users to retrieve request items from files.
package requestfile

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jzes/reqman/internal/domain/request"
)

type Item struct {
	Name      string
	Extension string
	Path      string
	Content   []byte
}

type FileSource struct {
	parsers map[string]Parser
}

func NewFileSource(parsers map[string]Parser) FileSource {
	return FileSource{parsers: parsers}
}

func (s FileSource) List(dir string) ([]request.Request, error) {
	entries, err := readDirectoryEntries(dir)
	if err != nil {
		return nil, err
	}

	items := []request.Request{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		item, err := readItem(dir, entry)
		if err != nil {
			return nil, err
		}

		parser, ok := s.parsers[item.Extension]
		if !ok {
			continue
		}

		parsed, err := parser(item)
		if err != nil {
			return nil, fmt.Errorf("parse request file %q: %w", item.Path, err)
		}

		items = append(items, parsed)
	}

	return items, nil
}

func (s FileSource) Write(request request.Request) error {
	content := requestContent(request)
	return os.WriteFile(request.Path, []byte(content), 0644)
}

func requestContent(request request.Request) string {
	args := []string{"curl"}
	if request.Method != "" {
		args = append(args, "-X", string(request.Method))
	}

	for _, key := range sortedHeaderKeys(request.Headers) {
		args = append(args, "-H", fmt.Sprintf("%s: %s", key, request.Headers[key]))
	}

	if request.Body != "" {
		args = append(args, "-d", request.Body)
	}

	if request.URL.String() != "" {
		args = append(args, request.URL.String())
	}

	return strings.Join(quoteArgs(args), " ")
}

func sortedHeaderKeys(headers map[string]string) []string {
	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
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

func readItem(dir string, entry os.DirEntry) (Item, error) {
	path := filepath.Join(dir, entry.Name())
	content, err := os.ReadFile(path)
	if err != nil {
		return Item{}, err
	}

	return Item{
		Name:      entry.Name(),
		Extension: filepath.Ext(entry.Name()),
		Path:      path,
		Content:   content,
	}, nil
}

func readDirectoryEntries(dir string) ([]os.DirEntry, error) {
	file, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return file.ReadDir(-1)
}
