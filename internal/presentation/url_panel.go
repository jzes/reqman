package presentation

import (
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type urlPanelMode int

type urlEscapeResult int

const (
	urlPanelModePanel urlPanelMode = iota
	urlPanelModeNavigate
	urlPanelModeInsert
)

const (
	urlEscapeIgnored urlEscapeResult = iota
	urlEscapeToNavigation
	urlEscapeToPanel
)

type urlPanel struct {
	input textinput.Model
	mode  urlPanelMode
}

func newURLPanel() urlPanel {
	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = "http://localhost"
	input.Cursor.SetMode(cursor.CursorStatic)
	input.Blur()
	return urlPanel{input: input}
}

func (p urlPanel) Mode() urlPanelMode {
	return p.mode
}

func (p urlPanel) Value() string {
	return p.input.Value()
}

func (p *urlPanel) SetValue(value string) {
	p.input.SetValue(value)
}

func (p *urlPanel) Blur() {
	p.input.Blur()
}

func (p *urlPanel) EnterNavigationMode() tea.Cmd {
	p.mode = urlPanelModeNavigate
	p.syncCursorMode()
	p.input.CursorStart()
	return p.input.Focus()
}

func (p *urlPanel) EnterInsertMode() tea.Cmd {
	p.mode = urlPanelModeInsert
	p.syncCursorMode()
	p.input.CursorEnd()
	return p.input.Focus()
}

func (p *urlPanel) HandleEscape() urlEscapeResult {
	switch p.mode {
	case urlPanelModeInsert:
		p.mode = urlPanelModeNavigate
		p.syncCursorMode()
		return urlEscapeToNavigation
	case urlPanelModeNavigate:
		p.mode = urlPanelModePanel
		p.syncCursorMode()
		p.Blur()
		return urlEscapeToPanel
	}

	return urlEscapeIgnored
}

func (p *urlPanel) UpdateEditor(msg tea.KeyMsg) (tea.Cmd, bool) {
	before := p.input.Value()
	var cmd tea.Cmd
	p.input, cmd = p.input.Update(msg)
	return cmd, before != p.input.Value()
}

func (p *urlPanel) UpdateNavigationMode(msg tea.KeyMsg) (tea.Cmd, bool) {
	if msg.Type != tea.KeyRunes || len(msg.Runes) != 1 {
		return nil, false
	}

	switch msg.Runes[0] {
	case 'i':
		return p.enterInsertMode(), true
	case 'a':
		p.moveCursor(1)
		return p.enterInsertMode(), true
	case 'A':
		p.input.CursorEnd()
		return p.enterInsertMode(), true
	case 'h':
		p.moveCursor(-1)
		return nil, true
	case 'l':
		p.moveCursor(1)
		return nil, true
	case 'w':
		p.moveCursorToNextWordStart()
		return nil, true
	case 'b':
		p.moveCursorToPreviousWordStart()
		return nil, true
	case 'e':
		p.moveCursorToWordEnd()
		return nil, true
	}

	return nil, false
}

func (p urlPanel) View(width int, focused bool) string {
	input := p.input
	input.Width = width
	if focused && p.mode != urlPanelModePanel {
		input.Focus()
	} else {
		input.Blur()
	}
	if focused && p.mode == urlPanelModeInsert {
		input.Cursor.SetMode(cursor.CursorHide)
		return p.viewWithBarCursor(input.View())
	}
	return input.View()
}

func (p urlPanel) Style(style lipgloss.Style, focused bool) lipgloss.Style {
	if !focused {
		return style
	}
	switch p.mode {
	case urlPanelModeNavigate:
		return style.BorderForeground(lipgloss.Color("#50FA7B"))
	case urlPanelModeInsert:
		return style.BorderForeground(lipgloss.Color("#FFB86C"))
	default:
		return style.BorderForeground(lipgloss.Color("99"))
	}
}

func (p urlPanel) CursorMode() cursor.Mode {
	return p.input.Cursor.Mode()
}

func (p *urlPanel) syncCursorMode() {
	mode := cursor.CursorStatic
	if p.mode == urlPanelModeInsert {
		mode = cursor.CursorBlink
	}
	p.input.Cursor.SetMode(mode)
}

func (p *urlPanel) enterInsertMode() tea.Cmd {
	p.mode = urlPanelModeInsert
	p.syncCursorMode()
	return p.input.Focus()
}

func (p urlPanel) viewWithBarCursor(view string) string {
	return replaceAtURLColumn(view, p.input.Position(), "|")
}

func replaceAtURLColumn(line string, column int, value string) string {
	if column < 0 {
		column = 0
	}

	for i, visibleColumn := 0, 0; i < len(line); {
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

func (p *urlPanel) moveCursor(delta int) {
	p.input.SetCursor(p.input.Position() + delta)
}

func (p *urlPanel) moveCursorToNextWordStart() {
	runes := []rune(p.input.Value())
	pos := clampRuneIndex(p.input.Position(), len(runes))
	for pos < len(runes) && isURLWordRune(runes[pos]) {
		pos++
	}
	for pos < len(runes) && !isURLWordRune(runes[pos]) {
		pos++
	}
	p.input.SetCursor(pos)
}

func (p *urlPanel) moveCursorToPreviousWordStart() {
	runes := []rune(p.input.Value())
	pos := clampRuneIndex(p.input.Position(), len(runes)) - 1
	for pos > 0 && !isURLWordRune(runes[pos]) {
		pos--
	}
	for pos > 0 && isURLWordRune(runes[pos-1]) {
		pos--
	}
	p.input.SetCursor(max(0, pos))
}

func (p *urlPanel) moveCursorToWordEnd() {
	runes := []rune(p.input.Value())
	pos := clampRuneIndex(p.input.Position(), len(runes))
	if pos >= len(runes) {
		return
	}
	for pos < len(runes) && !isURLWordRune(runes[pos]) {
		pos++
	}
	for pos < len(runes)-1 && isURLWordRune(runes[pos+1]) {
		pos++
	}
	p.input.SetCursor(pos + 1)
}

func clampRuneIndex(index, length int) int {
	if index < 0 {
		return 0
	}
	if index > length {
		return length
	}
	return index
}

func isURLWordRune(r rune) bool {
	return r == '_' || r == '-' || r == '.' || unicode.IsLetter(r) || unicode.IsDigit(r) || utf8.RuneLen(r) > 1
}
