// Package presentation contains the model for the TUI application.
package presentation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/jzes/reqman/internal/application"
	"github.com/jzes/reqman/internal/domain/request"
)

const (
	sidebarContentWidth  = 25
	methodContentWidth   = 14
	doButtonContentWidth = 8
	defaultScreenWidth   = 100
	defaultScreenHeight  = 30
)

type focusedPanel int

const (
	focusedPanelRequests focusedPanel = iota
	focusedPanelMethod
	focusedPanelURL
	focusedPanelDoButton
	focusedPanelHeaders
	focusedPanelBody
	focusedPanelResponse
)

type headerCell int

const (
	headerCellKey headerCell = iota
	headerCellValue
)

type bodyMode int

const (
	bodyModePanel bodyMode = iota
	bodyModeNavigate
	bodyModeInsert
)

type headerRow struct {
	key   string
	value string
}

type headersEditor struct {
	rows         []headerRow
	selectedRow  int
	selectedCell headerCell
}

var (
	ansiPrefixPattern   = regexp.MustCompile(`^(?:\x1b\[[0-9;]*m)+`)
	ansiSuffixPattern   = regexp.MustCompile(`(?:\x1b\[[0-9;]*m)+$`)
	ansiSequencePattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)
	titleStyle          = lipgloss.NewStyle().Bold(true)
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	sidebarStyle        = lipgloss.NewStyle().
				Width(25).
				Border(lipgloss.RoundedBorder()).
				Align(lipgloss.Left)
	inputStyle = lipgloss.NewStyle().
			Width(25).
			Border(lipgloss.RoundedBorder())
	commandPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#50FA7B"))
)

type requestResultMsg struct {
	response request.Response
	err      error
}

type Screen struct {
	textInput            string
	bodyTextArea         textarea.Model
	methodList           list.Model
	methodSelectorOpen   bool
	requests             []request.Request
	selectedRequestIndex int
	headersEditor        headersEditor
	bodyMode             bodyMode
	focusedPanel         focusedPanel
	insertMode           bool
	width                int
	height               int
	requestWriter        application.RequestWriter
	hasRequestWriter     bool
	requestDoer          application.RequestDoer
	hasRequestDoer       bool
	commandPanelOpen     bool
	commandInput         string
	response             request.Response
	hasResponse          bool
	responseError        string
	requestInFlight      bool
}

func NewScreen(requests []request.Request, dependencies ...any) Screen {
	screen := Screen{
		requests:      requests,
		headersEditor: newHeadersEditor(),
		bodyTextArea:  newBodyTextArea(),
		methodList:    newMethodList(),
	}
	for _, dependency := range dependencies {
		switch dependency := dependency.(type) {
		case application.RequestWriter:
			screen.requestWriter = dependency
			screen.hasRequestWriter = true
		case application.RequestDoer:
			screen.requestDoer = dependency
			screen.hasRequestDoer = true
		}
	}
	screen.showSelectedRequestURL()
	screen.showSelectedRequestHeaders()
	screen.showSelectedRequestBody()
	screen.showSelectedRequestMethod()
	return screen
}

func (m Screen) Init() tea.Cmd {
	return nil
}

func (m Screen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case requestResultMsg:
		m.requestInFlight = false
		if msg.err != nil {
			m.responseError = msg.err.Error()
			m.hasResponse = false
			m.response = request.Response{}
			return m, nil
		}

		m.responseError = ""
		m.response = msg.response
		m.hasResponse = true
		return m, nil
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		if m.commandPanelOpen {
			return m.updateCommandPanel(msg)
		}
		if msg.Type == tea.KeyEsc {
			if m.focusedPanel == focusedPanelBody {
				switch m.bodyMode {
				case bodyModeInsert:
					m.formatBodyJSONWithJQ()
					m.syncBodyToSelectedRequest()
					m.bodyMode = bodyModeNavigate
					return m, nil
				case bodyModeNavigate:
					m.formatBodyJSONWithJQ()
					m.syncBodyToSelectedRequest()
					m.bodyMode = bodyModePanel
					m.insertMode = false
					m.bodyTextArea.Blur()
					return m, nil
				}
			}
			if m.focusedPanel == focusedPanelHeaders {
				m.syncHeadersToSelectedRequest()
			}
			if m.focusedPanel == focusedPanelURL {
				m.syncURLToSelectedRequest()
			}
			m.insertMode = false
			m.methodSelectorOpen = false
			m.bodyTextArea.Blur()
			return m, nil
		}

		if m.insertMode {
			return m.updateInsertMode(msg)
		}

		return m.updateNormalMode(msg)
	}
	return m, nil
}

