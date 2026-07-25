package presentation

import (
	"reflect"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jzes/reqman/internal/application"
	"github.com/jzes/reqman/internal/domain/request"
)

func TestViewRendersBodyAndResponsePanels(t *testing.T) {
	screen := NewScreen([]request.Request{{Name: "req.curl", Body: "{}"}})
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

func TestResponsePanelVerticalNavigation(t *testing.T) {
	screen := NewScreen([]request.Request{{Name: "req.curl"}})
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
	screen := NewScreen([]request.Request{{Name: "req.curl"}})
	screen.focusedPanel = focusedPanelResponse

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	if screen.focusedPanel != focusedPanelRequests {
		t.Fatalf("focused panel = %v, want requests", screen.focusedPanel)
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
	screen := NewScreen([]request.Request{{Headers: map[string]string{}}})
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
	screen := NewScreen([]request.Request{{Headers: map[string]string{"Accept": "text/plain"}}})
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
	screen := NewScreen([]request.Request{{Headers: map[string]string{
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
	screen := NewScreen([]request.Request{{Headers: map[string]string{}}})
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
	screen := NewScreen([]request.Request{{Body: ""}})
	screen.focusedPanel = focusedPanelBody

	screen = enterBodyTextInsertMode(t, screen)
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("abc")})

	if got := screen.bodyTextArea.Value(); got != "abc" {
		t.Fatalf("body textarea = %q, want abc", got)
	}
	if got := screen.requests[0].Body; got != "abc" {
		t.Fatalf("request body = %q, want abc", got)
	}
}

func TestBodyEditorAutocompletesBraces(t *testing.T) {
	screen := NewScreen([]request.Request{{Body: ""}})
	screen.focusedPanel = focusedPanelBody

	screen = enterBodyTextInsertMode(t, screen)
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'{'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(`"a":1`)})

	if got := screen.requests[0].Body; got != `{"a":1}` {
		t.Fatalf("request body = %q, want {\"a\":1}", got)
	}
}

func TestBodyEditorAutocompletesBrackets(t *testing.T) {
	screen := NewScreen([]request.Request{{Body: ""}})
	screen.focusedPanel = focusedPanelBody

	screen = enterBodyTextInsertMode(t, screen)
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})

	if got := screen.requests[0].Body; got != `[1]` {
		t.Fatalf("request body = %q, want [1]", got)
	}
}

func TestBodyEditorFormatsJSON(t *testing.T) {
	screen := NewScreen([]request.Request{{Body: `{"a":[1,true]}`}})
	screen.focusedPanel = focusedPanelBody

	screen = enterBodyTextInsertMode(t, screen)
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyCtrlF})

	want := "{\n  \"a\": [\n    1,\n    true\n  ]\n}"
	if got := screen.bodyTextArea.Value(); got != want {
		t.Fatalf("body textarea = %q, want %q", got, want)
	}
	if got := screen.requests[0].Body; got != want {
		t.Fatalf("request body = %q, want %q", got, want)
	}
}

func TestBodyEditorFormatInvalidJSONDoesNotChangeBody(t *testing.T) {
	body := `{"bad":`
	screen := NewScreen([]request.Request{{Body: body}})
	screen.focusedPanel = focusedPanelBody

	screen = enterBodyTextInsertMode(t, screen)
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyCtrlF})

	if got := screen.bodyTextArea.Value(); got != body {
		t.Fatalf("body textarea = %q, want %q", got, body)
	}
	if got := screen.requests[0].Body; got != body {
		t.Fatalf("request body = %q, want %q", got, body)
	}
}

