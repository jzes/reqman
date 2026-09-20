package jsoneditor

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/cursor"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TestFormatJSONFormatsValidJSON(t *testing.T) {
	panel := NewPanel()
	panel.SetValue(`{"a":[1,true]}`)

	if ok := panel.FormatJSON(); !ok {
		t.Fatal("FormatJSON() = false, want true")
	}

	want := "{\n  \"a\": [\n    1,\n    true\n  ]\n}"
	if got := panel.Value(); got != want {
		t.Fatalf("Value() = %q, want %q", got, want)
	}
}

func TestFormatJSONInvalidJSONDoesNotChangeValue(t *testing.T) {
	panel := NewPanel()
	panel.SetValue(`{"bad":`)

	if ok := panel.FormatJSON(); ok {
		t.Fatal("FormatJSON() = true, want false")
	}

	want := `{"bad":`
	if got := panel.Value(); got != want {
		t.Fatalf("Value() = %q, want %q", got, want)
	}
}

func TestPersistChangesFormatsJSON(t *testing.T) {
	panel := NewPanel()
	panel.SetValue(`{"a":[1,true]}`)

	panel.PersistChanges()

	want := "{\n  \"a\": [\n    1,\n    true\n  ]\n}"
	if got := panel.Value(); got != want {
		t.Fatalf("Value() = %q, want %q", got, want)
	}
}

func TestPersistChangesInvalidJSONDoesNotChangeValue(t *testing.T) {
	panel := NewPanel()
	panel.SetValue(`{"bad":`)

	panel.PersistChanges()

	want := `{"bad":`
	if got := panel.Value(); got != want {
		t.Fatalf("Value() = %q, want %q", got, want)
	}
}

func TestModeAndCursorTransitions(t *testing.T) {
	panel := NewPanel()
	if panel.Mode() != PanelMode {
		t.Fatalf("initial mode = %v, want panel", panel.Mode())
	}
	if panel.CursorMode() != cursor.CursorStatic {
		t.Fatalf("initial cursor = %v, want static", panel.CursorMode())
	}

	panel.SetMode(InsertMode)
	if panel.CursorMode() != cursor.CursorBlink {
		t.Fatalf("insert cursor = %v, want blink", panel.CursorMode())
	}
	panel.SetMode(NavigateMode)
	if panel.CursorMode() != cursor.CursorStatic {
		t.Fatalf("navigate cursor = %v, want static", panel.CursorMode())
	}
}

func TestEnterNavigationAndHandleEscape(t *testing.T) {
	panel := NewPanel()
	panel.SetValue(`{"a":1}`)

	panel.EnterNavigationMode()
	if panel.Mode() != NavigateMode {
		t.Fatalf("mode after navigation = %v, want navigate", panel.Mode())
	}

	panel.UpdateNavigationMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	if panel.Mode() != InsertMode {
		t.Fatalf("mode after i = %v, want insert", panel.Mode())
	}
	if got := panel.HandleEscape(); got != EscapeToNavigation {
		t.Fatalf("escape from insert = %v, want navigation", got)
	}
	if panel.Mode() != NavigateMode {
		t.Fatalf("mode after insert escape = %v, want navigate", panel.Mode())
	}
	if got := panel.Value(); got != "{\n  \"a\": 1\n}" {
		t.Fatalf("value after escape = %q, want formatted JSON", got)
	}
	if got := panel.HandleEscape(); got != EscapeToPanel {
		t.Fatalf("escape from navigate = %v, want panel", got)
	}
	if got := panel.HandleEscape(); got != EscapeIgnored {
		t.Fatalf("escape from panel = %v, want ignored", got)
	}
}