func (m Screen) updateNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == ':' {
		m.commandPanelOpen = true
		m.commandInput = ":"
		return m, nil
	}

	if m.methodSelectorOpen {
		return m.updateMethodSelector(msg)
	}

	if msg.Type != tea.KeyRunes || len(msg.Runes) != 1 {
		if msg.Type == tea.KeyEnter || msg.Type == tea.KeySpace {
			if m.focusedPanel == focusedPanelDoButton {
				if m.requestInFlight {
					return m, nil
				}
				if cmd := m.requestCommand(); cmd != nil {
					m.requestInFlight = true
					return m, cmd
				}
				return m, nil
			}
			m.methodSelectorOpen = m.focusedPanel == focusedPanelMethod
		}
		return m, nil
	}

	switch msg.Runes[0] {
	case 'h':
		m.focusLeftPanel()
	case 'l':
		m.focusRightPanel()
	case 'i':
		if m.focusedPanel == focusedPanelBody {
			m.insertMode = true
			m.bodyMode = bodyModeNavigate
			cmd := m.bodyTextArea.Focus()
			m.moveBodyCursorToStart()
			return m, cmd
		}
		if m.focusedPanel != focusedPanelRequests && m.focusedPanel != focusedPanelMethod {
			m.insertMode = true
			if m.focusedPanel == focusedPanelHeaders {
				m.ensureEditableHeaderRow()
			}
		}
	case 'j':
		if m.focusedPanel == focusedPanelRequests {
			m.selectNextRequest()
		} else {
			m.focusLowerPanel()
		}
	case 'k':
		if m.focusedPanel == focusedPanelRequests {
			m.selectPreviousRequest()
		} else {
			m.focusUpperPanel()
		}
	case 'm':
		m.methodSelectorOpen = m.focusedPanel == focusedPanelMethod
	}

	return m, nil
}

func (m Screen) updateMethodSelector(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyEnter {
		if method, ok := m.methodList.SelectedItem().(methodItem); ok {
			m.setSelectedRequestMethod(request.Method(method))
		}
		m.methodSelectorOpen = false
		return m, nil
	}

	var cmd tea.Cmd
	m.methodList, cmd = m.methodList.Update(msg)
	return m, cmd
}

func (m Screen) updateCommandPanel(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.commandPanelOpen = false
		m.commandInput = ""
	case tea.KeyEnter:
		switch m.commandInput {
		case ":q":
			return m, tea.Quit
		case ":w":
			m.writeSelectedRequest()
		case ":r":
			if !m.requestInFlight {
				if cmd := m.requestCommand(); cmd != nil {
					m.requestInFlight = true
					m.commandPanelOpen = false
					m.commandInput = ""
					return m, cmd
				}
			}
		}
		m.commandPanelOpen = false
		m.commandInput = ""
	case tea.KeyRunes:
		m.commandInput += string(msg.Runes)
	case tea.KeyBackspace:
		m.commandInput, _ = removeLastRune(m.commandInput)
		if m.commandInput == "" {
			m.commandPanelOpen = false
		}
	}

	return m, nil
}

func (m Screen) updateInsertMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.focusedPanel {
	case focusedPanelURL:
		switch msg.Type {
		case tea.KeyRunes:
			m.textInput += string(msg.Runes)
		case tea.KeyBackspace:
			if len(m.textInput) > 0 {
				m.textInput = m.textInput[:len(m.textInput)-1]
			}
		}
	case focusedPanelHeaders:
		if m.updateHeadersEditor(msg) {
			m.syncHeadersToSelectedRequest()
		}
	case focusedPanelBody:
		switch m.bodyMode {
		case bodyModeNavigate:
			cmd, _ := m.updateBodyNavigationMode(msg)
			return m, cmd
		case bodyModeInsert:
			cmd, changed := m.updateBodyEditor(msg)
			if changed {
				m.syncBodyToSelectedRequest()
			}
			return m, cmd
		}
	}

	return m, nil
}