func TestBodyNormalModeIEntersNavigationMode(t *testing.T) {
	screen := NewScreen([]request.Request{{Body: ""}})
	screen.focusedPanel = focusedPanelBody

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	if !screen.insertMode {
		t.Fatal("insert mode is false, want true")
	}
	if screen.bodyMode != bodyModeNavigate {
		t.Fatalf("body mode = %v, want navigate", screen.bodyMode)
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if got := screen.requests[0].Body; got != "" {
		t.Fatalf("request body = %q, want empty", got)
	}
}

func TestBodyNavigationModeIEntersTextInsertMode(t *testing.T) {
	screen := NewScreen([]request.Request{{Body: ""}})
	screen.focusedPanel = focusedPanelBody

	screen = enterBodyTextInsertMode(t, screen)
	if screen.bodyMode != bodyModeInsert {
		t.Fatalf("body mode = %v, want insert", screen.bodyMode)
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if got := screen.requests[0].Body; got != "x" {
		t.Fatalf("request body = %q, want x", got)
	}
}

func TestBodyEditorFormatsJSONWhenLeavingTextInsertMode(t *testing.T) {
	screen := NewScreen([]request.Request{{Body: `{"a":[1,true]}`}})
	screen.focusedPanel = focusedPanelBody

	screen = enterBodyTextInsertMode(t, screen)
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEsc})

	want := "{\n  \"a\": [\n    1,\n    true\n  ]\n}"
	if got := screen.bodyTextArea.Value(); got != want {
		t.Fatalf("body textarea = %q, want %q", got, want)
	}
	if got := screen.requests[0].Body; got != want {
		t.Fatalf("request body = %q, want %q", got, want)
	}
}

func TestBodyEscTransitionsFromInsertToNavigateToPanel(t *testing.T) {
	screen := NewScreen([]request.Request{{Body: ""}})
	screen.focusedPanel = focusedPanelBody
	screen = enterBodyTextInsertMode(t, screen)

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEsc})
	if screen.bodyMode != bodyModeNavigate {
		t.Fatalf("body mode = %v, want navigate", screen.bodyMode)
	}
	if !screen.insertMode {
		t.Fatal("insert mode is false, want true after returning to body navigation")
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEsc})
	if screen.bodyMode != bodyModePanel {
		t.Fatalf("body mode = %v, want panel", screen.bodyMode)
	}
	if screen.insertMode {
		t.Fatal("insert mode is true, want false after returning to panel mode")
	}
}

func TestBodyNavigationModeMotionsStayInBodyPanel(t *testing.T) {
	body := "one\ntwo\nthree"
	screen := NewScreen([]request.Request{{Body: body}})
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
	screen := NewScreen([]request.Request{{Body: body}})
	screen.focusedPanel = focusedPanelBody

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if screen.focusedPanel != focusedPanelHeaders {
		t.Fatalf("focused panel = %v, want headers", screen.focusedPanel)
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if screen.focusedPanel != focusedPanelBody {
		t.Fatalf("focused panel = %v, want body", screen.focusedPanel)
	}

	if got := screen.bodyTextArea.Value(); got != body {
		t.Fatalf("body textarea = %q, want %q", got, body)
	}
	if got := screen.requests[0].Body; got != body {
		t.Fatalf("request body = %q, want %q", got, body)
	}
}

func TestBodyNavigationWordMotionsMoveCursor(t *testing.T) {
	screen := NewScreen([]request.Request{{Body: "one two three"}})
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
	screen := NewScreen([]request.Request{{Body: `{"a":[1,true]}`}})
	screen.focusedPanel = focusedPanelBody

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEsc})

	want := "{\n  \"a\": [\n    1,\n    true\n  ]\n}"
	if got := screen.bodyTextArea.Value(); got != want {
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
	screen := NewScreen([]request.Request{{URL: parsedURL}}, application.NewRequestWriter(writer))
	screen.focusedPanel = focusedPanelBody
	screen = enterBodyTextInsertMode(t, screen)
	if got := screen.bodyTextArea.Value(); got != "" {
		t.Fatalf("body textarea = %q, want empty", got)
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEsc})
	if got := writer.calls; got != 0 {
		t.Fatalf("write calls = %d, want 0", got)
	}
	if screen.bodyMode != bodyModeNavigate {
		t.Fatalf("body mode = %v, want navigate", screen.bodyMode)
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
	screen := NewScreen([]request.Request{{Name: "req.curl", Path: "req.curl", URL: parsedURL}}, application.NewRequestWriter(writer))

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	screen, cmd := updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatalf("command after :w = %p, want nil", cmd)
	}
	if got := writer.calls; got != 1 {
		t.Fatalf("write calls = %d, want 1", got)
	}
	if got := writer.request.Name; got != "req.curl" {
		t.Fatalf("written request name = %q, want req.curl", got)
	}
	if screen.commandPanelOpen {
		t.Fatal("command panel is open, want closed")
	}
}

