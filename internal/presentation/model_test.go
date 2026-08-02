package presentation

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/cursor"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jzes/reqman/internal/domain/request"
	presentationbody "github.com/jzes/reqman/internal/presentation/body"
)

func TestViewRendersBodyAndResponsePanels(t *testing.T) {
	screen := newScreen([]request.Request{{Name: "req.curl", Body: "{}"}})
	screen.width = 100
	screen.height = 30

	view := screen.View()
	if !strings.Contains(view, "Body") {
		t.Fatal("view does not contain Body panel title")
	}
	if !strings.Contains(view, "Response") {
		t.Fatal("view does not contain Response panel title")
	}
}

func TestViewRendersRequestNamesWithoutExtension(t *testing.T) {
	screen := newScreen([]request.Request{{Name: "books.curl"}})

	view := screen.View()
	if !strings.Contains(view, "books") {
		t.Fatalf("view does not contain request name without extension: %q", view)
	}
	if strings.Contains(view, "books.curl") {
		t.Fatalf("view contains request extension: %q", view)
	}
}

func TestScreenLoadsOnlySelectedRequest(t *testing.T) {
	loader := &countingLoader{requests: map[string]request.Request{
		"one.curl": {Name: "one.curl", Body: "one"},
		"two.curl": {Name: "two.curl", Body: "two"},
	}}
	screen := NewScreen([]string{"one.curl", "two.curl"}, nil, nil, loader)

	if got := loader.calls["one.curl"]; got != 1 {
		t.Fatalf("one.curl load calls = %d, want 1", got)
	}
	if got := loader.calls["two.curl"]; got != 0 {
		t.Fatalf("two.curl load calls = %d, want 0", got)
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if got := loader.calls["two.curl"]; got != 1 {
		t.Fatalf("two.curl load calls after selection = %d, want 1", got)
	}
}

func TestViewRendersTopBar(t *testing.T) {
	screen := newScreen([]request.Request{{Name: "req.curl"}})
	screen.width = 80

	view := ansiSequencePattern.ReplaceAllString(screen.View(), "")
	firstLine := strings.Split(view, "\n")[0]
	if !strings.HasPrefix(firstLine, "Req-Man") {
		t.Fatalf("top bar = %q, want Req-Man on the left", firstLine)
	}
	if !strings.HasSuffix(firstLine, ": to open commands") {
		t.Fatalf("top bar = %q, want command hint on the right", firstLine)
	}
}

func TestRequestsPanelScrollsToSelectedRequest(t *testing.T) {
	requests := make([]request.Request, 10)
	for i := range requests {
		requests[i] = request.Request{Name: fmt.Sprintf("req-%d.curl", i+1)}
	}
	screen := newScreen(requests)
	screen.height = 6
	screen.focusedPanel = focusedPanelRequests

	for range 5 {
		screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	}

	if got := screen.requestScrollOffset; got != 3 {
		t.Fatalf("request scroll offset = %d, want 3", got)
	}
	view := ansiSequencePattern.ReplaceAllString(screen.View(), "")
	if strings.Contains(view, "1. req-1") {
		t.Fatalf("view contains scrolled-out first request: %q", view)
	}
	if !strings.Contains(view, "❯ 6. req-6") {
		t.Fatalf("view does not contain selected request in visible window: %q", view)
	}
}

func TestResponsePanelVerticalNavigation(t *testing.T) {
	screen := newScreen([]request.Request{{Name: "req.curl"}})
	screen.focusedPanel = focusedPanelHeaders

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if screen.focusedPanel != focusedPanelBody {
		t.Fatalf("focused panel after headers j = %v, want body", screen.focusedPanel)
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if screen.focusedPanel != focusedPanelResponse {
		t.Fatalf("focused panel after body j = %v, want response", screen.focusedPanel)
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if screen.focusedPanel != focusedPanelBody {
		t.Fatalf("focused panel after response k = %v, want body", screen.focusedPanel)
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if screen.focusedPanel != focusedPanelHeaders {
		t.Fatalf("focused panel after body k = %v, want headers", screen.focusedPanel)
	}
}

func TestResponsePanelLeftNavigationFocusesRequests(t *testing.T) {
	screen := newScreen([]request.Request{{Name: "req.curl"}})
	screen.focusedPanel = focusedPanelResponse

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	if screen.focusedPanel != focusedPanelRequests {
		t.Fatalf("focused panel = %v, want requests", screen.focusedPanel)
	}
}

func TestResponseRendersTabs(t *testing.T) {
	screen := newScreen([]request.Request{{Name: "req.curl"}})
	screen.hasResponse = true
	screen.response = request.Response{Status: "200 OK", StatusCode: 200, Duration: 123 * time.Millisecond}

	view := screen.renderResponse(80)
	for _, tab := range []string{"Stats", "Body", "Headers", "Raw"} {
		if !strings.Contains(view, tab) {
			t.Fatalf("response view does not contain %s tab: %q", tab, view)
		}
	}
	if !strings.Contains(view, "╭") || !strings.Contains(view, "╯") {
		t.Fatalf("response view does not contain bordered tabs: %q", view)
	}
	plainView := ansiSequencePattern.ReplaceAllString(view, "")
	plainLines := strings.Split(plainView, "\n")
	if len(plainLines) < 3 {
		t.Fatalf("response view does not contain tab content border: %q", view)
	}
	if !strings.HasPrefix(plainLines[2], "│") || !strings.Contains(plainLines[2], "╰") {
		t.Fatalf("response content border is not connected to active tab: %q", view)
	}
	if !strings.Contains(view, "Duration: 123ms") {
		t.Fatalf("response view does not contain stats content: %q", view)
	}
}

func TestResponseBodyTabFormatsJSON(t *testing.T) {
	screen := newScreen([]request.Request{{Name: "req.curl"}})
	screen.hasResponse = true
	screen.selectedResponseTab = responseTabBody
	screen.response = request.Response{Body: `{"name":"x","items":[1,2]}`}

	view := screen.renderResponse(80)
	if !strings.Contains(view, "  \"name\": \"x\"") {
		t.Fatalf("response body tab does not contain formatted JSON object: %q", view)
	}
	if !strings.Contains(view, "  \"items\": [") {
		t.Fatalf("response body tab does not contain formatted JSON array: %q", view)
	}
}

func TestResponseHeadersTabRendersSortedTable(t *testing.T) {
	screen := newScreen([]request.Request{{Name: "req.curl"}})
	screen.hasResponse = true
	screen.selectedResponseTab = responseTabHeaders
	screen.response = request.Response{Headers: map[string][]string{
		"X-Zeta":       {"last"},
		"Content-Type": {"application/json"},
		"Accept":       {"application/json", "text/plain"},
	}}

	view := screen.renderResponse(80)
	if !strings.Contains(view, "Header") || !strings.Contains(view, "Value") {
		t.Fatalf("response headers tab does not contain table header: %q", view)
	}
	if !strings.Contains(view, "Accept") || !strings.Contains(view, "application/json, text/plain") {
		t.Fatalf("response headers tab does not contain joined Accept header: %q", view)
	}
	acceptIndex := strings.Index(view, "Accept")
	contentTypeIndex := strings.Index(view, "Content-Type")
	zetaIndex := strings.Index(view, "X-Zeta")
	if acceptIndex == -1 || contentTypeIndex == -1 || zetaIndex == -1 {
		t.Fatalf("response headers tab is missing expected headers: %q", view)
	}
	if !(acceptIndex < contentTypeIndex && contentTypeIndex < zetaIndex) {
		t.Fatalf("response headers are not sorted: %q", view)
	}
}

func TestResponseTabNavigation(t *testing.T) {
	screen := newScreen([]request.Request{{Name: "req.curl"}})
	screen.focusedPanel = focusedPanelResponse

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}})
	if screen.selectedResponseTab != responseTabBody {
		t.Fatalf("selected response tab after ] = %v, want body", screen.selectedResponseTab)
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	if screen.selectedResponseTab != responseTabStats {
		t.Fatalf("selected response tab after [ = %v, want stats", screen.selectedResponseTab)
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	if screen.selectedResponseTab != responseTabRaw {
		t.Fatalf("selected response tab after wrapped [ = %v, want raw", screen.selectedResponseTab)
	}
}

func TestResponseInsertModeTabNavigation(t *testing.T) {
	screen := newScreen([]request.Request{{Name: "req.curl"}})
	screen.focusedPanel = focusedPanelResponse

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	if !screen.insertMode {
		t.Fatal("insert mode is false, want true")
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyTab})
	if screen.selectedResponseTab != responseTabBody {
		t.Fatalf("selected response tab after tab = %v, want body", screen.selectedResponseTab)
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyShiftTab})
	if screen.selectedResponseTab != responseTabStats {
		t.Fatalf("selected response tab after shift+tab = %v, want stats", screen.selectedResponseTab)
	}
}

func TestResponseInsertModeRendersGreenBorder(t *testing.T) {
	screen := newScreen([]request.Request{{Name: "req.curl"}})
	screen.width = 100
	screen.height = 30
	screen.focusedPanel = focusedPanelResponse
	screen.insertMode = true

	view := screen.View()
	greenBorder := lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Render("╭")
	if !strings.Contains(view, greenBorder) {
		t.Fatalf("response insert mode view does not contain green border: %q", view)
	}
}

func TestHeadersFromMapSortsRows(t *testing.T) {
	rows := headersFromMap(map[string]string{
		"X-Zeta":  "last",
		"Accept":  "application/json",
		"Content": "text/plain",
	})

	got := []string{rows[0].key, rows[1].key, rows[2].key}
	want := []string{"Accept", "Content", "X-Zeta"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("headers order = %v, want %v", got, want)
	}
}

func TestHeadersEditorCreatesFirstHeader(t *testing.T) {
	screen := newScreen([]request.Request{{Headers: map[string]string{}}})
	screen.focusedPanel = focusedPanelHeaders

	screen = enterHeaderInsertMode(t, screen)
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Accept")})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyTab})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("application/json")})

	if got := screen.requests[0].Headers["Accept"]; got != "application/json" {
		t.Fatalf("Accept header = %q, want application/json", got)
	}
}

func TestHeadersEditorEditsKeyAndValue(t *testing.T) {
	screen := newScreen([]request.Request{{Headers: map[string]string{"Accept": "text/plain"}}})
	screen.focusedPanel = focusedPanelHeaders

	screen = enterHeaderInsertMode(t, screen)
	for range len("Accept") {
		screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyBackspace})
	}
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Content-Type")})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyTab})
	for range len("text/plain") {
		screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyBackspace})
	}
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("application/json")})

	if _, ok := screen.requests[0].Headers["Accept"]; ok {
		t.Fatal("old Accept key still exists")
	}
	if got := screen.requests[0].Headers["Content-Type"]; got != "application/json" {
		t.Fatalf("Content-Type header = %q, want application/json", got)
	}
}