func (m *Screen) updateBodyEditor(msg tea.KeyMsg) (tea.Cmd, bool) {
	if msg.Type == tea.KeyCtrlF {
		return nil, m.formatBodyJSON()
	}

	if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 {
		switch msg.Runes[0] {
		case '{':
			m.bodyTextArea.InsertString("{}")
			return m.updateBodyTextAreaWithKey(tea.KeyMsg{Type: tea.KeyLeft}), true
		case '[':
			m.bodyTextArea.InsertString("[]")
			return m.updateBodyTextAreaWithKey(tea.KeyMsg{Type: tea.KeyLeft}), true
		}
	}

	var cmd tea.Cmd
	m.bodyTextArea, cmd = m.bodyTextArea.Update(msg)
	return cmd, true
}

func (m *Screen) updateBodyNavigationMode(msg tea.KeyMsg) (tea.Cmd, bool) {
	if msg.Type != tea.KeyRunes || len(msg.Runes) != 1 {
		return nil, false
	}

	switch msg.Runes[0] {
	case 'i':
		m.bodyMode = bodyModeInsert
		return m.bodyTextArea.Focus(), true
	case 'h':
		return m.updateBodyTextAreaWithKey(tea.KeyMsg{Type: tea.KeyLeft}), true
	case 'j':
		return m.updateBodyTextAreaWithKey(tea.KeyMsg{Type: tea.KeyDown}), true
	case 'k':
		return m.updateBodyTextAreaWithKey(tea.KeyMsg{Type: tea.KeyUp}), true
	case 'l':
		return m.updateBodyTextAreaWithKey(tea.KeyMsg{Type: tea.KeyRight}), true
	case 'w':
		m.moveBodyCursorToNextWordStart()
		return nil, true
	case 'b':
		return m.updateBodyTextAreaWithKey(tea.KeyMsg{Type: tea.KeyLeft, Alt: true}), true
	case 'e':
		return m.moveBodyCursorToWordEnd(), true
	}

	return nil, false
}

func (m *Screen) moveBodyCursorToWordEnd() tea.Cmd {
	for !m.cursorAtEndOfBody() && m.bodyCursorIsOnWhitespace() {
		m.updateBodyTextAreaWithKey(tea.KeyMsg{Type: tea.KeyRight})
	}
	for !m.cursorAtEndOfBody() && !m.bodyCursorIsOnWhitespace() {
		m.updateBodyTextAreaWithKey(tea.KeyMsg{Type: tea.KeyRight})
	}
	return nil
}

func (m *Screen) moveBodyCursorToNextWordStart() {
	m.updateBodyTextAreaWithKey(tea.KeyMsg{Type: tea.KeyRight, Alt: true})
	for !m.cursorAtEndOfBody() && !m.bodyCursorIsOnWhitespace() {
		m.updateBodyTextAreaWithKey(tea.KeyMsg{Type: tea.KeyRight})
	}
	for !m.cursorAtEndOfBody() && m.bodyCursorIsOnWhitespace() {
		m.updateBodyTextAreaWithKey(tea.KeyMsg{Type: tea.KeyRight})
	}
}

func (m *Screen) moveBodyCursorToStart() {
	m.updateBodyTextAreaWithKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'<'}, Alt: true})
}

func (m Screen) bodyCursorIsOnWhitespace() bool {
	line := m.currentBodyLineRunes()
	column := m.bodyTextArea.LineInfo().StartColumn + m.bodyTextArea.LineInfo().ColumnOffset
	if column < 0 || column >= len(line) {
		return true
	}
	return unicode.IsSpace(line[column])
}

func (m Screen) cursorAtEndOfBody() bool {
	line := m.currentBodyLineRunes()
	lineInfo := m.bodyTextArea.LineInfo()
	return m.bodyTextArea.Line() == m.bodyTextArea.LineCount()-1 && lineInfo.StartColumn+lineInfo.ColumnOffset >= len(line)
}

