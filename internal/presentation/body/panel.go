package body

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type Mode int

type EscapeResult int

const (
	PanelMode Mode = iota
	NavigateMode
	InsertMode
)

const (
	EscapeIgnored EscapeResult = iota
	EscapeToNavigation
	EscapeToPanel
)

type Panel struct {
	textArea textarea.Model
	mode     Mode
}

func NewPanel() Panel {
	body := textarea.New()
	body.Placeholder = "(empty)"
	body.Prompt = ""
	body.ShowLineNumbers = false
	body.Cursor.SetMode(cursor.CursorStatic)
	body.Blur()
	return Panel{textArea: body}
}

func (p Panel) Mode() Mode {
	return p.mode
}

func (p *Panel) SetMode(mode Mode) {
	p.mode = mode
	p.syncCursorMode()
}

func (p Panel) Value() string {
	return p.textArea.Value()
}

func (p *Panel) SetValue(value string) {
	p.textArea.SetValue(value)
}

func (p *Panel) Blur() {
	p.textArea.Blur()
}

func (p *Panel) EnterNavigationMode() tea.Cmd {
	p.mode = NavigateMode
	p.syncCursorMode()
	cmd := p.textArea.Focus()
	p.MoveCursorToStart()
	return cmd
}

func (p *Panel) PersistChanges() {
	p.FormatJSONWithJQ()
}

func (p *Panel) HandleEscape() EscapeResult {
	switch p.mode {
	case InsertMode:
		p.PersistChanges()
		p.mode = NavigateMode
		p.syncCursorMode()
		return EscapeToNavigation
	case NavigateMode:
		p.PersistChanges()
		p.mode = PanelMode
		p.syncCursorMode()
		p.Blur()
		return EscapeToPanel
	}

	return EscapeIgnored
}

func (p *Panel) FormatJSON() bool {
	raw := strings.TrimSpace(p.textArea.Value())
	if raw == "" {
		return false
	}

	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return false
	}

	formatted, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return false
	}

	p.textArea.SetValue(string(formatted))
	return true
}

func (p *Panel) FormatJSONWithJQ() bool {
	raw := strings.TrimSpace(p.textArea.Value())
	if raw == "" {
		return false
	}

	cmd := exec.Command("jq", ".")
	cmd.Stdin = strings.NewReader(raw)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return false
	}

	formatted := strings.TrimRight(stdout.String(), "\n")
	p.textArea.SetValue(formatted)
	return true
}

func (p *Panel) UpdateEditor(msg tea.KeyMsg) (tea.Cmd, bool) {
	if msg.Type == tea.KeyCtrlF {
		return nil, p.FormatJSON()
	}

	if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 {
		switch msg.Runes[0] {
		case '{':
			p.textArea.InsertString("{}")
			return p.updateTextAreaWithKey(tea.KeyMsg{Type: tea.KeyLeft}), true
		case '[':
			p.textArea.InsertString("[]")
			return p.updateTextAreaWithKey(tea.KeyMsg{Type: tea.KeyLeft}), true
		}
	}

	var cmd tea.Cmd
	p.textArea, cmd = p.textArea.Update(msg)
	return cmd, true
}

func (p *Panel) UpdateNavigationMode(msg tea.KeyMsg) (tea.Cmd, bool) {
	if msg.Type != tea.KeyRunes || len(msg.Runes) != 1 {
		return nil, false
	}

	switch msg.Runes[0] {
	case 'i':
		return p.enterInsertMode(), true
	case 'a':
		p.updateTextAreaWithKey(tea.KeyMsg{Type: tea.KeyRight})
		return p.enterInsertMode(), true
	case 'A':
		p.textArea.CursorEnd()
		return p.enterInsertMode(), true
	case 'h':
		return p.updateTextAreaWithKey(tea.KeyMsg{Type: tea.KeyLeft}), true
	case 'j':
		return p.updateTextAreaWithKey(tea.KeyMsg{Type: tea.KeyDown}), true
	case 'k':
		return p.updateTextAreaWithKey(tea.KeyMsg{Type: tea.KeyUp}), true
	case 'l':
		return p.updateTextAreaWithKey(tea.KeyMsg{Type: tea.KeyRight}), true
	case 'w':
		p.moveCursorToNextWordStart()
		return nil, true
	case 'b':
		return p.updateTextAreaWithKey(tea.KeyMsg{Type: tea.KeyLeft, Alt: true}), true
	case 'e':
		return p.moveCursorToWordEnd(), true
	}

	return nil, false
}

func (p Panel) View(width, height int, focused bool) string {
	textArea := p.textArea
	textArea.SetWidth(width)
	textArea.SetHeight(height)
	if focused && p.mode != PanelMode {
		textArea.Focus()
	} else {
		textArea.Blur()
	}
	if focused && p.mode == InsertMode {
		textArea.Cursor.SetMode(cursor.CursorHide)
		return p.viewWithBarCursor(textArea.View(), width)
	}
	return textArea.View()
}

