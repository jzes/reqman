package presentation

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jzes/reqman/internal/app/apperror"
	"github.com/jzes/reqman/internal/domain/request"
	presentationRequest "github.com/jzes/reqman/internal/presentation/request"
)

func (scr *Screen) selectNextRequest() {
	if scr.selectedRequestIndex < len(scr.requestPaths)-1 {
		scr.selectedRequestIndex++
		scr.ensureSelectedRequestVisible()
		scr.clearResponse()
		scr.showSelectedRequestURL()
		scr.showSelectedRequestHeaders()
		scr.showSelectedRequestBody()
		scr.showSelectedRequestMethod()
	}
}

func (scr *Screen) selectPreviousRequest() {
	if scr.selectedRequestIndex > 0 {
		scr.selectedRequestIndex--
		scr.ensureSelectedRequestVisible()
		scr.clearResponse()
		scr.showSelectedRequestURL()
		scr.showSelectedRequestHeaders()
		scr.showSelectedRequestBody()
		scr.showSelectedRequestMethod()
	}
}

func (scr *Screen) createRequest(name string) {
	name = strings.TrimSpace(name)
	if name == "" || scr.requestWriter == nil {
		return
	}
	if filepath.Ext(name) != ".curl" {
		name += ".curl"
	}
	name = filepath.Base(name)

	newRequest := request.Request{
		Name:    name,
		Method:  request.MethodGet,
		Headers: request.NewHeaders(),
	}
	if err := scr.requestWriter.WriteToFile(newRequest); err != nil {
		scr.showResponseError(fmt.Sprintf("Failed to create request: %v", err))
		return
	}

	scr.requestPaths = append(scr.requestPaths, newRequest.Name)
	scr.requests = append(scr.requests, newRequest)
	scr.loadedRequests[len(scr.requests)-1] = true
	scr.selectedRequestIndex = len(scr.requestPaths) - 1
	scr.ensureSelectedRequestVisible()
	scr.showSelectedRequestURL()
	scr.showSelectedRequestHeaders()
	scr.showSelectedRequestBody()
	scr.showSelectedRequestMethod()
	scr.clearResponse()
	scr.focusedPanel = focusedPanelURL
	scr.insertMode = true
	scr.url.EnterInsertMode()
	scr.methodSelectorOpen = false
	scr.body.Blur()
}

func (scr *Screen) clearResponse() {
	scr.hasResponse = false
	scr.responseError = ""
	scr.fatalError = false
	scr.response = request.Response{}
}

func (scr *Screen) showSelectedRequestURL() {
	if !scr.loadSelectedRequest() {
		scr.url.SetValue("")
		return
	}

	scr.url.SetValue(scr.requests[scr.selectedRequestIndex].URL.String())
}

func (scr *Screen) showSelectedRequestHeaders() {
	if !scr.loadSelectedRequest() {
		scr.headersEditor.SetRows(nil)
		return
	}

	scr.headersEditor.SetRows(headersFromDomain(scr.requests[scr.selectedRequestIndex].Headers))
}

func (scr *Screen) showSelectedRequestBody() {
	if !scr.loadSelectedRequest() {
		scr.body.SetValue("")
		return
	}

	scr.body.SetValue(scr.requests[scr.selectedRequestIndex].Body)
}

func (scr *Screen) showSelectedRequestMethod() {
	if !scr.loadSelectedRequest() {
		scr.methodList.Select(0)
		return
	}

	selectedMethod := scr.requests[scr.selectedRequestIndex].Method
	for i, method := range httpMethodItems() {
		if request.Method(method.(methodItem)) == selectedMethod {
			scr.methodList.Select(i)
			return
		}
	}
}

func (scr *Screen) setSelectedRequestMethod(method request.Method) {
	if !scr.loadSelectedRequest() {
		return
	}

	scr.requests[scr.selectedRequestIndex].Method = method
}

func (scr *Screen) syncBodyToSelectedRequest() {
	if !scr.loadSelectedRequest() {
		return
	}

	scr.requests[scr.selectedRequestIndex].Body = scr.body.Value()
}

func (scr *Screen) syncURLToSelectedRequest() {
	if !scr.loadSelectedRequest() {
		return
	}

	url, err := request.NewURL(scr.url.Value())
	if err != nil {
		scr.showResponseError(fmt.Sprintf("Invalid URL: %v", err))
		return
	}
	scr.requests[scr.selectedRequestIndex].URL = url
	scr.responseError = ""
}