func (m Screen) currentBodyLineRunes() []rune {
	lines := strings.Split(m.bodyTextArea.Value(), "\n")
	line := m.bodyTextArea.Line()
	if line < 0 || line >= len(lines) {
		return nil
	}
	return []rune(lines[line])
}

func (m *Screen) formatBodyJSON() bool {
	raw := strings.TrimSpace(m.bodyTextArea.Value())
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

	m.bodyTextArea.SetValue(string(formatted))
	return true
}

func (m *Screen) formatBodyJSONWithJQ() bool {
	raw := strings.TrimSpace(m.bodyTextArea.Value())
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
	m.bodyTextArea.SetValue(formatted)
	return true
}

func (m *Screen) updateBodyTextAreaWithKey(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	m.bodyTextArea, cmd = m.bodyTextArea.Update(msg)
	return cmd
}

func (m *Screen) selectNextRequest() {
	if m.selectedRequestIndex < len(m.requests)-1 {
		m.selectedRequestIndex++
		m.showSelectedRequestURL()
		m.showSelectedRequestHeaders()
		m.showSelectedRequestBody()
		m.showSelectedRequestMethod()
		m.clearResponse()
	}
}

func (m *Screen) selectPreviousRequest() {
	if m.selectedRequestIndex > 0 {
		m.selectedRequestIndex--
		m.showSelectedRequestURL()
		m.showSelectedRequestHeaders()
		m.showSelectedRequestBody()
		m.showSelectedRequestMethod()
		m.clearResponse()
	}
}

func (m *Screen) clearResponse() {
	m.hasResponse = false
	m.responseError = ""
	m.response = request.Response{}
}

func (m *Screen) focusLeftPanel() {
	switch m.focusedPanel {
	case focusedPanelURL:
		m.focusedPanel = focusedPanelMethod
	case focusedPanelDoButton:
		m.focusedPanel = focusedPanelURL
	case focusedPanelMethod, focusedPanelHeaders, focusedPanelBody, focusedPanelResponse:
		m.focusedPanel = focusedPanelRequests
	}
}

func (m *Screen) focusRightPanel() {
	if m.focusedPanel == focusedPanelRequests {
		m.focusedPanel = focusedPanelMethod
		return
	}

	if m.focusedPanel == focusedPanelMethod {
		m.focusedPanel = focusedPanelURL
		return
	}

	if m.focusedPanel == focusedPanelURL {
		m.focusedPanel = focusedPanelDoButton
	}
}

func (m *Screen) focusLowerPanel() {
	switch m.focusedPanel {
	case focusedPanelMethod, focusedPanelURL, focusedPanelDoButton:
		m.focusedPanel = focusedPanelHeaders
	case focusedPanelHeaders:
		m.focusedPanel = focusedPanelBody
	case focusedPanelBody:
		m.focusedPanel = focusedPanelResponse
	}
}

func (m *Screen) focusUpperPanel() {
	switch m.focusedPanel {
	case focusedPanelResponse:
		m.focusedPanel = focusedPanelBody
	case focusedPanelBody:
		m.focusedPanel = focusedPanelHeaders
	case focusedPanelHeaders:
		m.focusedPanel = focusedPanelMethod
	}
}

func (m *Screen) showSelectedRequestURL() {
	if len(m.requests) == 0 {
		m.textInput = ""
		return
	}

	m.textInput = m.requests[m.selectedRequestIndex].URL.String()
}

func (m *Screen) showSelectedRequestHeaders() {
	if len(m.requests) == 0 {
		m.headersEditor.setRows(nil)
		return
	}

	m.headersEditor.setRows(headersFromMap(m.requests[m.selectedRequestIndex].Headers))
}

func headersFromMap(headers map[string]string) []headerRow {
	rows := make([]headerRow, 0, len(headers))
	for key, value := range headers {
		rows = append(rows, headerRow{key: key, value: value})
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].key < rows[j].key
	})
	return rows
}

func (e *headersEditor) setRows(rows []headerRow) {
	e.rows = rows
	e.clampSelection()
}

func (e *headersEditor) clampSelection() {
	if e.selectedRow < 0 {
		e.selectedRow = 0
	}
	if len(e.rows) == 0 {
		e.selectedRow = 0
		e.selectedCell = headerCellKey
		return
	}
	if e.selectedRow >= len(e.rows) {
		e.selectedRow = len(e.rows) - 1
	}
	if e.selectedCell != headerCellValue {
		e.selectedCell = headerCellKey
	}
}

