package commandpanel

import (
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type Action int

const (
	ActionNone Action = iota
	ActionQuit
	ActionWrite
	ActionRun
)

type Panel struct {
	Open  bool
	Input string
	Title string
}

func (p *Panel) Activate(input string) {
	p.ActivateWithTitle("Command", input)
}

func (p *Panel) ActivateWithTitle(title string, input string) {
	p.Open = true
	p.Input = input
	p.Title = title
}

func (p *Panel) Close() {
	p.Open = false
	p.Input = ""
	p.Title = ""
}

func (p *Panel) HandleKey(msg tea.KeyMsg) Action {
	switch msg.Type {
	case tea.KeyEsc:
		p.Close()
	case tea.KeyEnter:
		switch p.Input {
		case ":q":
			p.Close()
			return ActionQuit
		case ":w":
			p.Close()
			return ActionWrite
		case ":r":
			p.Close()
			return ActionRun
		default:
			p.Close()
		}
	case tea.KeyRunes:
		p.Input += string(msg.Runes)
	case tea.KeyBackspace:
		p.Input, _ = removeLastRune(p.Input)
		if p.Input == "" {
			p.Close()
		}
	}

	return ActionNone
}

func (p Panel) Render(baseView string, screenWidth int) string {
	if !p.Open {
		return baseView
	}

	panelWidth := min(50, max(20, screenWidth/2))
	title := p.Title
	if title == "" {
		title = "Command"
	}
	panel := renderPanelWithTitle(
		commandPanelStyle.Width(panelWidth),
		title,
		textWithCursor(p.Input, true),
	)
	panelLines := strings.Split(panel, "\n")
	baseLines := strings.Split(baseView, "\n")
	insertAt := 4
	leftOffset := max(0, (screenWidth-lipgloss.Width(panel))/2)

	for len(baseLines) < insertAt+len(panelLines) {
		baseLines = append(baseLines, "")
	}
	for i, line := range panelLines {
		baseLines[insertAt+i] = overlayLine(baseLines[insertAt+i], line, leftOffset)
	}
	return strings.Join(baseLines, "\n")
}

var (
	commandPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#50FA7B"))
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	ansiPrefixPattern   = regexp.MustCompile(`^(?:\x1b\[[0-9;]*m)+`)
	ansiSuffixPattern   = regexp.MustCompile(`(?:\x1b\[[0-9;]*m)+$`)
	ansiSequencePattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)
)

func textWithCursor(text string, showCursor bool) string {
	if !showCursor {
		return text
	}
	return text + focusedStyle.Render("█")
}

func renderPanelWithTitle(style lipgloss.Style, title string, content string) string {
	rendered := style.Render(content)
	lines := strings.Split(rendered, "\n")
	if len(lines) == 0 {
		return rendered
	}

	titleText := " " + title + " "
	titleWidth := lipgloss.Width(titleText)
	prefix := ansiPrefixPattern.FindString(lines[0])
	suffix := ansiSuffixPattern.FindString(lines[0])
	visibleTopBorder := ansiSequencePattern.ReplaceAllString(lines[0], "")
	topBorder := []rune(visibleTopBorder)
	if len(topBorder) <= titleWidth+1 {
		return rendered
	}

	lines[0] = prefix + string(topBorder[:1]) + titleText + string(topBorder[1+titleWidth:]) + suffix
	return strings.Join(lines, "\n")
}

func overlayLine(baseLine, overlay string, leftOffset int) string {
	overlayWidth := lipgloss.Width(overlay)
	left := ansi.Cut(baseLine, 0, leftOffset)
	right := ansi.Cut(baseLine, leftOffset+overlayWidth, lipgloss.Width(baseLine))

	if leftWidth := lipgloss.Width(left); leftWidth < leftOffset {
		left += strings.Repeat(" ", leftOffset-leftWidth)
	}

	return left + overlay + right
}

func removeLastRune(text string) (string, bool) {
	if text == "" {
		return "", false
	}
	last := 0
	for i := range text {
		last = i
	}
	if last == 0 {
		return "", true
	}
	return text[:last], true
}
