package presentation

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"

	"github.com/jzes/reqman/internal/domain/request"
)

func (scr Screen) View() string {
	screenWidth := scr.width
	if screenWidth == 0 {
		screenWidth = defaultScreenWidth
	}
	screenHeight := scr.height
	if screenHeight == 0 {
		screenHeight = defaultScreenHeight
	}

	var sidebarBuilder strings.Builder
	if len(scr.requests) == 0 {
		sidebarBuilder.WriteString("  (empty)\n")
	} else {
		for i, req := range scr.requests {
			displayReq := req.Name
			if len(displayReq) > 28 {
				displayReq = displayReq[:25] + "..."
			}

			prefix := "  "
			if i == scr.selectedRequestIndex {
				prefix = focusedStyle.Render("> ")
			}
			fmt.Fprintf(&sidebarBuilder, "%s%d. %s\n", prefix, i+1, displayReq)
		}
	}

	sidebarPanelStyle := styleForPanel(sidebarStyle, scr.focusedPanel == focusedPanelRequests, false)
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
	methodStyle := styleForPanel(inputStyle.Width(methodContentWidth), scr.focusedPanel == focusedPanelMethod, scr.methodSelectorOpen)
	headersStyle := styleForPanel(panelStyle, scr.focusedPanel == focusedPanelHeaders, scr.insertMode)
	bodyStyle := scr.body.Style(panelStyle, scr.focusedPanel == focusedPanelBody)

	methodPanel := renderPanelWithTitle(methodStyle, "Method", focusedStyle.Render("> ")+scr.renderMethodSelector())
	doButton := scr.renderDoButton()
	urlContentWidth := max(0, detailsWidth-lipgloss.Width(methodPanel)-lipgloss.Width(doButton)-inputStyle.GetHorizontalFrameSize())
	urlStyle := scr.url.Style(inputStyle.Width(urlContentWidth), scr.focusedPanel == focusedPanelURL)
	requestURL := scr.url.View(urlContentWidth, scr.focusedPanel == focusedPanelURL)
	urlPanel := renderPanelWithTitle(urlStyle, "URL", focusedStyle.Render("> ")+requestURL)
	requestLine := lipgloss.JoinHorizontal(lipgloss.Top, methodPanel, urlPanel, doButton)

	headerColumnWidth := min(25, max(10, panelContentWidth/3))
	valueColumnWidth := max(10, panelContentWidth-headerColumnWidth-3)
	headers := renderPanelWithTitle(headersStyle, "Headers", scr.renderHeadersEditor(headerColumnWidth, valueColumnWidth))

	usedHeight := lipgloss.Height(requestLine) + lipgloss.Height(headers)
	remainingPanelHeight := max(2, screenHeight-usedHeight-inputStyle.GetVerticalFrameSize()*2)
	bodyPanelHeight := max(1, remainingPanelHeight/2)
	responsePanelHeight := max(1, remainingPanelHeight-bodyPanelHeight)
	body := renderPanelWithTitle(
		bodyStyle.Height(bodyPanelHeight),
		"Body",
		scr.body.View(panelContentWidth, bodyPanelHeight, scr.focusedPanel == focusedPanelBody),
	)
	responseStyle := styleForPanel(panelStyle, scr.focusedPanel == focusedPanelResponse, false)
	response := renderPanelWithTitle(responseStyle.Height(responsePanelHeight), "Response", scr.renderResponse())
	details := lipgloss.JoinVertical(lipgloss.Left, requestLine, headers, body, response)

	mainView := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, details)
	mainView = scr.newRequestPanel.Render(mainView, screenWidth)
	return scr.commandPanel.Render(mainView, screenWidth)
}

func (scr Screen) renderResponse() string {
	if scr.requestInFlight {
		return "Executing request..."
	}

	if scr.responseError != "" {
		return "Error:\n" + scr.responseError
	}

	if !scr.hasResponse {
		return "(empty)"
	}

	var builder strings.Builder
	fmt.Fprintf(&builder, "URL: %s\n", scr.response.URL.String())
	fmt.Fprintf(&builder, "Status: %s\n", scr.response.Status)
	fmt.Fprintf(&builder, "Status Code: %d\n", scr.response.StatusCode)
	fmt.Fprintf(&builder, "Duration: %s\n", scr.response.Duration)

	if len(scr.response.Headers) == 0 {
		builder.WriteString("Headers:\n  (empty)\n")
	} else {
		builder.WriteString("Headers:\n")
		keys := make([]string, 0, len(scr.response.Headers))
		for key := range scr.response.Headers {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			fmt.Fprintf(&builder, "  %s: %s\n", key, strings.Join(scr.response.Headers[key], ", "))
		}
	}

	builder.WriteString("Body:\n")
	if scr.response.Body == "" {
		builder.WriteString("  (empty)")
	} else {
		builder.WriteString(scr.response.Body)
	}

	return strings.TrimRight(builder.String(), "\n")
}

func (scr Screen) renderDoButton() string {
	contentStyle := lipgloss.NewStyle().
		Width(doButtonContentWidth).
		Align(lipgloss.Left)

	panelStyle := inputStyle.Width(doButtonContentWidth + 2)
	if scr.focusedPanel == focusedPanelDoButton {
		panelStyle = panelStyle.BorderForeground(lipgloss.Color("99"))
	}
	if scr.requestInFlight {
		panelStyle = panelStyle.BorderForeground(lipgloss.Color("#F1FA8C"))
	}

	content := "Ready"
	if scr.requestInFlight {
		content = "Doing... " + scr.statusSpinner.View()
	}

	return renderPanelWithTitle(panelStyle, "Status", contentStyle.Render(content))
}

func (scr Screen) renderHeadersEditor(headerColumnWidth, valueColumnWidth int) string {
	var builder strings.Builder
	builder.WriteString(padOrTruncate("Header", headerColumnWidth))
	builder.WriteString("  ")
	builder.WriteString(padOrTruncate("Value", valueColumnWidth))

	if len(scr.headersEditor.rows) == 0 {
		builder.WriteString("\n  (empty)")
		return builder.String()
	}

	for i, row := range scr.headersEditor.rows {
		selected := scr.focusedPanel == focusedPanelHeaders && i == scr.headersEditor.selectedRow
		key := row.key
		value := row.value
		if selected && scr.insertMode {
			if scr.headersEditor.selectedCell == headerCellKey {
				key = textWithCursor(key, true)
			} else {
				value = textWithCursor(value, true)
			}
		}

		keyCell := padOrTruncate(key, headerColumnWidth)
		valueCell := padOrTruncate(value, valueColumnWidth)
		if selected {
			if scr.headersEditor.selectedCell == headerCellKey {
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

func (scr Screen) renderMethodSelector() string {
	if scr.methodSelectorOpen {
		methodList := scr.methodList
		methodList.SetSize(14, 7)
		return methodList.View()
	}

	method := request.MethodGet
	if len(scr.requests) > 0 {
		method = scr.requests[scr.selectedRequestIndex].Method
	}

	selector := fmt.Sprintf("[%s]", method)
	if scr.focusedPanel == focusedPanelURL {
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

func (i methodItem) FilterValue() string { return string(i) }
func (i methodItem) Title() string       { return string(i) }
func (i methodItem) Description() string { return "" }
