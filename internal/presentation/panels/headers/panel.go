package headers

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Row struct {
	Key   string
	Value string
}

type cell int

const (
	cellKey cell = iota
	cellValue
)

type Panel struct {
	rows         []Row
	selectedRow  int
	selectedCell cell
}

type ViewOptions struct {
	Focused     bool
	InsertMode  bool
	FocusMarker string
	FocusStyle  lipgloss.Style
}

func NewPanel() Panel {
	return Panel{}
}

func (p *Panel) SetRows(rows []Row) {
	p.rows = rows
	p.clampSelection()
}

func (p Panel) Rows() []Row {
	return append([]Row(nil), p.rows...)
}

func (p *Panel) EnsureEditableRow() {
	if len(p.rows) == 0 {
		p.rows = []Row{{}}
		p.selectedRow = 0
		p.selectedCell = cellKey
	}
}

func (p *Panel) Update(msg tea.KeyMsg) bool {
	if len(p.rows) == 0 {
		p.EnsureEditableRow()
	}

	mutated := false
	switch msg.Type {
	case tea.KeyRunes:
		p.appendToSelectedCell(string(msg.Runes))
		mutated = true
	case tea.KeyBackspace:
		mutated = p.backspaceSelectedCell()
	case tea.KeyTab, tea.KeyRight:
		p.selectedCell = cellValue
	case tea.KeyShiftTab, tea.KeyLeft:
		p.selectedCell = cellKey
	case tea.KeyEnter:
		if p.selectedCell == cellKey {
			p.selectedCell = cellValue
		} else {
			p.insertRowAfterSelection()
			mutated = true
		}
	case tea.KeyUp:
		p.selectPreviousRow()
	case tea.KeyDown:
		p.selectNextRow()
	case tea.KeyDelete, tea.KeyCtrlD:
		mutated = p.deleteSelectedRow()
	}
	return mutated
}

func (p Panel) View(headerColumnWidth, valueColumnWidth int, opts ViewOptions) string {
	var builder strings.Builder
	builder.WriteString(padOrTruncate("Header", headerColumnWidth))
	builder.WriteString("  ")
	builder.WriteString(padOrTruncate("Value", valueColumnWidth))

	if len(p.rows) == 0 {
		builder.WriteString("\n  (empty)")
		return builder.String()
	}

	for i, row := range p.rows {
		selected := opts.Focused && i == p.selectedRow
		key := row.Key
		value := row.Value
		if selected && opts.InsertMode {
			if p.selectedCell == cellKey {
				key = textWithCursor(key, opts.FocusStyle)
			} else {
				value = textWithCursor(value, opts.FocusStyle)
			}
		}

		keyCell := padOrTruncate(key, headerColumnWidth)
		valueCell := padOrTruncate(value, valueColumnWidth)
		if selected {
			if p.selectedCell == cellKey {
				keyCell = opts.FocusStyle.Render(keyCell)
			} else {
				valueCell = opts.FocusStyle.Render(valueCell)
			}
		}

		prefix := "  "
		if selected {
			prefix = opts.FocusStyle.Render(opts.FocusMarker)
		}
		builder.WriteString("\n")
		builder.WriteString(prefix)
		builder.WriteString(keyCell)
		builder.WriteString("  ")
		builder.WriteString(valueCell)
	}

	return builder.String()
}

func (p *Panel) clampSelection() {
	if p.selectedRow < 0 {
		p.selectedRow = 0
	}
	if len(p.rows) == 0 {
		p.selectedRow = 0
		p.selectedCell = cellKey
		return
	}
	if p.selectedRow >= len(p.rows) {
		p.selectedRow = len(p.rows) - 1
	}
	if p.selectedCell != cellValue {
		p.selectedCell = cellKey
	}
}

func (p *Panel) appendToSelectedCell(text string) {
	if len(p.rows) == 0 {
		return
	}
	row := &p.rows[p.selectedRow]
	if p.selectedCell == cellValue {
		row.Value += text
		return
	}
	row.Key += text
}

func (p *Panel) backspaceSelectedCell() bool {
	if len(p.rows) == 0 {
		return false
	}
	row := &p.rows[p.selectedRow]
	if p.selectedCell == cellValue {
		newValue, changed := removeLastRune(row.Value)
		row.Value = newValue
		return changed
	}
	newKey, changed := removeLastRune(row.Key)
	row.Key = newKey
	return changed
}

func (p *Panel) insertRowAfterSelection() {
	newRowIndex := p.selectedRow + 1
	p.rows = append(p.rows, Row{})
	copy(p.rows[newRowIndex+1:], p.rows[newRowIndex:])
	p.rows[newRowIndex] = Row{}
	p.selectedRow = newRowIndex
	p.selectedCell = cellKey
}

func (p *Panel) deleteSelectedRow() bool {
	if len(p.rows) == 0 {
		return false
	}
	p.rows = append(p.rows[:p.selectedRow], p.rows[p.selectedRow+1:]...)
	p.clampSelection()
	return true
}

func (p *Panel) selectPreviousRow() {
	if p.selectedRow > 0 {
		p.selectedRow--
	}
}

func (p *Panel) selectNextRow() {
	if p.selectedRow < len(p.rows)-1 {
		p.selectedRow++
	}
}

func removeLastRune(text string) (string, bool) {
	runes := []rune(text)
	if len(runes) == 0 {
		return text, false
	}
	return string(runes[:len(runes)-1]), true
}

func textWithCursor(text string, style lipgloss.Style) string {
	return text + style.Render("█")
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
