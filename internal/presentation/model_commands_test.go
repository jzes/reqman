package presentation

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jzes/reqman/internal/app/apperror"
	"github.com/jzes/reqman/internal/domain/request"
	"github.com/jzes/reqman/internal/presentation/jsoneditor"
)

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
	if screen.body.Mode() != jsoneditor.NavigateMode {
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

func TestCommandPanelWriteCommandSavesRequestWithName(t *testing.T) {
	parsedURL, err := request.NewURL("http://localhost/books")
	if err != nil {
		t.Fatalf("NewURL() error = %v", err)
	}

	writer := &fakeWriter{}
	screen := newScreenWithDeps([]request.Request{{Name: "req.curl", Path: "req.curl", URL: parsedURL}}, writer, nil)

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("w nova-req")})
	screen, cmd := updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatalf("command after ❯w nova-req = %p, want nil", cmd)
	}
	if got := writer.calls; got != 1 {
		t.Fatalf("write calls = %d, want 1", got)
	}
	if got := writer.request.Name; got != "nova-req.curl" {
		t.Fatalf("written request name = %q, want nova-req.curl", got)
	}
	if got := writer.request.Path; got != "" {
		t.Fatalf("written request path = %q, want empty", got)
	}
	if got := screen.requests[screen.selectedRequestIndex].Name; got != "nova-req.curl" {
		t.Fatalf("selected request name = %q, want nova-req.curl", got)
	}
	if got := screen.requestPaths[screen.selectedRequestIndex]; got != "nova-req.curl" {
		t.Fatalf("selected request path = %q, want nova-req.curl", got)
	}
	if screen.commandPanel.Open {
		t.Fatal("command panel is open, want closed")
	}
}

func TestCommandPanelAddCommandCreatesEmptyRequest(t *testing.T) {
	writer := &fakeWriter{}
	screen := newScreenWithDeps(nil, writer, nil)

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	screen, cmd := updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatalf("command after ❯a = %p, want nil", cmd)
	}
	if got := writer.calls; got != 0 {
		t.Fatalf("write calls = %d, want 0", got)
	}
	if len(screen.requests) != 1 {
		t.Fatalf("requests len = %d, want 1", len(screen.requests))
	}
	if got := screen.requests[0].Name; got != newRequestName {
		t.Fatalf("new request name = %q, want %q", got, newRequestName)
	}
	if got := screen.requests[0].Method; got != request.MethodGet {
		t.Fatalf("new request method = %q, want GET", got)
	}
	if got := screen.url.Value(); got != "" {
		t.Fatalf("url value = %q, want empty", got)
	}
	if got := screen.body.Value(); got != "" {
		t.Fatalf("body value = %q, want empty", got)
	}
	if screen.focusedPanel != focusedPanelURL {
		t.Fatalf("focused panel = %v, want URL", screen.focusedPanel)
	}
	if !screen.insertMode {
		t.Fatal("insert mode is false, want true")
	}
	if screen.commandPanel.Open {
		t.Fatal("command panel is open, want closed")
	}
}

