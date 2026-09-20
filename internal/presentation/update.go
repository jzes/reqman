package presentation

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/jzes/reqman/internal/domain/request"
	presentationcommandpanel "github.com/jzes/reqman/internal/presentation/commandpanel"
	"github.com/jzes/reqman/internal/presentation/jsoneditor"
)

func (scr Screen) updateWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	scr.width = msg.Width
	scr.height = msg.Height
	scr.ensureSelectedRequestVisible()
	return scr, nil
}

func (scr Screen) processKeyMessage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyCtrlC {
		return scr, tea.Quit
	}
	if scr.fatalError {
		return scr, tea.Quit
	}
	if scr.helpOpen {
		if msg.Type == tea.KeyEsc {
			scr.helpOpen = false
		}
		return scr, nil
	}
	if scr.commandPanel.Open {
		action := scr.commandPanel.HandleKey(msg)
		switch action {
		case presentationcommandpanel.ActionQuit:
			return scr, tea.Quit
		case presentationcommandpanel.ActionWrite:
			scr.writeSelectedRequest()
		case presentationcommandpanel.ActionRun:
			if !scr.requestInFlight {
				if cmd := scr.startRequest(); cmd != nil {
					return scr, tea.Batch(cmd, scr.statusSpinner.Tick)
				}
			}
		case presentationcommandpanel.ActionWriteRun:
			if !scr.writeSelectedRequest() {
				return scr, nil
			}
			if !scr.requestInFlight {
				if cmd := scr.startRequest(); cmd != nil {
					return scr, tea.Batch(cmd, scr.statusSpinner.Tick)
				}
			}
		case presentationcommandpanel.ActionWriteQuit:
			if !scr.writeSelectedRequest() {
				return scr, nil
			}
			return scr, tea.Quit
		case presentationcommandpanel.ActionHelp:
			scr.helpOpen = true
		}
		return scr, nil
	}
	if scr.newRequestPanel.Open {
		return scr.processNewRequestPanel(msg)
	}
	if msg.Type == tea.KeyEsc && scr.requestInFlight {
		scr.cancelRequest()
		return scr, nil
	}
	if msg.Type == tea.KeyEsc {
		return scr.processEscapeKey()
	}

	if scr.insertMode {
		return scr.processInsertMode(msg)
	}

	return scr.processNormalMode(msg)
}

func (scr Screen) processEscapeKey() (tea.Model, tea.Cmd) {
	if scr.focusedPanel == focusedPanelURL {
		result := scr.url.HandleEscape()
		if result == urlEscapeIgnored {
			return scr, nil
		}
		scr.syncURLToSelectedRequest()
		if result == urlEscapeToPanel {
			scr.insertMode = false
		}
		return scr, nil
	}
	if scr.focusedPanel == focusedPanelBody {
		result := scr.body.HandleEscape()
		if result == jsoneditor.EscapeIgnored {
			return scr, nil
		}
		scr.syncBodyToSelectedRequest()
		if result == jsoneditor.EscapeToPanel {
			scr.insertMode = false
		}
		return scr, nil
	}
	if scr.focusedPanel == focusedPanelHeaders {
		scr.syncHeadersToSelectedRequest()
	}
	scr.insertMode = false
	scr.methodSelectorOpen = false
	scr.url.Blur()
	scr.body.Blur()
	return scr, nil
}

