package presentation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

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
	contentHeight := max(1, screenHeight-topBarHeight)

	requestPanelContentHeight := scr.requestPanelContentHeight()
	sidebarLines := make([]string, 0, requestPanelContentHeight)
	if len(scr.requestPaths) == 0 {
		sidebarLines = append(sidebarLines, "  (empty)")
	} else {
		start := min(scr.requestScrollOffset, len(scr.requestPaths))
		end := min(len(scr.requestPaths), start+requestPanelContentHeight)
		for i, path := range scr.requestPaths[start:end] {
			requestIndex := start + i
			displayReq := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
			if len(displayReq) > 28 {
				displayReq = displayReq[:25] + "..."
			}

			prefix := "  "
			if requestIndex == scr.selectedRequestIndex {
				prefix = focusedStyle.Render(focusMarker)
			}
			sidebarLines = append(sidebarLines, fmt.Sprintf("%s%d. %s", prefix, requestIndex+1, displayReq))
		}
	}

	sidebarPanelStyle := styleForPanel(sidebarStyle, scr.focusedPanel == focusedPanelRequests, false)
	sidebar := renderPanelWithTitle(
		sidebarPanelStyle.
			Width(sidebarContentWidth).
			Height(requestPanelContentHeight),
		"Requests",
		strings.Join(sidebarLines, "\n"),
	)

	detailsWidth := max(0, screenWidth-lipgloss.Width(sidebar))
	panelContentWidth := max(0, detailsWidth-inputStyle.GetHorizontalFrameSize())
	panelStyle := inputStyle.Width(panelContentWidth)
	methodStyle := styleForPanel(inputStyle.Width(methodContentWidth), scr.focusedPanel == focusedPanelMethod, scr.methodSelectorOpen)
	headersStyle := styleForPanel(panelStyle, scr.focusedPanel == focusedPanelHeaders, scr.insertMode)
	bodyStyle := scr.body.Style(panelStyle, scr.focusedPanel == focusedPanelBody)

	methodPanel := renderPanelWithTitle(methodStyle, "Method", focusedStyle.Render(focusMarker)+scr.renderMethodSelector())
	doButton := scr.renderDoButton()
	urlContentWidth := max(0, detailsWidth-lipgloss.Width(methodPanel)-lipgloss.Width(doButton)-inputStyle.GetHorizontalFrameSize())
	urlStyle := scr.url.Style(inputStyle.Width(urlContentWidth), scr.focusedPanel == focusedPanelURL)
	requestURL := scr.url.View(urlContentWidth, scr.focusedPanel == focusedPanelURL)
	urlPanel := renderPanelWithTitle(urlStyle, "URL", focusedStyle.Render(focusMarker)+requestURL)
	requestLine := lipgloss.JoinHorizontal(lipgloss.Top, methodPanel, urlPanel, doButton)

	headerColumnWidth := min(25, max(10, panelContentWidth/3))
	valueColumnWidth := max(10, panelContentWidth-headerColumnWidth-3)
	headers := renderPanelWithTitle(headersStyle, "Headers", scr.renderHeadersEditor(headerColumnWidth, valueColumnWidth))

	usedHeight := lipgloss.Height(requestLine) + lipgloss.Height(headers)
	remainingPanelHeight := max(2, contentHeight-usedHeight-inputStyle.GetVerticalFrameSize()*2)
	bodyPanelHeight := max(1, remainingPanelHeight*2/5)
	responsePanelHeight := max(1, remainingPanelHeight-bodyPanelHeight)
	body := renderPanelWithTitle(
		bodyStyle.Height(bodyPanelHeight),
		"Body",
		scr.body.View(panelContentWidth, bodyPanelHeight, scr.focusedPanel == focusedPanelBody),
	)
	responseStyle := styleForPanel(panelStyle, scr.focusedPanel == focusedPanelResponse, scr.insertMode)
	response := renderPanelWithTitle(responseStyle.Height(responsePanelHeight), "Response", scr.renderResponse(panelContentWidth))
	details := lipgloss.JoinVertical(lipgloss.Left, requestLine, headers, body, response)

	mainView := lipgloss.JoinVertical(
		lipgloss.Left,
		renderTopBar(screenWidth),
		lipgloss.JoinHorizontal(lipgloss.Top, sidebar, details),
	)
	mainView = scr.newRequestPanel.Render(mainView, screenWidth)
	mainView = scr.renderHelpOverlay(mainView, screenWidth, screenHeight)
	return scr.commandPanel.Render(mainView, screenWidth)
}

