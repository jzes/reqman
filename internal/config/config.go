// Package config loads reqman's user configuration.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"
)

type EditorMode string

const (
	InLine         EditorMode = "in-line"
	SidePanelTmux  EditorMode = "side-panel-tmux"
	SidePanelHerdr EditorMode = "side-panel-herdr"
	GUI            EditorMode = "gui"
	Internal       EditorMode = "internal"
)

type Config struct {
	EditorMode EditorMode `yaml:"editor-mode"`
}

func Default() Config { return Config{EditorMode: Internal} }

// Load creates the default config on first use. The path is deliberately the
// same on all platforms, including macOS.
func Load() (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}
	path := filepath.Join(home, ".config", "reqman", "config.yaml")
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return Config{}, err
		}
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if errors.Is(err, os.ErrExist) {
			return Load()
		}
		if err != nil {
			return Config{}, err
		}
		_, writeErr := file.WriteString("editor-mode: internal\n")
		closeErr := file.Close()
		return Default(), errors.Join(writeErr, closeErr)
	}
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	return parse(data)
}

func parse(data []byte) (Config, error) {
	cfg := Default()
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil && !errors.Is(err, io.EOF) {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return Config{}, fmt.Errorf("config must contain one YAML document")
	}
	switch cfg.EditorMode {
	case InLine, SidePanelTmux, SidePanelHerdr, GUI, Internal:
		return cfg, nil
	default:
		return Config{}, fmt.Errorf("invalid editor-mode %q", cfg.EditorMode)
	}
}