func (scr Screen) processNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == ':' {
		scr.commandPanel.Activate(commandPrompt)
		return scr, nil
	}

	if scr.methodSelectorOpen {
		return scr.processMethodSelector(msg)
	}

	if msg.Type != tea.KeyRunes || len(msg.Runes) != 1 {
		if msg.Type == tea.KeyEnter || msg.Type == tea.KeySpace {
			scr.methodSelectorOpen = scr.focusedPanel == focusedPanelMethod
			if scr.focusedPanel == focusedPanelDoButton && !scr.requestInFlight {
				if cmd := scr.startRequest(); cmd != nil {
					return scr, tea.Batch(cmd, scr.statusSpinner.Tick)
				}
			}
		}
		return scr, nil
	}

	switch msg.Runes[0] {
	case 'a':
		if scr.focusedPanel == focusedPanelRequests {
			scr.newRequestPanel.ActivateWithTitle("New Request", "")
		}
	case '[':
		if scr.focusedPanel == focusedPanelResponse {
			scr.responsePanel.SelectPreviousTab()
		}
	case ']':
		if scr.focusedPanel == focusedPanelResponse {
			scr.responsePanel.SelectNextTab()
		}
	case 'h':
		scr.focusLeftPanel()
	case 'l':
		scr.focusRightPanel()
	case 'i':
		if scr.focusedPanel == focusedPanelURL {
			scr.insertMode = true
			cmd := scr.url.EnterNavigationMode()
			return scr, cmd
		}
		if scr.focusedPanel == focusedPanelBody {
			scr.insertMode = true
			cmd := scr.body.EnterNavigationMode()
			return scr, cmd
		}
		if scr.focusedPanel != focusedPanelRequests && scr.focusedPanel != focusedPanelMethod && scr.focusedPanel != focusedPanelDoButton {
			scr.insertMode = true
			if scr.focusedPanel == focusedPanelHeaders {
				scr.ensureEditableHeaderRow()
			}
		}
	case 'j':
		if scr.focusedPanel == focusedPanelRequests {
			scr.selectNextRequest()
		} else {
			scr.focusLowerPanel()
		}
	case 'k':
		if scr.focusedPanel == focusedPanelRequests {
			scr.selectPreviousRequest()
		} else {
			scr.focusUpperPanel()
		}
	case 'm':
		scr.methodSelectorOpen = scr.focusedPanel == focusedPanelMethod
	}

	return scr, nil
}

func (scr Screen) processNewRequestPanel(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		scr.newRequestPanel.Close()
		scr.focusedPanel = focusedPanelRequests
	case tea.KeyEnter:
		name := scr.newRequestPanel.Input
		scr.newRequestPanel.Close()
		scr.createRequest(name)
	case tea.KeyRunes:
		scr.newRequestPanel.Input += string(msg.Runes)
	case tea.KeyBackspace:
		scr.newRequestPanel.Input, _ = removeLastRune(scr.newRequestPanel.Input)
	}

	return scr, nil
}

func (scr Screen) processMethodSelector(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyEnter {
		if method, ok := scr.methodList.SelectedItem().(methodItem); ok {
			scr.setSelectedRequestMethod(request.Method(method))
		}
		scr.methodSelectorOpen = false
		return scr, nil
	}

	var cmd tea.Cmd
	scr.methodList, cmd = scr.methodList.Update(msg)
	return scr, cmd
}

func (scr Screen) processInsertMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch scr.focusedPanel {
	case focusedPanelURL:
		switch scr.url.Mode() {
		case urlPanelModeNavigate:
			cmd, _ := scr.url.UpdateNavigationMode(msg)
			return scr, cmd
		case urlPanelModeInsert:
			cmd, changed := scr.url.UpdateEditor(msg)
			if changed {
				scr.syncURLToSelectedRequest()
			}
			return scr, cmd
		}
	case focusedPanelHeaders:
		if scr.headersEditor.Update(msg) {
			scr.syncHeadersToSelectedRequest()
		}
	case focusedPanelBody:
		switch scr.body.Mode() {
		case jsoneditor.NavigateMode:
			cmd, _ := scr.body.UpdateNavigationMode(msg)
			return scr, cmd
		case jsoneditor.InsertMode:
			cmd, changed := scr.body.UpdateEditor(msg)
			if changed {
				scr.syncBodyToSelectedRequest()
			}
			return scr, cmd
		}
	case focusedPanelResponse:
		switch msg.Type {
		case tea.KeyTab:
			scr.responsePanel.SelectNextTab()
		case tea.KeyShiftTab:
			scr.responsePanel.SelectPreviousTab()
		}
	}

	return scr, nil
}