func (m *Screen) showSelectedRequestBody() {
	if len(m.requests) == 0 {
		m.bodyTextArea.SetValue("")
		return
	}

	m.bodyTextArea.SetValue(m.requests[m.selectedRequestIndex].Body)
}

func (m *Screen) showSelectedRequestMethod() {
	if len(m.requests) == 0 {
		m.methodList.Select(0)
		return
	}

	selectedMethod := m.requests[m.selectedRequestIndex].Method
	for i, method := range httpMethodItems() {
		if request.Method(method.(methodItem)) == selectedMethod {
			m.methodList.Select(i)
			return
		}
	}
}

func (m *Screen) setSelectedRequestMethod(method request.Method) {
	if len(m.requests) == 0 {
		return
	}

	m.requests[m.selectedRequestIndex].Method = method
}

func newHeadersEditor() headersEditor {
	return headersEditor{}
}

func (m *Screen) ensureEditableHeaderRow() {
	if len(m.headersEditor.rows) == 0 {
		m.headersEditor.rows = []headerRow{{}}
		m.headersEditor.selectedRow = 0
		m.headersEditor.selectedCell = headerCellKey
	}
}

func (m *Screen) updateHeadersEditor(msg tea.KeyMsg) bool {
	if len(m.headersEditor.rows) == 0 {
		m.ensureEditableHeaderRow()
	}

	mutated := false
	switch msg.Type {
	case tea.KeyRunes:
		m.headersEditor.appendToSelectedCell(string(msg.Runes))
		mutated = true
	case tea.KeyBackspace:
		mutated = m.headersEditor.backspaceSelectedCell()
	case tea.KeyTab, tea.KeyRight:
		m.headersEditor.selectedCell = headerCellValue
	case tea.KeyShiftTab, tea.KeyLeft:
		m.headersEditor.selectedCell = headerCellKey
	case tea.KeyEnter:
		if m.headersEditor.selectedCell == headerCellKey {
			m.headersEditor.selectedCell = headerCellValue
		} else {
			m.headersEditor.insertRowAfterSelection()
			mutated = true
		}
	case tea.KeyUp:
		m.headersEditor.selectPreviousRow()
	case tea.KeyDown:
		m.headersEditor.selectNextRow()
	case tea.KeyDelete, tea.KeyCtrlD:
		mutated = m.headersEditor.deleteSelectedRow()
	}
	return mutated
}

func (e *headersEditor) appendToSelectedCell(text string) {
	if len(e.rows) == 0 {
		return
	}
	row := &e.rows[e.selectedRow]
	if e.selectedCell == headerCellValue {
		row.value += text
		return
	}
	row.key += text
}

func (e *headersEditor) backspaceSelectedCell() bool {
	if len(e.rows) == 0 {
		return false
	}
	row := &e.rows[e.selectedRow]
	if e.selectedCell == headerCellValue {
		newValue, changed := removeLastRune(row.value)
		row.value = newValue
		return changed
	}
	newKey, changed := removeLastRune(row.key)
	row.key = newKey
	return changed
}

func removeLastRune(text string) (string, bool) {
	if text == "" {
		return text, false
	}
	_, size := utf8.DecodeLastRuneInString(text)
	return text[:len(text)-size], true
}

func (e *headersEditor) insertRowAfterSelection() {
	newRowIndex := e.selectedRow + 1
	e.rows = append(e.rows, headerRow{})
	copy(e.rows[newRowIndex+1:], e.rows[newRowIndex:])
	e.rows[newRowIndex] = headerRow{}
	e.selectedRow = newRowIndex
	e.selectedCell = headerCellKey
}

func (e *headersEditor) deleteSelectedRow() bool {
	if len(e.rows) == 0 {
		return false
	}
	e.rows = append(e.rows[:e.selectedRow], e.rows[e.selectedRow+1:]...)
	e.clampSelection()
	return true
}

func (e *headersEditor) selectPreviousRow() {
	if e.selectedRow > 0 {
		e.selectedRow--
	}
}

