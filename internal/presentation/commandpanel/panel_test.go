package commandpanel

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPanelActivationAndClose(t *testing.T) {
	panel := Panel{}
	panel.Activate("❯")

	if !panel.Open {
		t.Fatal("panel is closed, want open")
	}
	if panel.Title != "Command" {
		t.Fatalf("title = %q, want Command", panel.Title)
	}
	if panel.Input != "❯" {
		t.Fatalf("input = %q, want prompt", panel.Input)
	}

	panel.Close()
	if panel.Open || panel.Input != "" || panel.Title != "" {
		t.Fatalf("closed panel = %+v, want zero open state", panel)
	}
}

func TestHandleKeyReturnsActions(t *testing.T) {
	tests := []struct {
		input string
		want  Action
	}{
		{input: "❯q", want: ActionQuit},
		{input: "❯w", want: ActionWrite},
		{input: "❯w new-request", want: ActionWrite},
		{input: "❯r", want: ActionRun},
		{input: "❯wr", want: ActionWriteRun},
		{input: "❯wq", want: ActionWriteQuit},
		{input: "❯?", want: ActionHelp},
		{input: "❯a", want: ActionAdd},
		{input: "❯unknown", want: ActionNone},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			panel := Panel{Open: true, Input: tt.input, Title: "Command"}
			got := panel.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
			if got != tt.want {
				t.Fatalf("HandleKey() = %v, want %v", got, tt.want)
			}
			if panel.Open {
				t.Fatal("panel is open, want closed")
			}
		})
	}
}

func TestHandleKeyEditsInput(t *testing.T) {
	panel := Panel{Open: true, Input: "❯"}

	action := panel.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("wq")})
	if action != ActionNone {
		t.Fatalf("action after runes = %v, want none", action)
	}
	if panel.Input != "❯wq" {
		t.Fatalf("input = %q, want ❯wq", panel.Input)
	}

	panel.HandleKey(tea.KeyMsg{Type: tea.KeyBackspace})
	if panel.Input != "❯w" {
		t.Fatalf("input after backspace = %q, want ❯w", panel.Input)
	}
}

func TestHandleKeyEditsInputWithSpace(t *testing.T) {
	panel := Panel{Open: true, Input: "❯"}

	panel.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	panel.HandleKey(tea.KeyMsg{Type: tea.KeySpace})
	panel.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("nova_req")})

	if panel.Input != "❯w nova_req" {
		t.Fatalf("input = %q, want ❯w nova_req", panel.Input)
	}
}

func TestHandleKeyClosesWhenBackspaceClearsInput(t *testing.T) {
	panel := Panel{Open: true, Input: "x"}

	panel.HandleKey(tea.KeyMsg{Type: tea.KeyBackspace})
	if panel.Open {
		t.Fatal("panel is open, want closed")
	}
}

func TestRenderOverlaysCommandPanelAndHelp(t *testing.T) {
	panel := Panel{Open: true, Input: "❯w", Title: "Command"}
	base := strings.Join([]string{"AAAAAAAAAA", "BBBBBBBBBB", "CCCCCCCCCC", "DDDDDDDDDD", "EEEEEEEEEE"}, "\n")

	view := panel.Render(base, 80)
	plain := ansiSequencePattern.ReplaceAllString(view, "")
	if !strings.Contains(plain, "Command") || !strings.Contains(plain, "❯w") {
		t.Fatalf("rendered view does not contain command panel: %q", plain)
	}
	if !strings.Contains(plain, "Commands") || !strings.Contains(plain, "Save and run") {
		t.Fatalf("rendered view does not contain command help: %q", plain)
	}
}

func TestRenderClosedPanelReturnsBaseView(t *testing.T) {
	base := "base"
	if got := (Panel{}).Render(base, 80); got != base {
		t.Fatalf("Render() = %q, want base", got)
	}
}

func TestTextHelpers(t *testing.T) {
	if got := padOrTruncate("abcdef", 3); got != "abc" {
		t.Fatalf("padOrTruncate truncate = %q, want abc", got)
	}
	if got := padOrTruncate("a", 3); got != "a  " {
		t.Fatalf("padOrTruncate pad = %q, want padded", got)
	}
	if got := truncateRunes("ábc", 2); got != "áb" {
		t.Fatalf("truncateRunes() = %q, want áb", got)
	}
}