func TestHeadersEditorDeletesSelectedRow(t *testing.T) {
	screen := newScreen([]request.Request{{Headers: map[string]string{
		"Accept":        "application/json",
		"Authorization": "Bearer token",
	}}})
	screen.focusedPanel = focusedPanelHeaders

	screen = enterBodyTextInsertMode(t, screen)
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyDelete})

	if _, ok := screen.requests[0].Headers["Accept"]; ok {
		t.Fatal("deleted Accept key still exists")
	}
	if got := screen.requests[0].Headers["Authorization"]; got != "Bearer token" {
		t.Fatalf("Authorization header = %q, want Bearer token", got)
	}
}

func TestSyncHeadersSkipsEmptyKeysAndLastDuplicateWins(t *testing.T) {
	screen := newScreen([]request.Request{{Headers: map[string]string{}}})
	screen.headersEditor.rows = []headerRow{
		{key: "", value: "ignored"},
		{key: "Accept", value: "text/plain"},
		{key: "Accept", value: "application/json"},
	}

	screen.syncHeadersToSelectedRequest()

	if _, ok := screen.requests[0].Headers[""]; ok {
		t.Fatal("empty key was synced")
	}
	if got := screen.requests[0].Headers["Accept"]; got != "application/json" {
		t.Fatalf("Accept header = %q, want application/json", got)
	}
}

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
	if screen.body.Mode() != presentationbody.NavigateMode {
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
	if screen.body.Mode() != presentationbody.InsertMode {
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
	if screen.body.Mode() != presentationbody.NavigateMode {
		t.Fatalf("body mode = %v, want navigate", screen.body.Mode())
	}
	if !screen.insertMode {
		t.Fatal("insert mode is false, want true after returning to body navigation")
	}
	if got := screen.body.CursorMode(); got != cursor.CursorStatic {
		t.Fatalf("cursor mode after insert esc = %v, want static", got)
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEsc})
	if screen.body.Mode() != presentationbody.PanelMode {
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

func TestEscDoesNotSaveRequest(t *testing.T) {
	parsedURL, err := request.NewURL("http://localhost/books")
	if err != nil {
		t.Fatalf("NewURL() error = %v", err)
	}

	writer := &fakeWriter{}
	screen := newScreenWithDeps([]request.Request{{URL: parsedURL}}, writer, nil)
	screen.focusedPanel = focusedPanelBody
	screen = enterBodyTextInsertMode(t, screen)
	if got := screen.body.Value(); got != "" {
		t.Fatalf("body textarea = %q, want empty", got)
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEsc})
	if got := writer.calls; got != 0 {
		t.Fatalf("write calls = %d, want 0", got)
	}
	if screen.body.Mode() != presentationbody.NavigateMode {
		t.Fatalf("body mode = %v, want navigate", screen.body.Mode())
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEsc})
	if got := writer.calls; got != 0 {
		t.Fatalf("write calls = %d, want 0 after leaving body panel", got)
	}
}

