package presentation

import (
	"reflect"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jzes/reqman/internal/domain/request"
	"github.com/jzes/reqman/internal/presentation/panels/headers"
)

func TestHeadersFromDomainSortsRows(t *testing.T) {
	rows := headersFromDomain(request.NewHeadersFrom(map[string][]string{
		"X-Zeta":  {"last"},
		"Accept":  {"application/json", "text/plain"},
		"Content": {"text/plain"},
	}))

	got := []headers.Row{rows[0], rows[1], rows[2], rows[3]}
	want := []headers.Row{
		{Key: "Accept", Value: "application/json"},
		{Key: "Accept", Value: "text/plain"},
		{Key: "Content", Value: "text/plain"},
		{Key: "X-Zeta", Value: "last"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("header rows = %v, want %v", got, want)
	}
}

func TestHeadersEditorCreatesFirstHeader(t *testing.T) {
	screen := newScreen([]request.Request{{}})
	screen.focusedPanel = focusedPanelHeaders

	screen = enterHeaderInsertMode(t, screen)
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Accept")})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyTab})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("application/json")})

	if got := screen.requests[0].Headers.Values("Accept"); !reflect.DeepEqual(got, []string{"application/json"}) {
		t.Fatalf("Accept header = %v, want [application/json]", got)
	}
}

func TestHeadersEditorEditsKeyAndValue(t *testing.T) {
	screen := newScreen([]request.Request{{Headers: request.NewHeadersFrom(map[string][]string{"Accept": {"text/plain"}})}})
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

	if got := screen.requests[0].Headers.Values("Accept"); len(got) != 0 {
		t.Fatal("old Accept key still exists")
	}
	if got := screen.requests[0].Headers.Values("Content-Type"); !reflect.DeepEqual(got, []string{"application/json"}) {
		t.Fatalf("Content-Type header = %v, want [application/json]", got)
	}
}

func TestHeadersEditorDeletesSelectedRow(t *testing.T) {
	screen := newScreen([]request.Request{{Headers: request.NewHeadersFrom(map[string][]string{
		"Accept":        {"application/json"},
		"Authorization": {"Bearer token"},
	})}})
	screen.focusedPanel = focusedPanelHeaders

	screen = enterBodyTextInsertMode(t, screen)
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyDelete})

	if got := screen.requests[0].Headers.Values("Accept"); len(got) != 0 {
		t.Fatal("deleted Accept key still exists")
	}
	if got := screen.requests[0].Headers.Values("Authorization"); !reflect.DeepEqual(got, []string{"Bearer token"}) {
		t.Fatalf("Authorization header = %v, want [Bearer token]", got)
	}
}

func TestSyncHeadersSkipsEmptyKeysAndPreservesDuplicates(t *testing.T) {
	screen := newScreen([]request.Request{{}})
	screen.headersEditor.SetRows([]headers.Row{
		{Key: "", Value: "ignored"},
		{Key: "Accept", Value: "text/plain"},
		{Key: "Accept", Value: "application/json"},
	})

	screen.syncHeadersToSelectedRequest()

	if got := screen.requests[0].Headers.Values(""); len(got) != 0 {
		t.Fatal("empty key was synced")
	}
	if got := screen.requests[0].Headers.Values("Accept"); !reflect.DeepEqual(got, []string{"text/plain", "application/json"}) {
		t.Fatalf("Accept header = %v, want [text/plain application/json]", got)
	}
}
