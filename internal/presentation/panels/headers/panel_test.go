package headers

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TestPanelEditsRows(t *testing.T) {
	panel := NewPanel()

	if !panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Accept")}) {
		t.Fatal("first key update = false, want true")
	}
	panel.Update(tea.KeyMsg{Type: tea.KeyTab})
	panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("application/json")})

	rows := panel.Rows()
	if len(rows) != 1 {
		t.Fatalf("rows len = %d, want 1", len(rows))
	}
	if rows[0] != (Row{Key: "Accept", Value: "application/json"}) {
		t.Fatalf("row = %+v, want Accept header", rows[0])
	}
}

func TestPanelInsertNavigateAndDeleteRows(t *testing.T) {
	panel := NewPanel()
	panel.SetRows([]Row{{Key: "Accept", Value: "json"}})
	panel.Update(tea.KeyMsg{Type: tea.KeyTab})
	if !panel.Update(tea.KeyMsg{Type: tea.KeyEnter}) {
		t.Fatal("enter on value cell = false, want row insertion")
	}
	panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Authorization")})
	panel.Update(tea.KeyMsg{Type: tea.KeyUp})

	if !panel.Update(tea.KeyMsg{Type: tea.KeyDelete}) {
		t.Fatal("delete update = false, want true")
	}
	rows := panel.Rows()
	if len(rows) != 1 || rows[0].Key != "Authorization" {
		t.Fatalf("rows after delete = %+v, want Authorization only", rows)
	}
}

func TestPanelBackspaceReportsMutation(t *testing.T) {
	panel := NewPanel()
	panel.SetRows([]Row{{Key: "Accept", Value: "json"}})

	if !panel.Update(tea.KeyMsg{Type: tea.KeyBackspace}) {
		t.Fatal("backspace with text = false, want true")
	}
	if got := panel.Rows()[0].Key; got != "Accep" {
		t.Fatalf("key after backspace = %q, want Accep", got)
	}
}

func TestRowsReturnsCopy(t *testing.T) {
	panel := NewPanel()
	panel.SetRows([]Row{{Key: "Accept", Value: "json"}})

	rows := panel.Rows()
	rows[0].Key = "Changed"

	if got := panel.Rows()[0].Key; got != "Accept" {
		t.Fatalf("panel row key = %q, want Accept", got)
	}
}

func TestViewRendersEmptyAndFocusedRows(t *testing.T) {
	panel := NewPanel()
	view := panel.View(10, 10, ViewOptions{})
	if !strings.Contains(view, "Header") || !strings.Contains(view, "(empty)") {
		t.Fatalf("empty view = %q, want header and empty placeholder", view)
	}

	panel.SetRows([]Row{{Key: "Accept", Value: "json"}})
	view = panel.View(10, 10, ViewOptions{
		Focused:     true,
		InsertMode:  true,
		FocusMarker: "❯ ",
		FocusStyle:  lipgloss.NewStyle(),
	})
	if !strings.Contains(view, "❯ ") || !strings.Contains(view, "Accept") || !strings.Contains(view, "█") {
		t.Fatalf("focused view = %q, want marker, row and cursor", view)
	}
}