func (e *headersEditor) selectNextRow() {
	if e.selectedRow < len(e.rows)-1 {
		e.selectedRow++
	}
}

func (m *Screen) syncBodyToSelectedRequest() {
	if len(m.requests) == 0 {
		return
	}

	m.requests[m.selectedRequestIndex].Body = m.bodyTextArea.Value()
}

func (m *Screen) syncURLToSelectedRequest() {
	if len(m.requests) == 0 {
		return
	}

	url, err := request.NewURL(m.textInput)
	if err != nil {
		return
	}
	m.requests[m.selectedRequestIndex].URL = url
}

func (m *Screen) writeSelectedRequest() {
	if !m.hasRequestWriter || len(m.requests) == 0 {
		return
	}

	_ = m.requestWriter.WriteToFile(m.requests[m.selectedRequestIndex])
}

func (m Screen) requestCommand() tea.Cmd {
	if !m.hasRequestDoer || len(m.requests) == 0 {
		return nil
	}

	requestToDo := m.requests[m.selectedRequestIndex]
	return func() tea.Msg {
		response, err := m.requestDoer.Do(requestToDo)
		return requestResultMsg{response: response, err: err}
	}
}

func (m *Screen) syncHeadersToSelectedRequest() {
	if len(m.requests) == 0 {
		return
	}

	headers := make(map[string]string)
	for _, row := range m.headersEditor.rows {
		if row.key == "" {
			continue
		}
		headers[row.key] = row.value
	}
	m.requests[m.selectedRequestIndex].Headers = headers
}

func newMethodList() list.Model {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.SetHeight(1)
	delegate.SetSpacing(0)

	methods := list.New(httpMethodItems(), delegate, 14, 7)
	methods.SetShowTitle(false)
	methods.SetShowStatusBar(false)
	methods.SetShowPagination(false)
	methods.SetShowHelp(false)
	methods.SetFilteringEnabled(false)
	methods.DisableQuitKeybindings()
	return methods
}

func httpMethodItems() []list.Item {
	return []list.Item{
		methodItem(request.MethodGet),
		methodItem(request.MethodPost),
		methodItem(request.MethodPut),
		methodItem(request.MethodPatch),
		methodItem(request.MethodDelete),
		methodItem(request.MethodHead),
		methodItem(request.MethodOptions),
	}
}

type methodItem request.Method

func (i methodItem) FilterValue() string {
	return string(i)
}

func (i methodItem) Title() string {
	return string(i)
}

func (i methodItem) Description() string {
	return ""
}

func newBodyTextArea() textarea.Model {
	body := textarea.New()
	body.Placeholder = "(empty)"
	body.Prompt = ""
	body.ShowLineNumbers = false
	body.Blur()
	return body
}