type fakeWriter struct {
	request request.Request
	calls   int
}

func (w *fakeWriter) Write(req request.Request) error {
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

func TestDoButtonRunsSelectedRequest(t *testing.T) {
	parsedURL, err := request.NewURL("http://localhost/books")
	if err != nil {
		t.Fatalf("NewURL() error = %v", err)
	}

	responseURL, err := request.NewURL("http://localhost/books")
	if err != nil {
		t.Fatalf("NewURL() error = %v", err)
	}

	doer := &fakeDoer{response: request.Response{
		URL:        responseURL,
		Status:     "200 OK",
		StatusCode: 200,
		Headers:    map[string][]string{"Content-Type": []string{"application/json"}},
		Body:       `{"ok":true}`,
		Duration:   12 * time.Millisecond,
	}}
	screen := NewScreen([]request.Request{{Name: "req.curl", URL: parsedURL}}, application.NewRequestDoer(doer))
	screen.focusedPanel = focusedPanelDoButton

	updated, cmd := updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("do command is nil")
	}
	if !updated.requestInFlight {
		t.Fatal("request in flight is false, want true")
	}

	updated, _ = updateScreen(t, updated, cmd())
	if got := doer.calls; got != 1 {
		t.Fatalf("do calls after enter = %d, want 1", got)
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

func TestDoButtonRendersAsInputPanel(t *testing.T) {
	screen := NewScreen([]request.Request{{Name: "req.curl"}})
	screen.focusedPanel = focusedPanelDoButton

	panel := screen.renderDoButton()
	if !strings.Contains(panel, "Do Req") {
		t.Fatal("panel does not contain Do Req title")
	}
	if !strings.Contains(panel, "Do") {
		t.Fatal("panel does not contain Do label")
	}
	if !strings.Contains(panel, "╭") || !strings.Contains(panel, "─") {
		t.Fatalf("panel does not look like a panel: %q", panel)
	}
}

func TestDoButtonEnterShowsLoadingState(t *testing.T) {
	doer := &fakeDoer{response: request.Response{Status: "200 OK"}}
	screen := NewScreen([]request.Request{{Name: "req.curl"}}, application.NewRequestDoer(doer))
	screen.focusedPanel = focusedPanelDoButton

	updated, cmd := updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("do command is nil")
	}
	if !updated.requestInFlight {
		t.Fatal("request in flight is false, want true")
	}
	if got := updated.renderResponse(); got != "Executing request..." {
		t.Fatalf("response view = %q, want executing state", got)
	}

	updated, _ = updateScreen(t, updated, cmd())
	if updated.requestInFlight {
		t.Fatal("request in flight is true, want false")
	}
	if got := updated.renderResponse(); !strings.Contains(got, "Status: 200 OK") {
		t.Fatalf("response view = %q, want status", got)
	}
}

func TestCommandPanelRendersFloatingCommandInput(t *testing.T) {
	screen := NewScreen([]request.Request{{Name: "Req"}})
	screen.width = 80
	screen.height = 20
	screen.commandPanelOpen = true
	screen.commandInput = ":q"

	view := screen.View()
	if !strings.Contains(view, " Command ") {
		t.Fatal("view does not contain command panel title")
	}
	if !strings.Contains(view, ":q") {
		t.Fatal("view does not contain command input")
	}
}

func TestCommandPanelOverlayPreservesBaseLineOutsidePanel(t *testing.T) {
	screen := NewScreen([]request.Request{{Name: "Req"}})
	screen.commandPanelOpen = true
	screen.commandInput = ":"

	baseLine := "AAAAAAAAAABBBBBBBBBBBBBBBBBBBBCCCCCCCCCC"
	baseView := strings.Join([]string{
		baseLine,
		baseLine,
		baseLine,
		baseLine,
		baseLine,
	}, "\n")
	view := screen.renderCommandPanel(baseView, 40)
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

func updateScreen(t *testing.T, screen Screen, msg tea.Msg) (Screen, tea.Cmd) {
	t.Helper()
	model, cmd := screen.Update(msg)
	updated, ok := model.(Screen)
	if !ok {
		t.Fatalf("updated model has type %T, want Screen", model)
	}
	return updated, cmd
}
