package response

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/jzes/reqman/internal/domain/request"
	"github.com/jzes/reqman/internal/presentation/theme"
)

var responseTestTheme = theme.Terminal()

func TestTabSelection(t *testing.T) {
	panel := NewPanel()
	if panel.SelectedTab() != TabStats {
		t.Fatalf("initial tab = %v, want stats", panel.SelectedTab())
	}

	panel.SelectNextTab()
	if panel.SelectedTab() != TabBody {
		t.Fatalf("next tab = %v, want body", panel.SelectedTab())
	}
	panel.SelectPreviousTab()
	if panel.SelectedTab() != TabStats {
		t.Fatalf("previous tab = %v, want stats", panel.SelectedTab())
	}
	panel.SelectPreviousTab()
	if panel.SelectedTab() != TabRaw {
		t.Fatalf("wrapped previous tab = %v, want raw", panel.SelectedTab())
	}
	panel.SelectTab(TabHeaders)
	panel.SelectTab(Tab(-1))
	if panel.SelectedTab() != TabHeaders {
		t.Fatalf("invalid SelectTab changed tab to %v", panel.SelectedTab())
	}
}

func TestViewRendersTabsAndStats(t *testing.T) {
	view := NewPanel().View(request.Response{Status: "200 OK", StatusCode: 200, Duration: time.Second}, 60, testOptions())
	for _, want := range []string{"Stats", "Body", "Headers", "Raw", "Status: 200 OK", "Duration: 1s"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view does not contain %q: %q", want, view)
		}
	}
}

func TestBodyTabFormatsJSONAndFallsBackToRaw(t *testing.T) {
	panel := NewPanel()
	panel.SelectTab(TabBody)

	view := panel.View(request.Response{Body: `{"ok":true}`}, 60, testOptions())
	if !strings.Contains(view, "  \"ok\": true") {
		t.Fatalf("body view does not contain formatted JSON: %q", view)
	}

	view = panel.View(request.Response{Body: `{"bad":`}, 60, testOptions())
	if !strings.Contains(view, `{"bad":`) {
		t.Fatalf("body view does not contain raw invalid JSON: %q", view)
	}

	view = panel.View(request.Response{}, 60, testOptions())
	if !strings.Contains(view, "(empty)") {
		t.Fatalf("body view does not contain empty placeholder: %q", view)
	}
}

func TestHeadersAndRawTabs(t *testing.T) {
	url, err := request.NewURL("http://localhost/books")
	if err != nil {
		t.Fatalf("NewURL() error = %v", err)
	}
	res := request.Response{
		URL:        url,
		Status:     "201 Created",
		StatusCode: 201,
		Headers: request.NewHeadersFrom(map[string][]string{
			"Content-Type": {"application/json"},
			"Accept":       {"application/json", "text/plain"},
		}),
		Body: `{"ok":true}`,
	}

	panel := NewPanel()
	panel.SelectTab(TabHeaders)
	view := panel.View(res, 60, testOptions())
	if !strings.Contains(view, "Accept") || !strings.Contains(view, "application/json, text/plain") {
		t.Fatalf("headers view missing joined headers: %q", view)
	}

	panel.SelectTab(TabRaw)
	view = panel.View(res, 60, testOptions())
	for _, want := range []string{"URL: http://localhost/books", "Status: 201 Created", "Headers:", "Body:"} {
		if !strings.Contains(view, want) {
			t.Fatalf("raw view missing %q: %q", want, view)
		}
	}
}

func TestRawTabRendersEmptyHeadersAndBody(t *testing.T) {
	panel := NewPanel()
	panel.SelectTab(TabRaw)

	view := panel.View(request.Response{}, 40, testOptions())
	if !strings.Contains(view, "Headers:") || !strings.Contains(view, "(empty)") {
		t.Fatalf("raw empty view = %q, want empty placeholders", view)
	}
}

func TestResponseBodyWrapPreservesContent(t *testing.T) {
	for _, body := range []string{
		`{"message":"` + strings.Repeat("abcdefghij", 12) + `END"}`,
		strings.Repeat("long text with spaces  ", 6) + "END",
		"first line\n\n  " + strings.Repeat("abc", 30) + "\nlast line",
		strings.Repeat("日本語🙂e\u0301", 20) + "END",
		"",
	} {
		for _, tab := range []Tab{TabBody, TabRaw} {
			panel := NewPanel()
			panel.SelectTab(tab)
			for _, width := range []int{20, 40, 80} {
				content := ansi.Strip(panel.renderSelectedTabContent(request.Response{Body: body}, width, testOptions()))
				lines := strings.Split(content, "\n")
				var displayed strings.Builder
				inBody := tab == TabBody
				for _, line := range lines[1 : len(lines)-1] {
					if got := lipgloss.Width(line); got != width {
						t.Fatalf("tab %v: line width = %d, want %d: %q", tab, got, width, line)
					}
					text := strings.TrimSuffix(strings.TrimPrefix(line, "│ "), " │")
					if inBody {
						displayed.WriteString(text)
					} else if strings.TrimSpace(text) == "Body:" {
						inBody = true
					}
				}
				want := body
				if tab == TabBody {
					want = formatJSONBody(body)
				} else if body == "" {
					want = "(empty)"
				}
				// Ignore layout padding and line breaks; every non-whitespace
				// character (including the end of a long line) must survive.
				compact := func(s string) string { return strings.Join(strings.Fields(s), "") }
				if got := compact(displayed.String()); got != compact(want) {
					t.Fatalf("tab %v, width %d: displayed %q, want %q", tab, width, got, compact(want))
				}
			}
		}
	}
}

func TestBodyWrapPreservesWhitespaceAndExistingLines(t *testing.T) {
	panel := NewPanel()
	panel.SelectTab(TabBody)
	body := "abc  defgh\n\n  ijklmnop"
	got := panel.renderSelectedTab(request.Response{Body: body}, 9)
	want := "abc  \ndefgh\n\n  ijk\nlmnop"
	if got != want {
		t.Fatalf("wrapped body = %q, want %q", got, want)
	}
}

func testOptions() ViewOptions {
	return ViewOptions{FocusedColor: responseTestTheme.Focused, DefaultColor: responseTestTheme.DefaultBorder}
}