func (m Screen) View() string {
	screenWidth := m.width
	if screenWidth == 0 {
		screenWidth = defaultScreenWidth
	}
	screenHeight := m.height
	if screenHeight == 0 {
		screenHeight = defaultScreenHeight
	}

	var sidebarBuilder strings.Builder
	if len(m.requests) == 0 {
		sidebarBuilder.WriteString("  (empty)\n")
	} else {
		for i, req := range m.requests {
			displayReq := req.Name
			if len(displayReq) > 28 {
				displayReq = displayReq[:25] + "..."
			}

			prefix := "  "
			if i == m.selectedRequestIndex {
				prefix = focusedStyle.Render("> ")
			}
			fmt.Fprintf(&sidebarBuilder, "%s%d. %s\n", prefix, i+1, displayReq)
		}
	}

	sidebarPanelStyle := styleForPanel(sidebarStyle, m.focusedPanel == focusedPanelRequests, false)
	sidebar := renderPanelWithTitle(
		sidebarPanelStyle.
			Width(sidebarContentWidth).
			Height(max(0, screenHeight-sidebarStyle.GetVerticalFrameSize())),
		"Requests",
		sidebarBuilder.String(),
	)

	detailsWidth := max(0, screenWidth-lipgloss.Width(sidebar))
	panelContentWidth := max(0, detailsWidth-inputStyle.GetHorizontalFrameSize())
	panelStyle := inputStyle.Width(panelContentWidth)
	methodStyle := styleForPanel(inputStyle.Width(methodContentWidth), m.focusedPanel == focusedPanelMethod, m.methodSelectorOpen)
	headersStyle := styleForPanel(panelStyle, m.focusedPanel == focusedPanelHeaders, m.insertMode)
	bodyStyle := styleForBodyPanel(panelStyle, m.focusedPanel == focusedPanelBody, m.bodyMode)

	methodPanel := renderPanelWithTitle(methodStyle, "Method", focusedStyle.Render("> ")+m.renderMethodSelector())
	doButton := m.renderDoButton()
	urlContentWidth := max(0, detailsWidth-lipgloss.Width(methodPanel)-lipgloss.Width(doButton)-inputStyle.GetHorizontalFrameSize())
	urlStyle := styleForPanel(inputStyle.Width(urlContentWidth), m.focusedPanel == focusedPanelURL, m.insertMode)
	requestURL := textWithCursor(m.textInput, m.focusedPanel == focusedPanelURL && m.insertMode)
	urlPanel := renderPanelWithTitle(urlStyle, "URL", focusedStyle.Render("> ")+requestURL)
	requestLine := lipgloss.JoinHorizontal(lipgloss.Top, methodPanel, urlPanel, doButton)

	headerColumnWidth := min(25, max(10, panelContentWidth/3))
	valueColumnWidth := max(10, panelContentWidth-headerColumnWidth-3)
	headers := renderPanelWithTitle(headersStyle, "Headers", m.renderHeadersEditor(headerColumnWidth, valueColumnWidth))

	usedHeight := lipgloss.Height(requestLine) + lipgloss.Height(headers)
	remainingPanelHeight := max(2, screenHeight-usedHeight-inputStyle.GetVerticalFrameSize()*2)
	bodyPanelHeight := max(1, remainingPanelHeight/2)
	responsePanelHeight := max(1, remainingPanelHeight-bodyPanelHeight)
	bodyTextArea := m.bodyTextArea
	bodyTextArea.SetWidth(panelContentWidth)
	bodyTextArea.SetHeight(bodyPanelHeight)
	if m.focusedPanel == focusedPanelBody && m.bodyMode != bodyModePanel {
		bodyTextArea.Focus()
	} else {
		bodyTextArea.Blur()
	}
	body := renderPanelWithTitle(bodyStyle.Height(bodyPanelHeight), "Body", bodyTextArea.View())
	responseStyle := styleForPanel(panelStyle, m.focusedPanel == focusedPanelResponse, false)
	response := renderPanelWithTitle(responseStyle.Height(responsePanelHeight), "Response", m.renderResponse())
	details := lipgloss.JoinVertical(lipgloss.Left, requestLine, headers, body, response)

	mainView := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, details)
	if m.commandPanelOpen {
		return m.renderCommandPanel(mainView, screenWidth)
	}
	return mainView
}

func (m Screen) renderResponse() string {
	if m.requestInFlight {
		return "Executing request..."
	}

	if m.responseError != "" {
		return "Error:\n" + m.responseError
	}

	if !m.hasResponse {
		return "(empty)"
	}

	var builder strings.Builder
	fmt.Fprintf(&builder, "URL: %s\n", m.response.URL.String())
	fmt.Fprintf(&builder, "Status: %s\n", m.response.Status)
	fmt.Fprintf(&builder, "Status Code: %d\n", m.response.StatusCode)
	fmt.Fprintf(&builder, "Duration: %s\n", m.response.Duration)

	if len(m.response.Headers) == 0 {
		builder.WriteString("Headers:\n  (empty)\n")
	} else {
		builder.WriteString("Headers:\n")
		keys := make([]string, 0, len(m.response.Headers))
		for key := range m.response.Headers {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			fmt.Fprintf(&builder, "  %s: %s\n", key, strings.Join(m.response.Headers[key], ", "))
		}
	}

	builder.WriteString("Body:\n")
	if m.response.Body == "" {
		builder.WriteString("  (empty)")
	} else {
		builder.WriteString(m.response.Body)
	}

	return strings.TrimRight(builder.String(), "\n")
}

