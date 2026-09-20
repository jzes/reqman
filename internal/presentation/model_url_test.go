package presentation

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/cursor"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/jzes/reqman/internal/domain/request"
)

func TestURLEditorSyncsURLToSelectedRequest(t *testing.T) {
	screen := newScreen([]request.Request{{}})
	screen.focusedPanel = focusedPanelURL

	screen = enterURLTextInsertMode(t, screen)
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("localhost/books")})

	if got := screen.url.Value(); got != "localhost/books" {
		t.Fatalf("URL input = %q, want localhost/books", got)
	}
	if got := screen.requests[0].URL.String(); got != "localhost/books" {
		t.Fatalf("request URL = %q, want localhost/books", got)
	}
}

func TestURLEditorShowsInvalidURLError(t *testing.T) {
	parsedURL, err := request.NewURL("http://localhost/books")
	if err != nil {
		t.Fatalf("NewURL() error = %v", err)
	}
	screen := newScreen([]request.Request{{URL: parsedURL}})
	screen.focusedPanel = focusedPanelURL

	screen.url.SetValue("http://[::1")
	screen.syncURLToSelectedRequest()

	if got := screen.requests[0].URL.String(); got != "http://localhost/books" {
		t.Fatalf("request URL = %q, want previous URL", got)
	}
	if got := screen.renderResponse(); !strings.Contains(got, "Invalid URL:") {
		t.Fatalf("response view = %q, want invalid URL error", got)
	}
}

func TestURLNormalModeIEntersNavigationMode(t *testing.T) {
	screen := newScreen([]request.Request{{}})
	screen.focusedPanel = focusedPanelURL

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	if !screen.insertMode {
		t.Fatal("insert mode is false, want true")
	}
	if screen.url.Mode() != urlPanelModeNavigate {
		t.Fatalf("URL mode = %v, want navigate", screen.url.Mode())
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if got := screen.url.Value(); got != "" {
		t.Fatalf("URL input = %q, want empty", got)
	}
}

func TestURLNavigationModeIEntersTextInsertMode(t *testing.T) {
	screen := newScreen([]request.Request{{}})
	screen.focusedPanel = focusedPanelURL

	screen = enterURLTextInsertMode(t, screen)
	if screen.url.Mode() != urlPanelModeInsert {
		t.Fatalf("URL mode = %v, want insert", screen.url.Mode())
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if got := screen.url.Value(); got != "x" {
		t.Fatalf("URL input = %q, want x", got)
	}
}

func TestURLNavigationModeAEntersTextInsertModeAfterCurrentCharacter(t *testing.T) {
	parsedURL, err := request.NewURL("abc/def")
	if err != nil {
		t.Fatalf("NewURL() error = %v", err)
	}
	screen := newScreen([]request.Request{{URL: parsedURL}})
	screen.focusedPanel = focusedPanelURL

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'X'}})

	if got := screen.url.Value(); got != "aXbc/def" {
		t.Fatalf("URL after a = %q, want %q", got, "aXbc/def")
	}
}

func TestURLNavigationModeShiftAEntersTextInsertModeAtEndOfLine(t *testing.T) {
	parsedURL, err := request.NewURL("abc/def")
	if err != nil {
		t.Fatalf("NewURL() error = %v", err)
	}
	screen := newScreen([]request.Request{{URL: parsedURL}})
	screen.focusedPanel = focusedPanelURL

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'X'}})

	if got := screen.url.Value(); got != "abc/defX" {
		t.Fatalf("URL after A = %q, want %q", got, "abc/defX")
	}
}