func (scr Screen) renderHelpOverlay(baseView string, screenWidth int, screenHeight int) string {
	if !scr.helpOpen {
		return baseView
	}

	panelWidth := min(90, max(42, screenWidth*3/4))
	panelHeight := min(max(12, screenHeight*2/3), max(12, screenHeight-4))
	contentWidth := max(0, panelWidth-helpPanelStyle.GetHorizontalFrameSize())
	contentHeight := max(1, panelHeight-helpPanelStyle.GetVerticalFrameSize())
	content := renderHelpContent(contentWidth, contentHeight)
	panel := helpPanelStyle.Width(contentWidth).Height(contentHeight).Render(content)
	block := lipgloss.JoinVertical(lipgloss.Left, renderHelpTitleBar(lipgloss.Width(panel)), panel)
	panelLines := strings.Split(block, "\n")
	baseLines := strings.Split(baseView, "\n")
	insertAt := max(1, (screenHeight-lipgloss.Height(block))/2)
	leftOffset := max(0, (screenWidth-lipgloss.Width(block))/2)

	for len(baseLines) < insertAt+len(panelLines) {
		baseLines = append(baseLines, "")
	}
	for i, line := range panelLines {
		baseLines[insertAt+i] = overlayLine(baseLines[insertAt+i], line, leftOffset)
	}
	return strings.Join(baseLines, "\n")
}

func renderHelpTitleBar(width int) string {
	const (
		leftCap  = ""
		rightCap = ""
	)

	label := " Help "
	contentWidth := width - lipgloss.Width(leftCap) - lipgloss.Width(rightCap)
	if contentWidth <= 0 {
		return ""
	}
	leftPadding := max(0, (contentWidth-lipgloss.Width(label))/2)
	rightPadding := max(0, contentWidth-leftPadding-lipgloss.Width(label))
	content := strings.Repeat(" ", leftPadding) + label + strings.Repeat(" ", rightPadding)
	return helpTitleCapStyle(helpGradientColor(0, contentWidth)).Render(leftCap) + renderHelpTitleGradient(content, contentWidth) + helpTitleCapStyle(helpGradientColor(contentWidth-1, contentWidth)).Render(rightCap)
}

func renderHelpTitleGradient(text string, width int) string {
	var builder strings.Builder
	column := 0
	for _, r := range padOrTruncate(text, width) {
		builder.WriteString(helpTitleStyle(helpGradientColor(column, width)).Render(string(r)))
		column += lipgloss.Width(string(r))
	}
	return builder.String()
}

func helpGradientColor(column int, width int) lipgloss.Color {
	if width <= 1 {
		return lipgloss.Color("#93C5FD")
	}

	start := [3]int{0x93, 0xC5, 0xFD}
	end := [3]int{0x1D, 0x4E, 0xD8}
	ratio := float64(column) / float64(width-1)
	r := int(float64(start[0]) + (float64(end[0]-start[0]) * ratio))
	g := int(float64(start[1]) + (float64(end[1]-start[1]) * ratio))
	b := int(float64(start[2]) + (float64(end[2]-start[2]) * ratio))
	return lipgloss.Color(fmt.Sprintf("#%02X%02X%02X", r, g, b))
}

func helpTitleStyle(color lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().Background(color).Foreground(lipgloss.Color("#EFF6FF")).Bold(true)
}

func helpTitleCapStyle(color lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(color)
}

func renderHelpContent(width int, height int) string {
	sections := []string{
		"Panels:",
		"- Requests lists the .curl files loaded from the current directory.",
		"- Method selects the HTTP verb for the active request.",
		"- URL edits the request target.",
		"- Headers edits key/value header rows.",
		"- Body edits the request payload.",
		"- Response shows status, body, headers, raw data, and request stats.",
		"Navigation is modal. In normal mode, use h/j/k/l to move between panels, i to enter the current panel editing mode, and : to open commands. In text fields, the first i opens internal navigation and another i starts inserting text.",
		"In editing mode, Esc returns to field navigation or leaves the panel editing mode. Headers uses Tab and Shift+Tab to switch between key and value. Response uses [ and ] to switch tabs.",
		"Main commands: :w saves, :r runs, :wr saves and runs, :wq saves and quits, :q quits, and :? opens this help. Press Esc to close this panel.",
	}

	lines := make([]string, 0, height)
	for i, section := range sections {
		if i == 7 {
			lines = append(lines, "")
		}
		lines = append(lines, wrapText(section, max(1, width-4))...)
	}

	for len(lines) < height {
		lines = append(lines, "")
	}
	if len(lines) > height {
		lines = lines[:height]
	}
	for i, line := range lines {
		lines[i] = "  " + padOrTruncate(line, max(0, width-4)) + "  "
	}
	return strings.Join(lines, "\n")
}

