package presentation

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jzes/reqman/internal/domain/request"
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
