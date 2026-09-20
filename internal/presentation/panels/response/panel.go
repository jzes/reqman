package response

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/jzes/reqman/internal/domain/request"
)

type Tab int

const (
	TabStats Tab = iota
	TabBody
	TabHeaders
	TabRaw
	tabCount
)

type Panel struct {
	selectedTab Tab
}

type ViewOptions struct {
	FocusedColor string
	DefaultColor string
}

func NewPanel() Panel {
	return Panel{}
}

func (p Panel) SelectedTab() Tab {
	return p.selectedTab
}

func (p *Panel) SelectTab(tab Tab) {
	if tab < 0 || tab >= tabCount {
		return
	}
	p.selectedTab = tab
}

func (p *Panel) SelectPreviousTab() {
	if p.selectedTab == TabStats {
		p.selectedTab = tabCount - 1
		return
	}
	p.selectedTab--
}

func (p *Panel) SelectNextTab() {
	p.selectedTab = (p.selectedTab + 1) % tabCount
}

func (p Panel) View(response request.Response, width int, opts ViewOptions) string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		p.renderConnectedTabs(opts),
		p.renderSelectedTabContent(response, width, opts),
	)
}

func (p Panel) renderConnectedTabs(opts ViewOptions) string {
	lines := strings.Split(p.renderTabs(opts), "\n")
	if len(lines) <= 1 {
		return strings.Join(lines, "\n")
	}
	return strings.Join(lines[:len(lines)-1], "\n")
}

func (p Panel) renderTabs(opts ViewOptions) string {
	tabs := []struct {
		tab   Tab
		label string
	}{
		{TabStats, "Stats"},
		{TabBody, "Body"},
		{TabHeaders, "Headers"},
		{TabRaw, "Raw"},
	}

	activeBorder := lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "│",
		BottomRight: "│",
	}
	activeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(opts.FocusedColor)).
		Border(activeBorder).
		BorderForeground(lipgloss.Color(opts.FocusedColor)).
		Padding(0, 1)
	inactiveStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(opts.DefaultColor)).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(opts.DefaultColor)).
		Padding(0, 1)

	renderedTabs := make([]string, 0, len(tabs))
	for _, tab := range tabs {
		label := tab.label
		if p.selectedTab == tab.tab {
			renderedTabs = append(renderedTabs, activeStyle.Render(label))
			continue
		}
		renderedTabs = append(renderedTabs, inactiveStyle.Render(label))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
}

func (p Panel) renderSelectedTabContent(response request.Response, width int, opts ViewOptions) string {
	content := p.renderSelectedTab(response, width)
	outerWidth := max(4, width)
	contentWidth := max(0, outerWidth-4)
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(opts.FocusedColor))

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = borderStyle.Render("│") + " " + padOrTruncate(line, contentWidth) + " " + borderStyle.Render("│")
	}

	parts := make([]string, 0, len(lines)+2)
	parts = append(parts, borderStyle.Render(p.renderTabContentTopBorder(outerWidth)))
	parts = append(parts, lines...)
	parts = append(parts, borderStyle.Render("╰"+strings.Repeat("─", max(0, outerWidth-2))+"╯"))
	return strings.Join(parts, "\n")
}

func (p Panel) renderTabContentTopBorder(width int) string {
	activeTabLeft, activeTabWidth := p.selectedTabPosition()
	activeTabRight := activeTabLeft + activeTabWidth - 1

	var builder strings.Builder
	for column := 0; column < width; column++ {
		char := "─"
		if column == 0 {
			char = "╭"
		}
		if column == width-1 {
			char = "╮"
		}

		if column >= activeTabLeft && column <= activeTabRight {
			switch column {
			case activeTabLeft:
				char = "╯"
				if column == 0 {
					char = "│"
				}
			case activeTabRight:
				char = "╰"
				if column == width-1 {
					char = "│"
				}
			default:
				char = " "
			}
		}

		builder.WriteString(char)
	}

	return builder.String()
}

func (p Panel) selectedTabPosition() (int, int) {
	tabs := []struct {
		tab   Tab
		label string
	}{
		{TabStats, "Stats"},
		{TabBody, "Body"},
		{TabHeaders, "Headers"},
		{TabRaw, "Raw"},
	}

	offset := 0
	for _, tab := range tabs {
		width := lipgloss.Width(tab.label) + 4
		if p.selectedTab == tab.tab {
			return offset, width
		}
		offset += width
	}

	return 0, lipgloss.Width("Stats") + 4
}

func (p Panel) renderSelectedTab(response request.Response, width int) string {
	switch p.selectedTab {
	case TabBody:
		return formatJSONBody(response.Body)
	case TabHeaders:
		return renderHeaders(response, width)
	case TabRaw:
		return renderRaw(response)
	default:
		return renderStats(response)
	}
}

func renderStats(response request.Response) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Status: %s\n", response.Status)
	fmt.Fprintf(&builder, "Status Code: %d\n", response.StatusCode)
	fmt.Fprintf(&builder, "Duration: %s", response.Duration)
	return builder.String()
}

func renderHeaders(response request.Response, width int) string {
	headerColumnWidth := min(25, max(10, width/3))
	valueColumnWidth := max(10, width-headerColumnWidth-2)

	var builder strings.Builder
	builder.WriteString(padOrTruncate("Header", headerColumnWidth))
	builder.WriteString("  ")
	builder.WriteString(padOrTruncate("Value", valueColumnWidth))

	if response.Headers.Len() == 0 {
		builder.WriteString("\n  (empty)")
		return builder.String()
	}

	for _, key := range response.Headers.Names() {
		builder.WriteString("\n")
		builder.WriteString(padOrTruncate(key, headerColumnWidth))
		builder.WriteString("  ")
		builder.WriteString(padOrTruncate(strings.Join(response.Headers.Values(key), ", "), valueColumnWidth))
	}

	return builder.String()
}

func renderRaw(response request.Response) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "URL: %s\n", response.URL.String())
	fmt.Fprintf(&builder, "Status: %s\n", response.Status)
	fmt.Fprintf(&builder, "Status Code: %d\n", response.StatusCode)
	fmt.Fprintf(&builder, "Duration: %s\n", response.Duration)

	if response.Headers.Len() == 0 {
		builder.WriteString("Headers:\n  (empty)\n")
	} else {
		builder.WriteString("Headers:\n")
		for _, key := range response.Headers.Names() {
			fmt.Fprintf(&builder, "  %s: %s\n", key, strings.Join(response.Headers.Values(key), ", "))
		}
	}

	builder.WriteString("Body:\n")
	if response.Body == "" {
		builder.WriteString("  (empty)")
	} else {
		builder.WriteString(response.Body)
	}

	return strings.TrimRight(builder.String(), "\n")
}

func formatJSONBody(body string) string {
	if strings.TrimSpace(body) == "" {
		return "(empty)"
	}

	var formatted bytes.Buffer
	if err := json.Indent(&formatted, []byte(body), "", "  "); err != nil {
		return body
	}
	return formatted.String()
}

func padOrTruncate(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(text) > width {
		return truncateRunes(text, width)
	}
	return text + strings.Repeat(" ", width-lipgloss.Width(text))
}

func truncateRunes(text string, width int) string {
	if width <= 0 {
		return ""
	}
	var builder strings.Builder
	for _, r := range text {
		candidate := builder.String() + string(r)
		if lipgloss.Width(candidate) > width {
			break
		}
		builder.WriteRune(r)
	}
	return builder.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