func TestUpdateEditorAutocompletesAndFormatsJSON(t *testing.T) {
	panel := NewPanel()
	panel.EnterNavigationMode()
	panel.UpdateNavigationMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

	_, changed := panel.UpdateEditor(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'{'}})
	if !changed {
		t.Fatal("changed after brace = false, want true")
	}
	panel.UpdateEditor(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(`"a":1`)})
	if got := panel.Value(); got != `{"a":1}` {
		t.Fatalf("value after brace autocomplete = %q, want object", got)
	}

	_, changed = panel.UpdateEditor(tea.KeyMsg{Type: tea.KeyCtrlF})
	if !changed {
		t.Fatal("changed after ctrl+f = false, want true")
	}
	if got := panel.Value(); got != "{\n  \"a\": 1\n}" {
		t.Fatalf("formatted value = %q, want formatted JSON", got)
	}
}

func TestUpdateEditorAutocompletesBrackets(t *testing.T) {
	panel := NewPanel()
	panel.EnterNavigationMode()
	panel.UpdateNavigationMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})

	panel.UpdateEditor(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	panel.UpdateEditor(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	if got := panel.Value(); got != `[1]` {
		t.Fatalf("value after bracket autocomplete = %q, want [1]", got)
	}
}

func TestUpdateNavigationModeMotionsAndInsertCommands(t *testing.T) {
	panel := NewPanel()
	panel.SetValue("one two three")
	panel.EnterNavigationMode()

	tests := []rune{'l', 'h', 'w', 'e', 'b', 'j', 'k'}
	for _, key := range tests {
		_, handled := panel.UpdateNavigationMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{key}})
		if !handled {
			t.Fatalf("navigation key %q handled = false, want true", key)
		}
	}
	_, handled := panel.UpdateNavigationMode(tea.KeyMsg{Type: tea.KeyEnter})
	if handled {
		t.Fatal("enter handled = true, want false")
	}

	panel.UpdateNavigationMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})
	panel.UpdateEditor(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'!'}})
	if got := panel.Value(); got != "one two three!" {
		t.Fatalf("value after A insert = %q, want appended", got)
	}
}

func TestViewRendersBarCursorOnlyInInsertMode(t *testing.T) {
	panel := NewPanel()
	panel.SetValue("abc")
	panel.SetMode(NavigateMode)

	if view := panel.View(20, 3, true); strings.Contains(view, "|") {
		t.Fatalf("navigate view contains bar cursor: %q", view)
	}
	panel.SetMode(InsertMode)
	if view := panel.View(20, 3, true); !strings.Contains(view, "|") {
		t.Fatalf("insert view does not contain bar cursor: %q", view)
	}
}

func TestStyleUsesModeSpecificFocusedBorders(t *testing.T) {
	panel := NewPanel()
	base := lipgloss.NewStyle()

	if got := panel.Style(base, false).GetBorderTopForeground(); got != base.GetBorderTopForeground() {
		t.Fatalf("unfocused border color = %v, want base", got)
	}
	for _, mode := range []Mode{PanelMode, NavigateMode, InsertMode} {
		panel.SetMode(mode)
		if got := panel.Style(base, true).GetBorderTopForeground(); got == nil {
			t.Fatalf("focused style for mode %v has empty border color", mode)
		}
	}
}

func TestInsertAtVisibleColumnSkipsANSISequences(t *testing.T) {
	line := "\x1b[38;5;240mabc\x1b[0m"
	want := "\x1b[38;5;240m|abc\x1b[0m"
	if got := insertAtVisibleColumn(line, 0, "|"); got != want {
		t.Fatalf("insertAtVisibleColumn() = %q, want %q", got, want)
	}
	if got := insertAtVisibleColumn("abc", 10, "|"); got != "abc|" {
		t.Fatalf("insert past end = %q, want abc|", got)
	}
}

func TestWrappedLineCount(t *testing.T) {
	if got := wrappedLineCount("", 10); got != 1 {
		t.Fatalf("empty wrapped lines = %d, want 1", got)
	}
	if got := wrappedLineCount("abcdef", 3); got != 2 {
		t.Fatalf("wrapped lines = %d, want 2", got)
	}
	if got := wrappedLineCount("abcdef", 0); got != 1 {
		t.Fatalf("zero width wrapped lines = %d, want 1", got)
	}
}
