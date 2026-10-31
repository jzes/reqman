package bodyeditor

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jzes/reqman/internal/config"
)

func writeScript(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+content), 0700); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestEditorRoundTrip(t *testing.T) {
	for _, mode := range []config.EditorMode{config.InLine, config.GUI, config.SidePanelTmux, config.SidePanelHerdr} {
		t.Run(string(mode), func(t *testing.T) {
			dir := t.TempDir()
			editor := writeScript(t, dir, "my editor", `test "$1" = '--wait' || exit 2
test "$(cat "$2")" = '{"old":true}' || exit 3
printf '{"edited":true}' > "$2"
`)
			log := filepath.Join(dir, "commands")
			t.Setenv("COMMAND_LOG", log)
			writeScript(t, dir, "tmux", `printf '%s\n' "$*" >> "$COMMAND_LOG"
if [ "$1" = split-window ]; then
  test "$2" = '-h' || exit 4
  sh -c "$6" || exit 5
  printf '%%42\n'
fi
`)
			writeScript(t, dir, "herdr", `printf '%s\n' "$*" >> "$COMMAND_LOG"
case "$2" in
  split) printf '{"result":{"pane":{"pane_id":"w1:p42"}}}' ;;
  run) sh -c "$4" ;;
esac
`)
			t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
			session, err := New(context.Background(), mode, quote(editor)+" --wait", "../my request.curl", `{"old":true}`)
			if err != nil {
				t.Fatal(err)
			}
			if filepath.Base(session.Path) != "my request.json" {
				t.Fatal(session.Path)
			}
			info, err := os.Stat(session.Path)
			if err != nil || info.Mode().Perm() != 0600 {
				t.Fatalf("file permissions: %v, %v", info, err)
			}
			if err := session.Run(); err != nil {
				t.Fatal(err)
			}
			if session.Body != `{"edited":true}` {
				t.Fatal(session.Body)
			}
			if _, err := os.Stat(session.dir); !os.IsNotExist(err) {
				t.Fatal("temporary directory not removed")
			}
			if mode == config.SidePanelHerdr {
				data, _ := os.ReadFile(log)
				if !strings.Contains(string(data), "--direction right --focus") || !strings.Contains(string(data), "close w1:p42") {
					t.Fatal(string(data))
				}
			}
		})
	}
}

func TestEditorFailuresAndEmptyBody(t *testing.T) {
	for _, tc := range []struct {
		name, script string
		wantErr      bool
	}{
		{"invalid JSON", `printf 'invalid' > "$1"`, true},
		{"editor failure", "exit 9", true},
		{"deleted file", `rm "$1"`, true},
		{"empty body", `printf '' > "$1"`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			editor := writeScript(t, t.TempDir(), "editor", tc.script)
			s, err := New(context.Background(), config.GUI, quote(editor), "test.curl", `{}`)
			if err != nil {
				t.Fatal(err)
			}
			if err := s.Run(); (err != nil) != tc.wantErr {
				t.Fatalf("Run() = %v", err)
			}
			if _, err := os.Stat(s.dir); !os.IsNotExist(err) {
				t.Fatal("temporary files leaked")
			}
		})
	}
	for _, editor := range []string{"", "'", "$(echo editor)"} {
		if _, err := New(context.Background(), config.GUI, editor, "test", ""); err == nil {
			t.Errorf("accepted EDITOR %q", editor)
		}
	}
}

func TestPaneWaitCanBeCanceled(t *testing.T) {
	dir := t.TempDir()
	writeScript(t, dir, "tmux", "printf '%%42\\n'\n")
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	s, err := New(ctx, config.SidePanelTmux, "vi", "test", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Run(); err == nil {
		t.Fatal("expected cancellation")
	}
	if _, err := os.Stat(s.dir); !os.IsNotExist(err) {
		t.Fatal("temporary files leaked")
	}
}

func TestManuallyClosedPane(t *testing.T) {
	dir := t.TempDir()
	writeScript(t, dir, "tmux", "if [ \"$1\" = split-window ]; then printf '%%42\\n'; else exit 1; fi\n")
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	s, err := New(ctx, config.SidePanelTmux, "vi", "test", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Run(); err == nil || !strings.Contains(err.Error(), "pane closed") {
		t.Fatalf("Run() = %v", err)
	}
	if _, err := os.Stat(s.dir); !os.IsNotExist(err) {
		t.Fatal("temporary files leaked")
	}
}
