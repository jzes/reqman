// Package requestfile is a package that provides functionality
// to load request items from files in a specified directory.
// It defines a FileSource struct that implements the requestload.
// Source interface, allowing users to retrieve request items from files.
package requestfile

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jzes/reqman/internal/domain/request"
)

type Item struct {
	Name      string
	Extension string
	Path      string
	Content   []byte
}

type FileSource struct {
	formats map[string]Format
}

func NewFileSource(formats map[string]Format) FileSource {
	return FileSource{formats: formats}
}

func (s FileSource) List(dir string) ([]string, error) {
	entries, err := readDirectoryEntries(dir)
	if err != nil {
		return nil, err
	}

	paths := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		extension := filepath.Ext(entry.Name())
		_, ok := s.formats[extension]
		if !ok {
			continue
		}

		paths = append(paths, filepath.Join(dir, entry.Name()))
	}

	return paths, nil
}

func (s FileSource) Load(path string) (request.Request, error) {
	item, err := readItem(path)
	if err != nil {
		return request.Request{}, err
	}

	format, ok := s.formats[item.Extension]
	if !ok || format.Parse == nil {
		return request.Request{}, fmt.Errorf("unsupported request file extension %q", item.Extension)
	}

	parsed, err := format.Parse(item.Name, item.Path, item.Content)
	if err != nil {
		return request.Request{}, fmt.Errorf("parse request file %q: %w", item.Path, err)
	}

	return parsed, nil
}

func (s FileSource) Write(request request.Request) error {
	extension := filepath.Ext(request.Path)
	format, ok := s.formats[extension]
	if !ok || format.Format == nil {
		return fmt.Errorf("unsupported request file extension %q", extension)
	}

	content, err := format.Format(request)
	if err != nil {
		return err
	}

	return writeFileAtomically(request.Path, content)
}

func writeFileAtomically(path string, content []byte) error {
	fileMode, err := getFileMode(path)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	tempFile, err := os.CreateTemp(
		dir,
		fmt.Sprintf(".%s.*.tmp", base),
	)
	if err != nil {
		return err
	}

	tempPath := tempFile.Name()
	removeTemp := true
	defer func() {
		if removeTemp {
			_ = os.Remove(tempPath)
		}
	}()

	if _, err := tempFile.Write(content); err != nil {
		_ = tempFile.Close()
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tempPath, fileMode); err != nil {
		return err
	}
	if err := os.Rename(tempPath, path); err != nil {
		return err
	}

	removeTemp = false
	return nil
}

func getFileMode(path string) (os.FileMode, error) {
	info, err := os.Stat(path)
	if err == nil {
		return info.Mode().Perm(), nil
	}
	if os.IsNotExist(err) {
		return os.FileMode(0644), nil
	}
	return 0, err
}

func readItem(path string) (Item, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Item{}, err
	}

	name := filepath.Base(path)
	return Item{
		Name:      name,
		Extension: filepath.Ext(name),
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