func TestCommandPanelWriteCommandSavesRequest(t *testing.T) {
	parsedURL, err := request.NewURL("http://localhost/books")
	if err != nil {
		t.Fatalf("NewURL() error = %v", err)
	}

	writer := &fakeWriter{}
	screen := newScreenWithDeps([]request.Request{{Name: "req.curl", Path: "req.curl", URL: parsedURL}}, writer, nil)

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	screen, cmd := updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatalf("command after ❯w = %p, want nil", cmd)
	}
	if got := writer.calls; got != 1 {
		t.Fatalf("write calls = %d, want 1", got)
	}
	if got := writer.request.Name; got != "req.curl" {
		t.Fatalf("written request name = %q, want req.curl", got)
	}
	if screen.commandPanel.Open {
		t.Fatal("command panel is open, want closed")
	}
}

func TestCommandPanelWriteRunCommandSavesAndRunsRequest(t *testing.T) {
	parsedURL, err := request.NewURL("http://localhost/books")
	if err != nil {
		t.Fatalf("NewURL() error = %v", err)
	}

	writer := &fakeWriter{}
	doer := &fakeDoer{response: request.Response{Status: "200 OK"}}
	screen := newScreenWithDeps([]request.Request{{Name: "req.curl", Path: "req.curl", URL: parsedURL}}, writer, doer)

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("wr")})
	screen, cmd := updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("do command is nil")
	}
	if got := writer.calls; got != 1 {
		t.Fatalf("write calls = %d, want 1", got)
	}
	if got := writer.request.Name; got != "req.curl" {
		t.Fatalf("written request name = %q, want req.curl", got)
	}
	if !screen.requestInFlight {
		t.Fatal("request in flight is false, want true")
	}

	screen, _ = updateScreen(t, screen, firstCommandMessage(t, cmd))
	if got := doer.calls; got != 1 {
		t.Fatalf("do calls after ❯wr = %d, want 1", got)
	}
	if got := doer.request.Name; got != "req.curl" {
		t.Fatalf("request name = %q, want req.curl", got)
	}
	if screen.commandPanel.Open {
		t.Fatal("command panel is open, want closed")
	}
}