func TestURLEscTransitionsFromInsertToNavigateToPanel(t *testing.T) {
	screen := newScreen([]request.Request{{}})
	screen.focusedPanel = focusedPanelURL
	screen = enterURLTextInsertMode(t, screen)

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEsc})
	if screen.url.Mode() != urlPanelModeNavigate {
		t.Fatalf("URL mode = %v, want navigate", screen.url.Mode())
	}
	if !screen.insertMode {
		t.Fatal("insert mode is false, want true after returning to URL navigation")
	}
	if got := screen.url.CursorMode(); got != cursor.CursorStatic {
		t.Fatalf("cursor mode after insert esc = %v, want static", got)
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEsc})
	if screen.url.Mode() != urlPanelModePanel {
		t.Fatalf("URL mode = %v, want panel", screen.url.Mode())
	}
	if screen.insertMode {
		t.Fatal("insert mode is true, want false after returning to panel mode")
	}
	if got := screen.url.CursorMode(); got != cursor.CursorStatic {
		t.Fatalf("cursor mode after navigation esc = %v, want static", got)
	}
}

func TestURLCursorModeTracksInsertAndNavigationModes(t *testing.T) {
	screen := newScreen([]request.Request{{}})
	screen.focusedPanel = focusedPanelURL

	if got := screen.url.CursorMode(); got != cursor.CursorStatic {
		t.Fatalf("initial cursor mode = %v, want static", got)
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	if got := screen.url.CursorMode(); got != cursor.CursorStatic {
		t.Fatalf("navigation cursor mode = %v, want static", got)
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	if got := screen.url.CursorMode(); got != cursor.CursorBlink {
		t.Fatalf("insert cursor mode = %v, want blink", got)
	}
}

func TestURLViewRendersBarCursorOnlyInTextInsertMode(t *testing.T) {
	parsedURL, err := request.NewURL("abc")
	if err != nil {
		t.Fatalf("NewURL() error = %v", err)
	}
	screen := newScreen([]request.Request{{URL: parsedURL}})
	screen.focusedPanel = focusedPanelURL

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	view := screen.url.View(20, true)
	if strings.Contains(view, "|") {
		t.Fatalf("navigation URL view contains bar cursor: %q", view)
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	view = screen.url.View(20, true)
	if !strings.Contains(view, "|") {
		t.Fatalf("insert URL view does not contain bar cursor: %q", view)
	}
	if !strings.Contains(view, "|abc") {
		t.Fatalf("insert URL view overwrote character under cursor: %q", view)
	}
}

func TestURLViewEmptyInsertModeDoesNotLeakANSISequence(t *testing.T) {
	screen := newScreen([]request.Request{{}})
	screen.focusedPanel = focusedPanelURL
	screen = enterURLTextInsertMode(t, screen)

	view := screen.url.View(20, true)
	if strings.Contains(view, "|[38;5;240m") {
		t.Fatalf("empty insert URL view leaks ANSI sequence: %q", view)
	}
}

func TestInsertAtURLColumnSkipsANSISequences(t *testing.T) {
	line := "\x1b[38;5;240mhttp://localhost\x1b[0m"
	want := "\x1b[38;5;240m|http://localhost\x1b[0m"
	if got := insertAtURLColumn(line, 0, "|"); got != want {
		t.Fatalf("insertAtURLColumn() = %q, want %q", got, want)
	}
}

func TestURLNavigationWordMotionsMoveCursor(t *testing.T) {
	parsedURL, err := request.NewURL("one/two/three")
	if err != nil {
		t.Fatalf("NewURL() error = %v", err)
	}

	screen := newScreen([]request.Request{{URL: parsedURL}})
	screen.focusedPanel = focusedPanelURL

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'X'}})
	if got := screen.url.Value(); got != "one/Xtwo/three" {
		t.Fatalf("URL after w = %q, want %q", got, "one/Xtwo/three")
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEsc})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Y'}})
	if got := screen.url.Value(); got != "one/XtwoY/three" {
		t.Fatalf("URL after e = %q, want %q", got, "one/XtwoY/three")
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEsc})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Z'}})
	if got := screen.url.Value(); got != "one/ZXtwoY/three" {
		t.Fatalf("URL after b = %q, want %q", got, "one/ZXtwoY/three")
	}
}
