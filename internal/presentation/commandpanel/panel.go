package commandpanel

import (
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/jzes/reqman/internal/presentation/theme"
)

var commandTheme = theme.Terminal()

type Action int

const (
	ActionNone Action = iota
	ActionQuit
	ActionWrite
	ActionRun
	ActionWriteRun
	ActionWriteQuit
	ActionHelp
	ActionAdd
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
		fields := strings.Fields(p.Input)
		command := ""
		if len(fields) > 0 {
			command = fields[0]
		}
		switch command {
		case "❯q":
			p.Close()
			return ActionQuit
		case "❯w":
			p.Close()
			return ActionWrite
		case "❯r":
			p.Close()
			return ActionRun
		case "❯wr":
			p.Close()
			return ActionWriteRun
		case "❯wq":
			p.Close()
			return ActionWriteQuit
		case "❯?":
			p.Close()
			return ActionHelp
		case "❯a":
			p.Close()
			return ActionAdd
		default:
			p.Close()
		}
	case tea.KeyRunes:
		p.Input += string(msg.Runes)
	case tea.KeySpace:
		p.Input += " "
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
	block := lipgloss.JoinVertical(lipgloss.Left, panel, renderCommandHelp(lipgloss.Width(panel)))
	panelLines := strings.Split(block, "\n")
	baseLines := strings.Split(baseView, "\n")
	insertAt := 4
	leftOffset := max(0, (screenWidth-lipgloss.Width(block))/2)

	for len(baseLines) < insertAt+len(panelLines) {
		baseLines = append(baseLines, "")
	}
	for i, line := range panelLines {
		baseLines[insertAt+i] = overlayLine(baseLines[insertAt+i], line, leftOffset)
	}
	return strings.Join(baseLines, "\n")
}

func renderCommandHelp(width int) string {
	if width <= 0 {
		return ""
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		renderCommandHelpTitle(width),
		renderCommandHelpTable(width),
	)
}

func renderCommandHelpTitle(width int) string {
	const (
		leftCap  = ""
		rightCap = ""
	)
	label := " Commands "
	contentWidth := width - lipgloss.Width(leftCap) - lipgloss.Width(rightCap)
	if contentWidth <= 0 {
		return commandHelpTitleStyle(commandTitleColor()).Render(truncateRunes(label, width))
	}

	leftPadding := max(0, (contentWidth-lipgloss.Width(label))/2)
	rightPadding := max(0, contentWidth-leftPadding-lipgloss.Width(label))
	content := strings.Repeat(" ", leftPadding) + label + strings.Repeat(" ", rightPadding)
	return commandHelpCapStyle(commandTitleColor()).Render(leftCap) + renderCommandHelpGradient(content, contentWidth) + commandHelpCapStyle(commandTitleColor()).Render(rightCap)
}

func renderCommandHelpTable(width int) string {
	rows := []struct {
		command     string
		description string
	}{
		{"a", "New request"},
		{"w [name]", "Save request"},
		{"r", "Run request"},
		{"q", "Quit"},
		{"?", "Open help"},
		{"wr", "Save and run"},
		{"wq", "Save and quit"},
	}

	commandColumnWidth := 8
	contentWidth := max(0, width-commandHelpTableStyle.GetHorizontalFrameSize())
	descriptionColumnWidth := max(0, contentWidth-commandColumnWidth-3)
	lines := make([]string, 0, len(rows)+1)
	lines = append(lines, padOrTruncate("Command", commandColumnWidth)+" | "+padOrTruncate("Action", descriptionColumnWidth))
	for _, row := range rows {
		lines = append(lines, padOrTruncate(row.command, commandColumnWidth)+" | "+padOrTruncate(row.description, descriptionColumnWidth))
	}

	return commandHelpTableStyle.Width(contentWidth).Render(strings.Join(lines, "\n"))
}

func renderCommandHelpGradient(text string, width int) string {
	var builder strings.Builder
	column := 0
	for _, r := range padOrTruncate(text, width) {
		builder.WriteString(commandHelpTitleStyle(commandTitleColor()).Render(string(r)))
		column += lipgloss.Width(string(r))
	}
	return builder.String()
}

func commandTitleColor() lipgloss.Color {
	return theme.Color(commandTheme.Command)
}

func commandHelpTitleStyle(color lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().Background(color).Foreground(theme.Color(commandTheme.CommandText))
}

func commandHelpCapStyle(color lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(color)
}

var (
	commandPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(theme.Color(commandTheme.Command))
	commandHelpTableStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(theme.Color(commandTheme.Command))
	focusedStyle        = lipgloss.NewStyle().Foreground(theme.Color(commandTheme.Focused))
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
