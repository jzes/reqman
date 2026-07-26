package presentation

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/jzes/reqman/internal/domain/request"
	presentationbody "github.com/jzes/reqman/internal/presentation/body"
	presentationcommandpanel "github.com/jzes/reqman/internal/presentation/commandpanel"
)

func (scr Screen) updateWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	scr.width = msg.Width
	scr.height = msg.Height
	return scr, nil
}

func (scr Screen) processKeyMessage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyCtrlC {
		return scr, tea.Quit
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
				if cmd := scr.requestCommand(); cmd != nil {
					scr.requestInFlight = true
					return scr, cmd
				}
			}
		}
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
	if scr.focusedPanel == focusedPanelBody {
		result := scr.body.HandleEscape()
		if result == presentationbody.EscapeIgnored {
			return scr, nil
		}
		scr.syncBodyToSelectedRequest()
		if result == presentationbody.EscapeToPanel {
			scr.insertMode = false
		}
		return scr, nil
	}
	if scr.focusedPanel == focusedPanelHeaders {
		scr.syncHeadersToSelectedRequest()
	}
	if scr.focusedPanel == focusedPanelURL {
		scr.syncURLToSelectedRequest()
	}
	scr.insertMode = false
	scr.methodSelectorOpen = false
	scr.body.Blur()
	return scr, nil
}

func (scr Screen) processNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == ':' {
		scr.commandPanel.Activate(":")
		return scr, nil
	}

	if scr.methodSelectorOpen {
		return scr.processMethodSelector(msg)
	}

	if msg.Type != tea.KeyRunes || len(msg.Runes) != 1 {
		if msg.Type == tea.KeyEnter || msg.Type == tea.KeySpace {
			if scr.focusedPanel == focusedPanelDoButton {
				if scr.requestInFlight {
					return scr, nil
				}
				if cmd := scr.requestCommand(); cmd != nil {
					scr.requestInFlight = true
					return scr, cmd
				}
				return scr, nil
			}
			scr.methodSelectorOpen = scr.focusedPanel == focusedPanelMethod
		}
		return scr, nil
	}

	switch msg.Runes[0] {
	case 'h':
		scr.focusLeftPanel()
	case 'l':
		scr.focusRightPanel()
	case 'i':
		if scr.focusedPanel == focusedPanelBody {
			scr.insertMode = true
			cmd := scr.body.EnterNavigationMode()
			return scr, cmd
		}
		if scr.focusedPanel != focusedPanelRequests && scr.focusedPanel != focusedPanelMethod {
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
		switch msg.Type {
		case tea.KeyRunes:
			scr.textInput += string(msg.Runes)
		case tea.KeyBackspace:
			if len(scr.textInput) > 0 {
				scr.textInput = scr.textInput[:len(scr.textInput)-1]
			}
		}
	case focusedPanelHeaders:
		if scr.updateHeadersEditor(msg) {
			scr.syncHeadersToSelectedRequest()
		}
	case focusedPanelBody:
		switch scr.body.Mode() {
		case presentationbody.NavigateMode:
			cmd, _ := scr.body.UpdateNavigationMode(msg)
			return scr, cmd
		case presentationbody.InsertMode:
			cmd, changed := scr.body.UpdateEditor(msg)
			if changed {
				scr.syncBodyToSelectedRequest()
			}
			return scr, cmd
		}
	}

	return scr, nil
}
