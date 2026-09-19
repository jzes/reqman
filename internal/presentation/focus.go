package presentation

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