func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{""}
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	lines := []string{words[0]}
	for _, word := range words[1:] {
		current := lines[len(lines)-1]
		if lipgloss.Width(current)+1+lipgloss.Width(word) <= width {
			lines[len(lines)-1] = current + " " + word
			continue
		}
		lines = append(lines, word)
	}
	return lines
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

func renderTopBar(width int) string {
	const (
		leftCap  = ""
		rightCap = ""
	)

	left := "Req-Man"
	right := ": to open commands"
	if width <= 0 {
		return ""
	}
	if width <= lipgloss.Width(leftCap)+lipgloss.Width(rightCap) {
		return topBarStyle.Width(width).Render(padOrTruncate(left, width))
	}

	contentWidth := width - lipgloss.Width(leftCap) - lipgloss.Width(rightCap)
	available := contentWidth - lipgloss.Width(left) - lipgloss.Width(right)
	if available < 1 {
		content := renderTopBarGradient(padOrTruncate(left, contentWidth), contentWidth, contentWidth)
		return topBarCapStyle(topBarGradientColor(0, contentWidth)).Render(leftCap) + content + topBarCapStyle(topBarGradientColor(contentWidth-1, contentWidth)).Render(rightCap)
	}

	rightStart := lipgloss.Width(left) + available
	content := renderTopBarGradient(left+strings.Repeat(" ", available)+right, contentWidth, rightStart)
	return topBarCapStyle(topBarGradientColor(0, contentWidth)).Render(leftCap) + content + topBarCapStyle(topBarGradientColor(contentWidth-1, contentWidth)).Render(rightCap)
}

func renderTopBarGradient(text string, width int, whiteFromColumn int) string {
	var builder strings.Builder
	column := 0
	for _, r := range padOrTruncate(text, width) {
		color := topBarGradientColor(column, width)
		style := topBarStyle.Background(color)
		if column >= whiteFromColumn {
			style = style.Foreground(lipgloss.Color("#FFFFFF"))
		}
		builder.WriteString(style.Render(string(r)))
		column += lipgloss.Width(string(r))
	}
	return builder.String()
}

func topBarGradientColor(column int, width int) lipgloss.Color {
	if width <= 1 {
		return lipgloss.Color("#C084FC")
	}

	start := [3]int{0xC0, 0x84, 0xFC}
	end := [3]int{0x6D, 0x28, 0xD9}
	ratio := float64(column) / float64(width-1)
	r := int(float64(start[0]) + (float64(end[0]-start[0]) * ratio))
	g := int(float64(start[1]) + (float64(end[1]-start[1]) * ratio))
	b := int(float64(start[2]) + (float64(end[2]-start[2]) * ratio))
	return lipgloss.Color(fmt.Sprintf("#%02X%02X%02X", r, g, b))
}

func topBarCapStyle(color lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(color)
}

func (scr Screen) renderResponse(widths ...int) string {
	if scr.requestInFlight {
		return "Executing request..."
	}

	if scr.responseError != "" {
		return "Error:\n" + scr.responseError
	}

	if !scr.hasResponse {
		return "(empty)"
	}

	width := defaultScreenWidth
	if len(widths) > 0 {
		width = widths[0]
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		scr.renderConnectedResponseTabs(),
		scr.renderSelectedResponseTabContent(width),
	)
}

func (scr Screen) renderConnectedResponseTabs() string {
	lines := strings.Split(scr.renderResponseTabs(), "\n")
	if len(lines) <= 1 {
		return strings.Join(lines, "\n")
	}
	return strings.Join(lines[:len(lines)-1], "\n")
}

func (scr Screen) renderResponseTabs() string {
	tabs := []struct {
		tab   responseTab
		label string
	}{
		{responseTabStats, "Stats"},
		{responseTabBody, "Body"},
		{responseTabHeaders, "Headers"},
		{responseTabRaw, "Raw"},
	}

	activeBorder := lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "│",
		BottomRight: "│",
	}
	activeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(focusedPurple)).
		Border(activeBorder).
		BorderForeground(lipgloss.Color(focusedPurple)).
		Padding(0, 1)
	inactiveStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(defaultPurple)).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(defaultPurple)).
		Padding(0, 1)

	renderedTabs := make([]string, 0, len(tabs))
	for _, tab := range tabs {
		label := tab.label
		if scr.selectedResponseTab == tab.tab {
			renderedTabs = append(renderedTabs, activeStyle.Render(label))
			continue
		}
		renderedTabs = append(renderedTabs, inactiveStyle.Render(label))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
}