func (p Panel) Style(style lipgloss.Style, focused bool) lipgloss.Style {
	if !focused {
		return style
	}
	switch p.mode {
	case NavigateMode:
		return style.BorderForeground(lipgloss.Color("#50FA7B"))
	case InsertMode:
		return style.BorderForeground(lipgloss.Color("#FFB86C"))
	default:
		return style.BorderForeground(lipgloss.Color("99"))
	}
}

func (p Panel) CursorMode() cursor.Mode {
	return p.textArea.Cursor.Mode()
}

func (p *Panel) MoveCursorToStart() {
	p.updateTextAreaWithKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'<'}, Alt: true})
}

func (p *Panel) updateTextAreaWithKey(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	p.textArea, cmd = p.textArea.Update(msg)
	return cmd
}

func (p *Panel) enterInsertMode() tea.Cmd {
	p.mode = InsertMode
	p.syncCursorMode()
	return p.textArea.Focus()
}

func (p *Panel) syncCursorMode() {
	mode := cursor.CursorStatic
	if p.mode == InsertMode {
		mode = cursor.CursorBlink
	}
	p.textArea.Cursor.SetMode(mode)
}

func (p Panel) viewWithBarCursor(view string, width int) string {
	lines := strings.Split(view, "\n")
	if len(lines) == 0 {
		return view
	}

	row := p.cursorDisplayRow(width)
	if row < 0 {
		row = 0
	}
	if row >= len(lines) {
		row = len(lines) - 1
	}

	column := p.textArea.LineInfo().ColumnOffset
	lines[row] = replaceAtVisibleColumn(lines[row], column, "|")
	return strings.Join(lines, "\n")
}

func (p Panel) cursorDisplayRow(width int) int {
	if width <= 0 {
		return 0
	}

	row := 0
	lines := strings.Split(p.textArea.Value(), "\n")
	currentLine := p.textArea.Line()
	for i := 0; i < currentLine && i < len(lines); i++ {
		row += wrappedLineCount(lines[i], width)
	}
	return row + p.textArea.LineInfo().RowOffset
}

func wrappedLineCount(line string, width int) int {
	if width <= 0 {
		return 1
	}
	lineWidth := ansi.StringWidth(line)
	if lineWidth == 0 {
		return 1
	}
	return (lineWidth + width - 1) / width
}

func replaceAtVisibleColumn(line string, column int, value string) string {
	visibleColumn := 0
	for i := 0; i < len(line); {
		if line[i] == '\x1b' {
			i = skipANSISequence(line, i)
			continue
		}

		r, size := utf8.DecodeRuneInString(line[i:])
		if r == utf8.RuneError && size == 0 {
			break
		}
		if visibleColumn >= column {
			return line[:i] + value + line[i+size:]
		}
		visibleColumn += ansi.StringWidth(string(r))
		i += size
	}

	return line + value
}

func skipANSISequence(line string, start int) int {
	if start+1 < len(line) && line[start+1] == '[' {
		for i := start + 2; i < len(line); i++ {
			if line[i] >= '@' && line[i] <= '~' {
				return i + 1
			}
		}
		return len(line)
	}

	for i := start + 1; i < len(line); i++ {
		if line[i] >= '@' && line[i] <= '~' {
			return i + 1
		}
	}
	return len(line)
}

func (p *Panel) moveCursorToWordEnd() tea.Cmd {
	for !p.cursorAtEnd() && p.cursorIsOnWhitespace() {
		p.updateTextAreaWithKey(tea.KeyMsg{Type: tea.KeyRight})
	}
	for !p.cursorAtEnd() && !p.cursorIsOnWhitespace() {
		p.updateTextAreaWithKey(tea.KeyMsg{Type: tea.KeyRight})
	}
	return nil
}

func (p *Panel) moveCursorToNextWordStart() {
	p.updateTextAreaWithKey(tea.KeyMsg{Type: tea.KeyRight, Alt: true})
	for !p.cursorAtEnd() && !p.cursorIsOnWhitespace() {
		p.updateTextAreaWithKey(tea.KeyMsg{Type: tea.KeyRight})
	}
	for !p.cursorAtEnd() && p.cursorIsOnWhitespace() {
		p.updateTextAreaWithKey(tea.KeyMsg{Type: tea.KeyRight})
	}
}

func (p Panel) cursorIsOnWhitespace() bool {
	line := p.currentLineRunes()
	column := p.textArea.LineInfo().StartColumn + p.textArea.LineInfo().ColumnOffset
	if column < 0 || column >= len(line) {
		return true
	}
	return unicode.IsSpace(line[column])
}

func (p Panel) cursorAtEnd() bool {
	line := p.currentLineRunes()
	lineInfo := p.textArea.LineInfo()
	return p.textArea.Line() == p.textArea.LineCount()-1 && lineInfo.StartColumn+lineInfo.ColumnOffset >= len(line)
}

func (p Panel) currentLineRunes() []rune {
	lines := strings.Split(p.textArea.Value(), "\n")
	line := p.textArea.Line()
	if line < 0 || line >= len(lines) {
		return nil
	}
	return []rune(lines[line])
}