func TestCommandPanelWriteQuitCommandSavesAndQuits(t *testing.T) {
	parsedURL, err := request.NewURL("http://localhost/books")
	if err != nil {
		t.Fatalf("NewURL() error = %v", err)
	}

	writer := &fakeWriter{}
	doer := &fakeDoer{response: request.Response{Status: "200 OK"}}
	screen := newScreenWithDeps([]request.Request{{Name: "req.curl", Path: "req.curl", URL: parsedURL}}, writer, doer)

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("wq")})
	screen, cmd := updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("quit command is nil")
	}
	if got := writer.calls; got != 1 {
		t.Fatalf("write calls = %d, want 1", got)
	}
	if got := writer.request.Name; got != "req.curl" {
		t.Fatalf("written request name = %q, want req.curl", got)
	}
	if got := doer.calls; got != 0 {
		t.Fatalf("do calls = %d, want 0", got)
	}
	if screen.commandPanel.Open {
		t.Fatal("command panel is open, want closed")
	}
}

func TestNewRequestFlowCreatesFileSelectsRequestAndEditsURL(t *testing.T) {
	writer := &fakeWriter{}
	screen := newScreenWithDeps(nil, writer, nil)
	screen.focusedPanel = focusedPanelRequests

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if !screen.newRequestPanel.Open {
		t.Fatal("new request panel is closed, want open")
	}
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Books")})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEnter})

	if screen.newRequestPanel.Open {
		t.Fatal("new request panel is open, want closed")
	}
	if got := writer.calls; got != 1 {
		t.Fatalf("write calls = %d, want 1", got)
	}
	if got := writer.request.Name; got != "Books.curl" {
		t.Fatalf("written request name = %q, want Books.curl", got)
	}
	if got := writer.request.Path; got != "" {
		t.Fatalf("written request path = %q, want empty", got)
	}
	if len(screen.requests) != 1 {
		t.Fatalf("requests len = %d, want 1", len(screen.requests))
	}
	if screen.selectedRequestIndex != 0 {
		t.Fatalf("selected request index = %d, want 0", screen.selectedRequestIndex)
	}
	if screen.focusedPanel != focusedPanelURL {
		t.Fatalf("focused panel = %v, want URL", screen.focusedPanel)
	}
	if !screen.insertMode {
		t.Fatal("insert mode is false, want true")
	}
}

