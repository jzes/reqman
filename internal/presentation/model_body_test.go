package presentation

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/cursor"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/jzes/reqman/internal/domain/request"
	"github.com/jzes/reqman/internal/presentation/jsoneditor"
)

func TestBodyEditorSyncsBodyToSelectedRequest(t *testing.T) {
	screen := newScreen([]request.Request{{Body: ""}})
	screen.focusedPanel = focusedPanelBody

	screen = enterBodyTextInsertMode(t, screen)
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("abc")})

	if got := screen.body.Value(); got != "abc" {
		t.Fatalf("body textarea = %q, want abc", got)
	}
	if got := screen.requests[0].Body; got != "abc" {
		t.Fatalf("request body = %q, want abc", got)
	}
}

func TestBodyEditorAutocompletesBraces(t *testing.T) {
	screen := newScreen([]request.Request{{Body: ""}})
	screen.focusedPanel = focusedPanelBody

	screen = enterBodyTextInsertMode(t, screen)
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'{'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(`"a":1`)})

	if got := screen.requests[0].Body; got != `{"a":1}` {
		t.Fatalf("request body = %q, want {\"a\":1}", got)
	}
}

func TestBodyEditorAutocompletesBrackets(t *testing.T) {
	screen := newScreen([]request.Request{{Body: ""}})
	screen.focusedPanel = focusedPanelBody

	screen = enterBodyTextInsertMode(t, screen)
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})

	if got := screen.requests[0].Body; got != `[1]` {
		t.Fatalf("request body = %q, want [1]", got)
	}
}

func TestBodyEditorFormatsJSON(t *testing.T) {
	screen := newScreen([]request.Request{{Body: `{"a":[1,true]}`}})
	screen.focusedPanel = focusedPanelBody

	screen = enterBodyTextInsertMode(t, screen)
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyCtrlF})

	want := "{\n  \"a\": [\n    1,\n    true\n  ]\n}"
	if got := screen.body.Value(); got != want {
		t.Fatalf("body textarea = %q, want %q", got, want)
	}
	if got := screen.requests[0].Body; got != want {
		t.Fatalf("request body = %q, want %q", got, want)
	}
}

func TestBodyEditorFormatInvalidJSONDoesNotChangeBody(t *testing.T) {
	body := `{"bad":`
	screen := newScreen([]request.Request{{Body: body}})
	screen.focusedPanel = focusedPanelBody

	screen = enterBodyTextInsertMode(t, screen)
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyCtrlF})

	if got := screen.body.Value(); got != body {
		t.Fatalf("body textarea = %q, want %q", got, body)
	}
	if got := screen.requests[0].Body; got != body {
		t.Fatalf("request body = %q, want %q", got, body)
	}
}

func TestBodyNormalModeIEntersNavigationMode(t *testing.T) {
	screen := newScreen([]request.Request{{Body: ""}})
	screen.focusedPanel = focusedPanelBody

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	if !screen.insertMode {
		t.Fatal("insert mode is false, want true")
	}
	if screen.body.Mode() != jsoneditor.NavigateMode {
		t.Fatalf("body mode = %v, want navigate", screen.body.Mode())
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if got := screen.requests[0].Body; got != "" {
		t.Fatalf("request body = %q, want empty", got)
	}
}

func TestBodyNavigationModeIEntersTextInsertMode(t *testing.T) {
	screen := newScreen([]request.Request{{Body: ""}})
	screen.focusedPanel = focusedPanelBody

	screen = enterBodyTextInsertMode(t, screen)
	if screen.body.Mode() != jsoneditor.InsertMode {
		t.Fatalf("body mode = %v, want insert", screen.body.Mode())
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if got := screen.requests[0].Body; got != "x" {
		t.Fatalf("request body = %q, want x", got)
	}
}

func TestBodyNavigationModeAEntersTextInsertModeAfterCurrentCharacter(t *testing.T) {
	screen := newScreen([]request.Request{{Body: "abc\ndef"}})
	screen.focusedPanel = focusedPanelBody

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'X'}})

	if got := screen.requests[0].Body; got != "aXbc\ndef" {
		t.Fatalf("request body after a = %q, want %q", got, "aXbc\ndef")
	}
}

func TestBodyNavigationModeShiftAEntersTextInsertModeAtEndOfCurrentLine(t *testing.T) {
	screen := newScreen([]request.Request{{Body: "abc\ndef"}})
	screen.focusedPanel = focusedPanelBody

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'X'}})

	if got := screen.requests[0].Body; got != "abcX\ndef" {
		t.Fatalf("request body after A = %q, want %q", got, "abcX\ndef")
	}
}

func TestBodyEditorFormatsJSONWhenLeavingTextInsertMode(t *testing.T) {
	screen := newScreen([]request.Request{{Body: `{"a":[1,true]}`}})
	screen.focusedPanel = focusedPanelBody

	screen = enterBodyTextInsertMode(t, screen)
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEsc})

	want := "{\n  \"a\": [\n    1,\n    true\n  ]\n}"
	if got := screen.body.Value(); got != want {
		t.Fatalf("body textarea = %q, want %q", got, want)
	}
	if got := screen.requests[0].Body; got != want {
		t.Fatalf("request body = %q, want %q", got, want)
	}
}