func (scr Screen) renderSelectedResponseTabContent(width int) string {
	content := scr.renderSelectedResponseTab(width)
	outerWidth := max(4, width)
	contentWidth := max(0, outerWidth-4)
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(focusedPurple))

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = borderStyle.Render("│") + " " + padOrTruncate(line, contentWidth) + " " + borderStyle.Render("│")
	}

	parts := make([]string, 0, len(lines)+2)
	parts = append(parts, borderStyle.Render(scr.renderResponseTabContentTopBorder(outerWidth)))
	parts = append(parts, lines...)
	parts = append(parts, borderStyle.Render("╰"+strings.Repeat("─", max(0, outerWidth-2))+"╯"))
	return strings.Join(parts, "\n")
}

func (scr Screen) renderResponseTabContentTopBorder(width int) string {
	activeTabLeft, activeTabWidth := scr.selectedResponseTabPosition()
	activeTabRight := activeTabLeft + activeTabWidth - 1

	var builder strings.Builder
	for column := 0; column < width; column++ {
		char := "─"
		if column == 0 {
			char = "╭"
		}
		if column == width-1 {
			char = "╮"
		}

		if column >= activeTabLeft && column <= activeTabRight {
			switch column {
			case activeTabLeft:
				char = "╯"
				if column == 0 {
					char = "│"
				}
			case activeTabRight:
				char = "╰"
				if column == width-1 {
					char = "│"
				}
			default:
				char = " "
			}
		}

		builder.WriteString(char)
	}

	return builder.String()
}

func (scr Screen) selectedResponseTabPosition() (int, int) {
	tabs := []struct {
		tab   responseTab
		label string
	}{
		{responseTabStats, "Stats"},
		{responseTabBody, "Body"},
		{responseTabHeaders, "Headers"},
		{responseTabRaw, "Raw"},
	}

	offset := 0
	for _, tab := range tabs {
		width := lipgloss.Width(tab.label) + 4
		if scr.selectedResponseTab == tab.tab {
			return offset, width
		}
		offset += width
	}

	return 0, lipgloss.Width("Stats") + 4
}

func (scr Screen) renderSelectedResponseTab(width int) string {
	switch scr.selectedResponseTab {
	case responseTabBody:
		return scr.renderResponseBody()
	case responseTabHeaders:
		return scr.renderResponseHeaders(width)
	case responseTabRaw:
		return scr.renderResponseRaw()
	default:
		return scr.renderResponseStats()
	}
}

func (scr Screen) renderResponseStats() string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Status: %s\n", scr.response.Status)
	fmt.Fprintf(&builder, "Status Code: %d\n", scr.response.StatusCode)
	fmt.Fprintf(&builder, "Duration: %s", scr.response.Duration)
	return builder.String()
}

func (scr Screen) renderResponseBody() string {
	return formatJSONBody(scr.response.Body)
}

func (scr Screen) renderResponseHeaders(width int) string {
	headerColumnWidth := min(25, max(10, width/3))
	valueColumnWidth := max(10, width-headerColumnWidth-2)

	var builder strings.Builder
	builder.WriteString(padOrTruncate("Header", headerColumnWidth))
	builder.WriteString("  ")
	builder.WriteString(padOrTruncate("Value", valueColumnWidth))

	if len(scr.response.Headers) == 0 {
		builder.WriteString("\n  (empty)")
		return builder.String()
	}

	keys := make([]string, 0, len(scr.response.Headers))
	for key := range scr.response.Headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		builder.WriteString("\n")
		builder.WriteString(padOrTruncate(key, headerColumnWidth))
		builder.WriteString("  ")
		builder.WriteString(padOrTruncate(strings.Join(scr.response.Headers[key], ", "), valueColumnWidth))
	}

	return builder.String()
}

func (scr Screen) renderResponseRaw() string {
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

func formatJSONBody(body string) string {
	if strings.TrimSpace(body) == "" {
		return "(empty)"
	}

	var formatted bytes.Buffer
	if err := json.Indent(&formatted, []byte(body), "", "  "); err != nil {
		return body
	}
	return formatted.String()
}

func (scr Screen) renderDoButton() string {
	contentStyle := lipgloss.NewStyle().
		Width(doButtonContentWidth).
		Align(lipgloss.Left)

	panelStyle := inputStyle.Width(doButtonContentWidth + 2)
	if scr.focusedPanel == focusedPanelDoButton {
		panelStyle = panelStyle.BorderForeground(lipgloss.Color(focusedPurple))
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
			prefix = focusedStyle.Render(focusMarker)
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
	if scr.loadedRequests[scr.selectedRequestIndex] {
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
	return style.BorderForeground(lipgloss.Color(focusedPurple))
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