func (m Screen) renderDoButton() string {
	contentStyle := lipgloss.NewStyle().
		Width(doButtonContentWidth).
		Align(lipgloss.Left)

	panelStyle := inputStyle.Width(doButtonContentWidth + 2)
	if m.focusedPanel == focusedPanelDoButton {
		panelStyle = panelStyle.BorderForeground(lipgloss.Color("99"))
	}
	if m.requestInFlight {
		panelStyle = panelStyle.BorderForeground(lipgloss.Color("#F1FA8C"))
	}

	content := "Do"
	if m.requestInFlight {
		content = "Busy"
	}

	return renderPanelWithTitle(panelStyle, "Do Req", contentStyle.Render(content))
}

func (m Screen) renderCommandPanel(baseView string, screenWidth int) string {
	panelWidth := min(50, max(20, screenWidth/2))
	panel := renderPanelWithTitle(
		commandPanelStyle.Width(panelWidth),
		"Command",
		textWithCursor(m.commandInput, true),
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

func overlayLine(baseLine, overlay string, leftOffset int) string {
	overlayWidth := lipgloss.Width(overlay)
	left := ansi.Cut(baseLine, 0, leftOffset)
	right := ansi.Cut(baseLine, leftOffset+overlayWidth, lipgloss.Width(baseLine))

	if leftWidth := lipgloss.Width(left); leftWidth < leftOffset {
		left += strings.Repeat(" ", leftOffset-leftWidth)
	}

	return left + overlay + right
}

func (m Screen) renderHeadersEditor(headerColumnWidth, valueColumnWidth int) string {
	var builder strings.Builder
	builder.WriteString(padOrTruncate("Header", headerColumnWidth))
	builder.WriteString("  ")
	builder.WriteString(padOrTruncate("Value", valueColumnWidth))

	if len(m.headersEditor.rows) == 0 {
		builder.WriteString("\n  (empty)")
		return builder.String()
	}

	for i, row := range m.headersEditor.rows {
		selected := m.focusedPanel == focusedPanelHeaders && i == m.headersEditor.selectedRow
		key := row.key
		value := row.value
		if selected && m.insertMode {
			if m.headersEditor.selectedCell == headerCellKey {
				key = textWithCursor(key, true)
			} else {
				value = textWithCursor(value, true)
			}
		}

		keyCell := padOrTruncate(key, headerColumnWidth)
		valueCell := padOrTruncate(value, valueColumnWidth)
		if selected {
			if m.headersEditor.selectedCell == headerCellKey {
				keyCell = focusedStyle.Render(keyCell)
			} else {
				valueCell = focusedStyle.Render(valueCell)
			}
		}

		prefix := "  "
		if selected {
			prefix = focusedStyle.Render("> ")
		}
		builder.WriteString("\n")
		builder.WriteString(prefix)
		builder.WriteString(keyCell)
		builder.WriteString("  ")
		builder.WriteString(valueCell)
	}

	return builder.String()
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
func (m Screen) renderMethodSelector() string {
	if m.methodSelectorOpen {
		methodList := m.methodList
		methodList.SetSize(14, 7)
		return methodList.View()
	}

	method := request.MethodGet
	if len(m.requests) > 0 {
		method = m.requests[m.selectedRequestIndex].Method
	}

	selector := fmt.Sprintf("[%s]", method)
	if m.focusedPanel == focusedPanelURL {
		return titleStyle.Render(selector)
	}
	return selector
}

func textWithCursor(text string, showCursor bool) string {
	if !showCursor {
		return text
	}
	return text + focusedStyle.Render("█")
}

func styleForBodyPanel(style lipgloss.Style, focused bool, mode bodyMode) lipgloss.Style {
	if !focused {
		return style
	}
	switch mode {
	case bodyModeNavigate:
		return style.BorderForeground(lipgloss.Color("#50FA7B"))
	case bodyModeInsert:
		return style.BorderForeground(lipgloss.Color("#FFB86C"))
	default:
		return style.BorderForeground(lipgloss.Color("99"))
	}
}

func styleForPanel(style lipgloss.Style, focused bool, insertMode bool) lipgloss.Style {
	if !focused {
		return style
	}
	if insertMode {
		return style.BorderForeground(lipgloss.Color("#50FA7B"))
	}
	return style.BorderForeground(lipgloss.Color("99"))
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
