package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCreatesDefaultAndReadsConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cfg, err := Load()
	if err != nil || cfg != Default() {
		t.Fatalf("Load() = %+v, %v", cfg, err)
	}
	path := filepath.Join(home, ".config", "reqman", "config.yaml")
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "editor-mode: internal\n" {
		t.Fatalf("config = %q, %v", data, err)
	}
	if err := os.WriteFile(path, []byte("editor-mode: gui\n"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err = Load()
	if err != nil || cfg.EditorMode != GUI {
		t.Fatalf("Load() = %+v, %v", cfg, err)
	}
}

func TestParse(t *testing.T) {
	for _, mode := range []EditorMode{InLine, SidePanelTmux, SidePanelHerdr, GUI, Internal} {
		t.Run(string(mode), func(t *testing.T) {
			cfg, err := parse([]byte("editor-mode: " + string(mode)))
			if err != nil || cfg.EditorMode != mode {
				t.Fatalf("parse = %+v, %v", cfg, err)
			}
		})
	}
	for _, input := range []string{"editor-mode: other", "other: internal", "editor-mode: [gui", "editor-mode: ''", "editor-mode: gui\n---\neditor-mode: internal", "editor-mode: gui\neditor-mode: internal"} {
		if _, err := parse([]byte(input)); err == nil {
			t.Errorf("parse(%q) succeeded", input)
		}
	}
	for _, input := range []string{"", "# comment\n", "{}"} {
		cfg, err := parse([]byte(input))
		if err != nil || cfg != Default() {
			t.Fatalf("parse(%q) = %+v, %v", input, cfg, err)
		}
	}
}

func TestLoadRejectsInvalidConfigWithoutOverwriting(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if _, err := Load(); err != nil {
		t.Fatal(err)
	}
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".config", "reqman", "config.yaml")
	input := []byte("editor-mode: invalid\n")
	if err := os.WriteFile(path, input, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(); err == nil {
		t.Fatal("expected error")
	}
	data, _ := os.ReadFile(path)
	if string(data) != string(input) {
		t.Fatal("existing config overwritten")
	}
}