func (scr *Screen) writeSelectedRequest() bool {
	if scr.requestWriter == nil || !scr.loadSelectedRequest() {
		return false
	}

	if err := scr.requestWriter.WriteToFile(scr.requests[scr.selectedRequestIndex]); err != nil {
		scr.showResponseError(fmt.Sprintf("Failed to save request: %v", err))
		return false
	}
	scr.responseError = ""
	return true
}

func (scr *Screen) startRequest() tea.Cmd {
	if scr.requestDoer == nil || !scr.loadSelectedRequest() {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	scr.requestCancel = cancel
	scr.nextRequestID++
	scr.inFlightRequestID = scr.nextRequestID
	scr.requestInFlight = true

	requestToDo := scr.requests[scr.selectedRequestIndex]
	requestID := scr.inFlightRequestID
	return func() tea.Msg {
		response, err := scr.requestDoer.Do(ctx, requestToDo)
		return presentationRequest.RequestResultMessage{RequestID: requestID, Response: response, Err: err, Fatal: errors.Is(err, apperror.ErrFatal)}
	}
}

func (scr *Screen) cancelRequest() {
	if scr.requestCancel != nil {
		scr.requestCancel()
	}
	scr.requestCancel = nil
	scr.requestInFlight = false
	scr.inFlightRequestID = 0
	scr.responseError = "Request canceled"
	scr.hasResponse = false
	scr.response = request.Response{}
}

func (scr *Screen) showResponseError(message string) {
	scr.responseError = message
	scr.fatalError = false
	scr.hasResponse = false
	scr.response = request.Response{}
}

func (scr *Screen) syncHeadersToSelectedRequest() {
	if !scr.loadSelectedRequest() {
		return
	}

	headers := request.NewHeaders()
	for _, row := range scr.headersEditor.Rows() {
		if row.Key == "" {
			continue
		}
		_ = headers.Add(row.Key, row.Value)
	}
	scr.requests[scr.selectedRequestIndex].Headers = headers
}

func (scr *Screen) loadSelectedRequest() bool {
	if len(scr.requestPaths) == 0 || scr.selectedRequestIndex < 0 || scr.selectedRequestIndex >= len(scr.requestPaths) {
		return false
	}
	if scr.loadedRequests[scr.selectedRequestIndex] {
		return true
	}

	path := scr.requestPaths[scr.selectedRequestIndex]
	if scr.requestLoader == nil {
		scr.requests[scr.selectedRequestIndex] = request.Request{
			Name:    filepath.Base(path),
			Path:    path,
			Method:  request.MethodGet,
			Headers: request.NewHeaders(),
		}
		scr.loadedRequests[scr.selectedRequestIndex] = true
		return true
	}

	loaded, err := scr.requestLoader.Load(path)
	if err != nil {
		scr.responseError = err.Error()
		return false
	}

	scr.requests[scr.selectedRequestIndex] = loaded
	scr.loadedRequests[scr.selectedRequestIndex] = true
	scr.responseError = ""
	return true
}

func (scr Screen) requestPanelContentHeight() int {
	screenHeight := scr.height
	if screenHeight == 0 {
		screenHeight = defaultScreenHeight
	}
	return max(0, screenHeight-topBarHeight-sidebarStyle.GetVerticalFrameSize())
}

func (scr *Screen) ensureSelectedRequestVisible() {
	visibleRequests := scr.requestPanelContentHeight()
	if visibleRequests <= 0 || len(scr.requestPaths) == 0 {
		scr.requestScrollOffset = 0
		return
	}

	if scr.selectedRequestIndex < scr.requestScrollOffset {
		scr.requestScrollOffset = scr.selectedRequestIndex
	}
	if scr.selectedRequestIndex >= scr.requestScrollOffset+visibleRequests {
		scr.requestScrollOffset = scr.selectedRequestIndex - visibleRequests + 1
	}

	maxOffset := max(0, len(scr.requestPaths)-visibleRequests)
	if scr.requestScrollOffset > maxOffset {
		scr.requestScrollOffset = maxOffset
	}
	if scr.requestScrollOffset < 0 {
		scr.requestScrollOffset = 0
	}
}