func TestNewRequestFlowEscCancelsCreation(t *testing.T) {
	writer := &fakeWriter{}
	screen := newScreenWithDeps(nil, writer, nil)
	screen.focusedPanel = focusedPanelRequests

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Books")})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEsc})

	if screen.newRequestPanel.Open {
		t.Fatal("new request panel is open, want closed")
	}
	if got := writer.calls; got != 0 {
		t.Fatalf("write calls = %d, want 0", got)
	}
	if len(screen.requests) != 0 {
		t.Fatalf("requests len = %d, want 0", len(screen.requests))
	}
	if screen.focusedPanel != focusedPanelRequests {
		t.Fatalf("focused panel = %v, want requests", screen.focusedPanel)
	}
}

type fakeWriter struct {
	request request.Request
	calls   int
}

func (w *fakeWriter) WriteToFile(req request.Request) error {
	w.request = req
	w.calls++
	return nil
}

type fakeDoer struct {
	request  request.Request
	calls    int
	response request.Response
	err      error
}

func (d *fakeDoer) Do(req request.Request) (request.Response, error) {
	d.request = req
	d.calls++
	return d.response, d.err
}

func TestCommandPanelRunCommandRunsSelectedRequest(t *testing.T) {
	parsedURL, err := request.NewURL("http://localhost/books")
	if err != nil {
		t.Fatalf("NewURL() error = %v", err)
	}

	responseURL, err := request.NewURL("http://localhost/books")
	if err != nil {
		t.Fatalf("NewURL() error = %v", err)
	}

	doer := &fakeDoer{
		response: request.Response{
			URL:        responseURL,
			Status:     "200 OK",
			StatusCode: 200,
			Headers: map[string][]string{
				"Content-Type": {
					"application/json",
				},
			},
			Body:     `{"ok":true}`,
			Duration: 12 * time.Millisecond,
		},
	}
	screen := newScreenWithDeps([]request.Request{{Name: "req.curl", URL: parsedURL}}, nil, doer)

	updated, _ := updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	updated, _ = updateScreen(t, updated, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	updated, cmd := updateScreen(t, updated, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("do command is nil")
	}
	if !updated.requestInFlight {
		t.Fatal("request in flight is false, want true")
	}

	updated, _ = updateScreen(t, updated, firstCommandMessage(t, cmd))
	if got := doer.calls; got != 1 {
		t.Fatalf("do calls after ❯r = %d, want 1", got)
	}
	if got := doer.request.Name; got != "req.curl" {
		t.Fatalf("request name = %q, want %q", got, "req.curl")
	}
	if updated.requestInFlight {
		t.Fatal("request in flight is true, want false")
	}
	if !updated.hasResponse {
		t.Fatal("hasResponse is false, want true")
	}
	if got := updated.renderResponse(); !strings.Contains(got, "Status: 200 OK") {
		t.Fatalf("response view = %q, want status", got)
	}
}

func TestStatusPanelRendersAsInputPanel(t *testing.T) {
	screen := newScreen([]request.Request{{Name: "req.curl"}})
	screen.focusedPanel = focusedPanelDoButton

	panel := screen.renderDoButton()
	if !strings.Contains(panel, "Status") {
		t.Fatal("panel does not contain Status title")
	}
	if !strings.Contains(panel, "Ready") {
		t.Fatal("panel does not contain Ready status")
	}
	if !strings.Contains(panel, "╭") || !strings.Contains(panel, "─") {
		t.Fatalf("panel does not look like a panel: %q", panel)
	}
}

func TestStatusPanelRendersDoingStateWithSpinner(t *testing.T) {
	screen := newScreen([]request.Request{{Name: "req.curl"}})
	screen.requestInFlight = true

	panel := screen.renderDoButton()
	if !strings.Contains(panel, "Status") {
		t.Fatal("panel does not contain Status title")
	}
	if !strings.Contains(panel, "Doing...") {
		t.Fatal("panel does not contain Doing status")
	}
	if spinnerView := screen.statusSpinner.View(); spinnerView == "" || !strings.Contains(panel, spinnerView) {
		t.Fatalf("panel does not contain spinner %q: %q", spinnerView, panel)
	}
}

func TestCommandPanelRunCommandShowsLoadingState(t *testing.T) {
	doer := &fakeDoer{response: request.Response{Status: "200 OK"}}
	screen := newScreenWithDeps([]request.Request{{Name: "req.curl"}}, nil, doer)

	updated, _ := updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	updated, _ = updateScreen(t, updated, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	updated, cmd := updateScreen(t, updated, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("do command is nil")
	}
	if !updated.requestInFlight {
		t.Fatal("request in flight is false, want true")
	}
	if got := updated.renderResponse(); got != "Executing request..." {
		t.Fatalf("response view = %q, want executing state", got)
	}

	updated, _ = updateScreen(t, updated, firstCommandMessage(t, cmd))
	if updated.requestInFlight {
		t.Fatal("request in flight is true, want false")
	}
	if got := updated.renderResponse(); !strings.Contains(got, "Status: 200 OK") {
		t.Fatalf("response view = %q, want status", got)
	}
}

func TestDoButtonDoesNotEnterInsertMode(t *testing.T) {
	screen := newScreen([]request.Request{{Name: "req.curl"}})
	screen.focusedPanel = focusedPanelDoButton

	updated, cmd := updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	if cmd != nil {
		t.Fatalf("command after i on do panel = %p, want nil", cmd)
	}
	if updated.insertMode {
		t.Fatal("insert mode is true, want false")
	}
}

func TestDoButtonEnterDoesNotRunRequest(t *testing.T) {
	doer := &fakeDoer{response: request.Response{Status: "200 OK"}}
	screen := newScreenWithDeps([]request.Request{{Name: "req.curl"}}, nil, doer)
	screen.focusedPanel = focusedPanelDoButton

	updated, cmd := updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatalf("command after enter on do panel = %p, want nil", cmd)
	}
	if updated.requestInFlight {
		t.Fatal("request in flight is true, want false")
	}
	if got := doer.calls; got != 0 {
		t.Fatalf("do calls = %d, want 0", got)
	}
}

func TestCommandPanelRendersFloatingCommandInput(t *testing.T) {
	screen := newScreen([]request.Request{{Name: "Req"}})
	screen.width = 80
	screen.height = 20
	screen.commandPanel.Open = true
	screen.commandPanel.Input = "❯q"

	view := screen.View()
	if !strings.Contains(view, " Command ") {
		t.Fatal("view does not contain command panel title")
	}
	if !strings.Contains(view, "❯q") {
		t.Fatal("view does not contain command input")
	}
	plainView := ansiSequencePattern.ReplaceAllString(view, "")
	if !strings.Contains(plainView, "") || !strings.Contains(plainView, "Commands") || !strings.Contains(plainView, "") {
		t.Fatalf("view does not contain command help title capsule: %q", plainView)
	}
	if !strings.Contains(plainView, "Command") || !strings.Contains(plainView, "Action") {
		t.Fatalf("view does not contain command help table header: %q", plainView)
	}
	if !strings.Contains(plainView, "wr") || !strings.Contains(plainView, "Save and run") {
		t.Fatalf("view does not contain write-run command help: %q", plainView)
	}
	if !strings.Contains(plainView, "wq") || !strings.Contains(plainView, "Save and quit") {
		t.Fatalf("view does not contain write-quit command help: %q", plainView)
	}
}

func TestCommandPanelOverlayPreservesBaseLineOutsidePanel(t *testing.T) {
	screen := newScreen([]request.Request{{Name: "Req"}})
	screen.commandPanel.Open = true
	screen.commandPanel.Input = "❯"

	baseLine := "AAAAAAAAAABBBBBBBBBBBBBBBBBBBBCCCCCCCCCC"
	baseView := strings.Join([]string{
		baseLine,
		baseLine,
		baseLine,
		baseLine,
		baseLine,
	}, "\n")
	view := screen.commandPanel.Render(baseView, 40)
	line := strings.Split(view, "\n")[4]

	if !strings.HasPrefix(line, "AAAAAAAAA") {
		t.Fatalf("left side was not preserved: %q", line)
	}
	if !strings.HasSuffix(line, "CCCCCCCCC") {
		t.Fatalf("right side was not preserved: %q", line)
	}
}

func enterHeaderInsertMode(t *testing.T, screen Screen) Screen {
	t.Helper()
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	return screen
}

func enterBodyTextInsertMode(t *testing.T, screen Screen) Screen {
	t.Helper()
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	return screen
}

func enterURLTextInsertMode(t *testing.T, screen Screen) Screen {
	t.Helper()
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	return screen
}

func newScreen(requests []request.Request) Screen {
	return newScreenWithDeps(requests, nil, nil)

}

func newScreenWithDeps(requests []request.Request, writer *fakeWriter, doer *fakeDoer) Screen {
	paths := make([]string, len(requests))
	requestsByPath := make(map[string]request.Request, len(requests))
	for i, req := range requests {
		path := req.Path
		if path == "" {
			path = req.Name
		}
		if path == "" {
			path = fmt.Sprintf("req-%d.curl", i)
		}
		paths[i] = path
		requestsByPath[path] = req
	}
	return NewScreen(paths, writer, doer, fakeLoader{requests: requestsByPath})
}

type fakeLoader struct {
	requests map[string]request.Request
}

func (l fakeLoader) Load(path string) (request.Request, error) {
	return l.requests[path], nil
}

type countingLoader struct {
	requests map[string]request.Request
	calls    map[string]int
}

func (l *countingLoader) Load(path string) (request.Request, error) {
	if l.calls == nil {
		l.calls = make(map[string]int)
	}
	l.calls[path]++
	return l.requests[path], nil
}

func updateScreen(t *testing.T, screen Screen, msg tea.Msg) (Screen, tea.Cmd) {
	t.Helper()
	model, cmd := screen.Update(msg)
	updated, ok := model.(Screen)
	if !ok {
		t.Fatalf("updated model has type %T, want Screen", model)
	}
	return updated, cmd
}

func firstCommandMessage(t *testing.T, cmd tea.Cmd) tea.Msg {
	t.Helper()
	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		return msg
	}
	if len(batch) == 0 {
		t.Fatal("batch command is empty")
	}
	return batch[0]()
}
