package presentation

import (
	"context"
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jzes/reqman/internal/config"
	"github.com/jzes/reqman/internal/domain/request"
)

func TestBodyExternalEditorKeyAndCompletion(t *testing.T) {
	for _, mode := range []config.EditorMode{config.InLine, config.GUI, config.SidePanelTmux, config.SidePanelHerdr} {
		t.Run(string(mode), func(t *testing.T) {
			t.Setenv("TMPDIR", t.TempDir())
			t.Setenv("EDITOR", "true")
			scr := newScreen([]request.Request{{Name: "one.curl", Body: `{}`}, {Name: "two.curl", Body: `[]`}})
			scr.config.EditorMode = mode
			scr.focusedPanel = focusedPanelBody
			scr, cmd := updateScreen(t, scr, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
			if cmd == nil || scr.editorCancel == nil || scr.insertMode {
				t.Fatal("external edit did not start in normal mode")
			}
			// Starting another edit must not replace the active session.
			scr, second := updateScreen(t, scr, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
			if second != nil {
				t.Fatal("duplicate editor started")
			}
			// Adapter tests exercise the other modes without opening real panes.
			if mode == config.GUI {
				scr, _ = updateScreen(t, scr, cmd())
			} else {
				scr.cancelBodyEdit()
			}
			scr.selectedRequestIndex = 1
			scr.showSelectedRequestBody()
			scr, _ = updateScreen(t, scr, bodyEditedMsg{index: 0, body: `{"saved":true}`})
			if scr.requests[0].Body != `{"saved":true}` || scr.body.Value() != `[]` || scr.editorCancel != nil {
				t.Fatal("edited wrong request")
			}
		})
	}
}

func TestBodyExternalEditorErrorPreservesBody(t *testing.T) {
	scr := newScreen([]request.Request{{Body: `{}`}})
	scr, _ = updateScreen(t, scr, bodyEditedMsg{index: 0, err: errors.New("editor failed")})
	if scr.requests[0].Body != `{}` || scr.body.Value() != `{}` || scr.responseError == "" {
		t.Fatal("editor failure not handled")
	}
	scr, _ = updateScreen(t, scr, bodyEditedMsg{index: 0, body: `{"saved":true}`})
	if scr.body.Value() != `{"saved":true}` {
		t.Fatal("body panel not refreshed")
	}
}

func TestBodyExternalEditorMissingEditorAndEmptyList(t *testing.T) {
	t.Setenv("EDITOR", "")
	scr := newScreen([]request.Request{{Body: `{}`}})
	scr.config.EditorMode = config.GUI
	scr.focusedPanel = focusedPanelBody
	scr, cmd := updateScreen(t, scr, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	if cmd != nil || scr.responseError == "" || scr.insertMode {
		t.Fatal("missing EDITOR not handled")
	}
	scr = newScreen(nil)
	scr.config.EditorMode = config.GUI
	scr.focusedPanel = focusedPanelBody
	_, cmd = updateScreen(t, scr, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	if cmd != nil {
		t.Fatal("started editor without request")
	}
}

func TestQuitCancelsBodyEditor(t *testing.T) {
	for _, quit := range []string{"ctrl+c", ":q", ":wq"} {
		t.Run(quit, func(t *testing.T) {
			scr := newScreenWithDeps([]request.Request{{Name: "test.curl"}}, &fakeWriter{}, nil)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			scr.editorCancel = cancel
			if quit == "ctrl+c" {
				scr, _ = updateScreen(t, scr, tea.KeyMsg{Type: tea.KeyCtrlC})
			} else {
				for _, key := range quit {
					scr, _ = updateScreen(t, scr, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{key}})
				}
				scr, _ = updateScreen(t, scr, tea.KeyMsg{Type: tea.KeyEnter})
			}
			if ctx.Err() == nil || scr.editorCancel != nil {
				t.Fatal("editor not canceled on quit")
			}
		})
	}
}