func TestCommandPanelWriteCommandShowsSaveError(t *testing.T) {
	parsedURL, err := request.NewURL("http://localhost/books")
	if err != nil {
		t.Fatalf("NewURL() error = %v", err)
	}

	writer := &fakeWriter{err: errors.New("disk full")}
	screen := newScreenWithDeps([]request.Request{{Name: "req.curl", Path: "req.curl", URL: parsedURL}}, writer, nil)

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	screen, cmd := updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatalf("command after failed ❯w = %p, want nil", cmd)
	}
	if got := writer.calls; got != 1 {
		t.Fatalf("write calls = %d, want 1", got)
	}
	if got := screen.renderResponse(); !strings.Contains(got, "Failed to save request: disk full") {
		t.Fatalf("response view = %q, want save error", got)
	}
	if screen.hasResponse {
		t.Fatal("hasResponse is true, want false")
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

func TestCommandPanelWriteRunDoesNotRunWhenSaveFails(t *testing.T) {
	parsedURL, err := request.NewURL("http://localhost/books")
	if err != nil {
		t.Fatalf("NewURL() error = %v", err)
	}

	writer := &fakeWriter{err: errors.New("permission denied")}
	doer := &fakeDoer{response: request.Response{Status: "200 OK"}}
	screen := newScreenWithDeps([]request.Request{{Name: "req.curl", Path: "req.curl", URL: parsedURL}}, writer, doer)

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("wr")})
	screen, cmd := updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatalf("command after failed ❯wr = %p, want nil", cmd)
	}
	if got := writer.calls; got != 1 {
		t.Fatalf("write calls = %d, want 1", got)
	}
	if got := doer.calls; got != 0 {
		t.Fatalf("do calls = %d, want 0", got)
	}
	if screen.requestInFlight {
		t.Fatal("request in flight is true, want false")
	}
	if got := screen.renderResponse(); !strings.Contains(got, "Failed to save request: permission denied") {
		t.Fatalf("response view = %q, want save error", got)
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

func TestCommandPanelWriteQuitDoesNotQuitWhenSaveFails(t *testing.T) {
	parsedURL, err := request.NewURL("http://localhost/books")
	if err != nil {
		t.Fatalf("NewURL() error = %v", err)
	}

	writer := &fakeWriter{err: errors.New("read-only filesystem")}
	screen := newScreenWithDeps([]request.Request{{Name: "req.curl", Path: "req.curl", URL: parsedURL}}, writer, nil)

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("wq")})
	screen, cmd := updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatalf("command after failed ❯wq = %p, want nil", cmd)
	}
	if got := writer.calls; got != 1 {
		t.Fatalf("write calls = %d, want 1", got)
	}
	if got := screen.renderResponse(); !strings.Contains(got, "Failed to save request: read-only filesystem") {
		t.Fatalf("response view = %q, want save error", got)
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

func TestNewRequestFlowShowsCreateError(t *testing.T) {
	writer := &fakeWriter{err: errors.New("cannot create file")}
	screen := newScreenWithDeps(nil, writer, nil)
	screen.focusedPanel = focusedPanelRequests

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Books")})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEnter})

	if got := writer.calls; got != 1 {
		t.Fatalf("write calls = %d, want 1", got)
	}
	if len(screen.requests) != 0 {
		t.Fatalf("requests len = %d, want 0", len(screen.requests))
	}
	if got := screen.renderResponse(); !strings.Contains(got, "Failed to create request: cannot create file") {
		t.Fatalf("response view = %q, want create error", got)
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
			Headers: request.NewHeadersFrom(map[string][]string{
				"Content-Type": {
					"application/json",
				},
			}),
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

func TestEscCancelsInFlightRequest(t *testing.T) {
	doer := &fakeDoer{}
	screen := newScreenWithDeps([]request.Request{{Name: "req.curl"}}, nil, doer)
	screen.focusedPanel = focusedPanelDoButton

	updated, cmd := updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("do command is nil")
	}
	updated, _ = updateScreen(t, updated, tea.KeyMsg{Type: tea.KeyEsc})

	if updated.requestInFlight {
		t.Fatal("request in flight is true, want false")
	}
	if got := updated.responseError; got != "Request canceled" {
		t.Fatalf("response error = %q, want canceled", got)
	}
	updated, _ = updateScreen(t, updated, firstCommandMessage(t, cmd))
	if doer.ctx == nil {
		t.Fatal("doer context is nil")
	}
	select {
	case <-doer.ctx.Done():
	default:
		t.Fatal("request context was not canceled")
	}
}

func TestFatalRequestErrorPromptsBeforeQuittingApplication(t *testing.T) {
	doer := &fakeDoer{err: fmt.Errorf("request setup failed: %w", apperror.ErrFatal)}
	screen := newScreenWithDeps([]request.Request{{Name: "req.curl"}}, nil, doer)
	screen.focusedPanel = focusedPanelDoButton

	updated, cmd := updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("do command is nil")
	}

	updated, quitCmd := updateScreen(t, updated, firstCommandMessage(t, cmd))
	if quitCmd != nil {
		t.Fatalf("quit command after fatal result = %p, want nil", quitCmd)
	}
	if !updated.fatalError {
		t.Fatal("fatal error is false, want true")
	}
	if updated.requestInFlight {
		t.Fatal("request in flight is true, want false")
	}
	if got := updated.responseError; got == "" {
		t.Fatal("response error is empty, want fatal error message")
	}
	if got := updated.renderResponse(); !strings.Contains(got, "Fatal error:") || !strings.Contains(got, "Press any key to quit.") {
		t.Fatalf("response view = %q, want fatal prompt", got)
	}

	_, quitCmd = updateScreen(t, updated, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if quitCmd == nil {
		t.Fatal("quit command is nil")
	}
	quitMsg := quitCmd()
	if _, ok := quitMsg.(tea.QuitMsg); !ok {
		t.Fatalf("quit command returned %T, want tea.QuitMsg", quitMsg)
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
	if !strings.Contains(plainView, "?") || !strings.Contains(plainView, "Open help") {
		t.Fatalf("view does not contain help command: %q", plainView)
	}
}

func TestCommandPanelHelpCommandOpensHelpPopup(t *testing.T) {
	screen := newScreen([]request.Request{{Name: "Req"}})
	screen.width = 80
	screen.height = 24

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	screen, cmd := updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatalf("command after :? = %p, want nil", cmd)
	}
	if screen.commandPanel.Open {
		t.Fatal("command panel is open, want closed")
	}
	if !screen.helpOpen {
		t.Fatal("help popup is closed, want open")
	}

	plainView := ansiSequencePattern.ReplaceAllString(screen.View(), "")
	if !strings.Contains(plainView, "Help") {
		t.Fatalf("view does not contain Help title: %q", plainView)
	}
	if !strings.Contains(plainView, "Panels:") || !strings.Contains(plainView, "- Requests lists the .curl files") {
		t.Fatalf("view does not contain help content: %q", plainView)
	}
}

func TestHelpPopupEscCloses(t *testing.T) {
	screen := newScreen([]request.Request{{Name: "Req"}})
	screen.helpOpen = true
	screen.commandPanel.Open = false

	screen, cmd := updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEsc})
	if cmd != nil {
		t.Fatalf("command after help Esc = %p, want nil", cmd)
	}
	if screen.helpOpen {
		t.Fatal("help popup is open, want closed")
	}
}

func TestHelpTitleBarRendersOutsideBorderAsCapsule(t *testing.T) {
	title := ansiSequencePattern.ReplaceAllString(renderHelpTitleBar(30), "")
	if !strings.HasPrefix(title, "") || !strings.HasSuffix(title, "") {
		t.Fatalf("help title bar is not a capsule: %q", title)
	}
	if !strings.Contains(title, "Help") {
		t.Fatalf("help title bar does not contain title: %q", title)
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
