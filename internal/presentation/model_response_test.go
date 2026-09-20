package presentation

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jzes/reqman/internal/domain/request"
	"github.com/jzes/reqman/internal/presentation/panels/response"
)

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
	screen.responsePanel.SelectTab(response.TabBody)
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
	screen.responsePanel.SelectTab(response.TabHeaders)
	screen.response = request.Response{Headers: request.NewHeadersFrom(map[string][]string{
		"X-Zeta":       {"last"},
		"Content-Type": {"application/json"},
		"Accept":       {"application/json", "text/plain"},
	})}

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
	if screen.responsePanel.SelectedTab() != response.TabBody {
		t.Fatalf("selected response tab after ] = %v, want body", screen.responsePanel.SelectedTab())
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	if screen.responsePanel.SelectedTab() != response.TabStats {
		t.Fatalf("selected response tab after [ = %v, want stats", screen.responsePanel.SelectedTab())
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	if screen.responsePanel.SelectedTab() != response.TabRaw {
		t.Fatalf("selected response tab after wrapped [ = %v, want raw", screen.responsePanel.SelectedTab())
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
	if screen.responsePanel.SelectedTab() != response.TabBody {
		t.Fatalf("selected response tab after tab = %v, want body", screen.responsePanel.SelectedTab())
	}

	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyShiftTab})
	if screen.responsePanel.SelectedTab() != response.TabStats {
		t.Fatalf("selected response tab after shift+tab = %v, want stats", screen.responsePanel.SelectedTab())
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

func TestDoButtonEnterRunsRequest(t *testing.T) {
	doer := &fakeDoer{response: request.Response{Status: "200 OK"}}
	screen := newScreenWithDeps([]request.Request{{Name: "req.curl"}}, nil, doer)
	screen.focusedPanel = focusedPanelDoButton

	updated, cmd := updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("command after enter on do panel is nil")
	}
	if !updated.requestInFlight {
		t.Fatal("request in flight is false, want true")
	}

	updated, _ = updateScreen(t, updated, firstCommandMessage(t, cmd))
	if got := doer.calls; got != 1 {
		t.Fatalf("do calls = %d, want 1", got)
	}
}
