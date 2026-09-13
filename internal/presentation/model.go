// Package presentation contains the model for the TUI application.
package presentation

import (
	"context"
	"regexp"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jzes/reqman/internal/domain/request"
	presentationcommandpanel "github.com/jzes/reqman/internal/presentation/commandpanel"
	"github.com/jzes/reqman/internal/presentation/jsoneditor"
	presentationrequest "github.com/jzes/reqman/internal/presentation/request"
)

const (
	sidebarContentWidth  = 25
	methodContentWidth   = 14
	doButtonContentWidth = 10
	topBarHeight         = 1
	defaultScreenWidth   = 100
	defaultScreenHeight  = 30
	focusMarker          = "❯ "
	commandPrompt        = "❯"
	defaultPurple        = "#6272A4"
	focusedPurple        = "#C084FC"
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

type responseTab int

const (
	responseTabStats responseTab = iota
	responseTabBody
	responseTabHeaders
	responseTabRaw
	responseTabCount
)

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
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color(focusedPurple))
	topBarStyle         = lipgloss.NewStyle().
				Background(lipgloss.Color("#C084FC")).
				Foreground(lipgloss.Color("#1F1235"))
	sidebarStyle = lipgloss.NewStyle().
			Width(25).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(defaultPurple)).
			Align(lipgloss.Left)
	inputStyle = lipgloss.NewStyle().
			Width(25).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(defaultPurple))
	helpPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#3B82F6"))
)

type Screen struct {
	url                  urlPanel
	body                 jsoneditor.Panel
	methodList           list.Model
	methodSelectorOpen   bool
	requestPaths         []string
	requests             []request.Request
	loadedRequests       map[int]bool
	selectedRequestIndex int
	requestScrollOffset  int
	headersEditor        headersEditor
	focusedPanel         focusedPanel
	insertMode           bool
	width                int
	height               int
	requestWriter        presentationrequest.RequestWriter
	requestDoer          presentationrequest.RequestDoer
	requestLoader        presentationrequest.RequestLoader
	statusSpinner        spinner.Model
	commandPanel         presentationcommandpanel.Panel
	newRequestPanel      presentationcommandpanel.Panel
	helpOpen             bool
	response             request.Response
	hasResponse          bool
	responseError        string
	requestInFlight      bool
	requestCancel        context.CancelFunc
	nextRequestID        int
	inFlightRequestID    int
	selectedResponseTab  responseTab
}

func NewScreen(requestPaths []string, rw presentationrequest.RequestWriter, rd presentationrequest.RequestDoer, rl presentationrequest.RequestLoader) Screen {
	screen := Screen{
		requestPaths:   requestPaths,
		requests:       make([]request.Request, len(requestPaths)),
		loadedRequests: make(map[int]bool),
		headersEditor:  newHeadersEditor(),
		url:            newURLPanel(),
		body:           jsoneditor.NewPanel(),
		methodList:     newMethodList(),
		statusSpinner:  spinner.New(spinner.WithSpinner(spinner.Line)),
		requestWriter:  rw,
		requestDoer:    rd,
		requestLoader:  rl,
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
		return m.updateWindowSize(msg)
	case presentationrequest.RequestResultMessage:
		if msg.RequestID != 0 && msg.RequestID != m.inFlightRequestID {
			return m, nil
		}
		msg.UpdateTarget(&m)
		m.requestCancel = nil
		m.inFlightRequestID = 0
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
