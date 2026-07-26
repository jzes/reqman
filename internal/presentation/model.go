// Package presentation contains the model for the TUI application.
package presentation

import (
	"path/filepath"
	"regexp"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jzes/reqman/internal/domain/request"
	presentationbody "github.com/jzes/reqman/internal/presentation/body"
	presentationcommandpanel "github.com/jzes/reqman/internal/presentation/commandpanel"
	presentationrequest "github.com/jzes/reqman/internal/presentation/request"
)

const (
	sidebarContentWidth  = 25
	methodContentWidth   = 14
	doButtonContentWidth = 10
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
)

type Screen struct {
	url                  urlPanel
	body                 presentationbody.Panel
	methodList           list.Model
	methodSelectorOpen   bool
	requests             []request.Request
	selectedRequestIndex int
	headersEditor        headersEditor
	focusedPanel         focusedPanel
	insertMode           bool
	width                int
	height               int
	requestWriter        presentationrequest.RequestWriter
	requestDirectory     string
	requestDoer          presentationrequest.RequestDoer
	statusSpinner        spinner.Model
	commandPanel         presentationcommandpanel.Panel
	newRequestPanel      presentationcommandpanel.Panel
	response             request.Response
	hasResponse          bool
	responseError        string
	requestInFlight      bool
}

func NewScreen(requests []request.Request, rw presentationrequest.RequestWriter, rd presentationrequest.RequestDoer) Screen {
	return NewScreenWithDirectory(requests, inferRequestDirectory(requests), rw, rd)
}

func NewScreenWithDirectory(requests []request.Request, requestDirectory string, rw presentationrequest.RequestWriter, rd presentationrequest.RequestDoer) Screen {
	if requestDirectory == "" {
		requestDirectory = "."
	}

	screen := Screen{
		requests:         requests,
		headersEditor:    newHeadersEditor(),
		url:              newURLPanel(),
		body:             presentationbody.NewPanel(),
		methodList:       newMethodList(),
		statusSpinner:    spinner.New(spinner.WithSpinner(spinner.Line)),
		requestWriter:    rw,
		requestDirectory: requestDirectory,
		requestDoer:      rd,
	}

	screen.showSelectedRequestURL()
	screen.showSelectedRequestHeaders()
	screen.showSelectedRequestBody()
	screen.showSelectedRequestMethod()
	return screen
}

func inferRequestDirectory(requests []request.Request) string {
	for _, req := range requests {
		if req.Path != "" {
			return filepath.Dir(req.Path)
		}
	}
	return "."
}

func (m Screen) Init() tea.Cmd {
	return nil
}

func (m Screen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.updateWindowSize(msg)
	case presentationrequest.RequestResultMessage:
		msg.UpdateTarget(&m)
		return m, nil
	case spinner.TickMsg:
		if !m.requestInFlight {
			return m, nil
		}

		var cmd tea.Cmd
		m.statusSpinner, cmd = m.statusSpinner.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		return m.processKeyMessage(msg)
	}
	return m, nil
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

func (m *Screen) SetRequestInFlight(inFlight bool) {
	m.requestInFlight = inFlight
}

func (m *Screen) SetResponseError(err string) {
	m.responseError = err
}

func (m *Screen) SetResponse(response request.Response) {
	m.response = response
}

func (m *Screen) SetHasResponse(hasResponse bool) {
	m.hasResponse = hasResponse
}