func TestBodyEscTransitionsFromInsertToNavigateToPanel(t *testing.T) {
	screen := newScreen([]request.Request{{Body: ""}})
	screen.focusedPanel = focusedPanelBody
	screen = enterBodyTextInsertMode(t, screen)

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEsc})
	if screen.body.Mode() != jsoneditor.NavigateMode {
		t.Fatalf("body mode = %v, want navigate", screen.body.Mode())
	}
	if !screen.insertMode {
		t.Fatal("insert mode is false, want true after returning to body navigation")
	}
	if got := screen.body.CursorMode(); got != cursor.CursorStatic {
		t.Fatalf("cursor mode after insert esc = %v, want static", got)
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEsc})
	if screen.body.Mode() != jsoneditor.PanelMode {
		t.Fatalf("body mode = %v, want panel", screen.body.Mode())
	}
	if screen.insertMode {
		t.Fatal("insert mode is true, want false after returning to panel mode")
	}
	if got := screen.body.CursorMode(); got != cursor.CursorStatic {
		t.Fatalf("cursor mode after navigation esc = %v, want static", got)
	}
}

func TestBodyCursorModeTracksInsertAndNavigationModes(t *testing.T) {
	screen := newScreen([]request.Request{{Body: ""}})
	screen.focusedPanel = focusedPanelBody

	if got := screen.body.CursorMode(); got != cursor.CursorStatic {
		t.Fatalf("initial cursor mode = %v, want static", got)
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	if got := screen.body.CursorMode(); got != cursor.CursorStatic {
		t.Fatalf("navigation cursor mode = %v, want static", got)
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	if got := screen.body.CursorMode(); got != cursor.CursorBlink {
		t.Fatalf("insert cursor mode = %v, want blink", got)
	}
}

func TestBodyViewRendersBarCursorOnlyInTextInsertMode(t *testing.T) {
	screen := newScreen([]request.Request{{Body: "abc"}})
	screen.focusedPanel = focusedPanelBody

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	view := screen.body.View(20, 3, true)
	if strings.Contains(view, "|") {
		t.Fatalf("navigation body view contains bar cursor: %q", view)
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	view = screen.body.View(20, 3, true)
	if !strings.Contains(view, "|") {
		t.Fatalf("insert body view does not contain bar cursor: %q", view)
	}
	if !strings.Contains(view, "|abc") {
		t.Fatalf("insert body view overwrote character under cursor: %q", view)
	}
}

func TestBodyNavigationModeMotionsStayInBodyPanel(t *testing.T) {
	body := "one\ntwo\nthree"
	screen := newScreen([]request.Request{{Body: body}})
	screen.focusedPanel = focusedPanelBody

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	for _, motion := range []rune{'h', 'j', 'k', 'l'} {
		screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{motion}})
		if screen.focusedPanel != focusedPanelBody {
			t.Fatalf("focused panel after %q = %v, want body", motion, screen.focusedPanel)
		}
	}
	if got := screen.requests[0].Body; got != body {
		t.Fatalf("request body = %q, want %q", got, body)
	}
}

func TestBodyNormalModePanelMotionsAreNotCapturedByText(t *testing.T) {
	body := "one\ntwo\nthree"
	screen := newScreen([]request.Request{{Body: body}})
	screen.focusedPanel = focusedPanelBody

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if screen.focusedPanel != focusedPanelHeaders {
		t.Fatalf("focused panel = %v, want headers", screen.focusedPanel)
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if screen.focusedPanel != focusedPanelBody {
		t.Fatalf("focused panel = %v, want body", screen.focusedPanel)
	}

	if got := screen.body.Value(); got != body {
		t.Fatalf("body textarea = %q, want %q", got, body)
	}
	if got := screen.requests[0].Body; got != body {
		t.Fatalf("request body = %q, want %q", got, body)
	}
}

func TestBodyNavigationWordMotionsMoveCursor(t *testing.T) {
	screen := newScreen([]request.Request{{Body: "one two three"}})
	screen.focusedPanel = focusedPanelBody

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'X'}})
	if got := screen.requests[0].Body; got != "one Xtwo three" {
		t.Fatalf("request body after w = %q, want %q", got, "one Xtwo three")
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEsc})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Y'}})
	if got := screen.requests[0].Body; got != "one XtwoY three" {
		t.Fatalf("request body after e = %q, want %q", got, "one XtwoY three")
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEsc})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Z'}})
	if got := screen.requests[0].Body; got != "one ZXtwoY three" {
		t.Fatalf("request body after b = %q, want %q", got, "one ZXtwoY three")
	}
}

func TestBodyNavigationModeExitFormatsJSONWithJQ(t *testing.T) {
	screen := newScreen([]request.Request{{Body: `{"a":[1,true]}`}})
	screen.focusedPanel = focusedPanelBody

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEsc})

	want := "{\n  \"a\": [\n    1,\n    true\n  ]\n}"
	if got := screen.body.Value(); got != want {
		t.Fatalf("body textarea = %q, want %q", got, want)
	}
	if got := screen.requests[0].Body; got != want {
		t.Fatalf("request body = %q, want %q", got, want)
	}
}
