package presentation

import (
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jzes/reqman/internal/domain/request"
	presentationRequest "github.com/jzes/reqman/internal/presentation/request"
)

func (scr *Screen) selectNextRequest() {
	if scr.selectedRequestIndex < len(scr.requests)-1 {
		scr.selectedRequestIndex++
		scr.showSelectedRequestURL()
		scr.showSelectedRequestHeaders()
		scr.showSelectedRequestBody()
		scr.showSelectedRequestMethod()
		scr.clearResponse()
	}
}

func (scr *Screen) selectPreviousRequest() {
	if scr.selectedRequestIndex > 0 {
		scr.selectedRequestIndex--
		scr.showSelectedRequestURL()
		scr.showSelectedRequestHeaders()
		scr.showSelectedRequestBody()
		scr.showSelectedRequestMethod()
		scr.clearResponse()
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
		Path:    filepath.Join(scr.requestDirectory, name),
		Method:  request.MethodGet,
		Headers: make(map[string]string),
	}
	if err := scr.requestWriter.WriteToFile(newRequest); err != nil {
		return
	}

	scr.requests = append(scr.requests, newRequest)
	scr.selectedRequestIndex = len(scr.requests) - 1
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
	scr.response = request.Response{}
}

func (scr *Screen) showSelectedRequestURL() {
	if len(scr.requests) == 0 {
		scr.url.SetValue("")
		return
	}

	scr.url.SetValue(scr.requests[scr.selectedRequestIndex].URL.String())
}

func (scr *Screen) showSelectedRequestHeaders() {
	if len(scr.requests) == 0 {
		scr.headersEditor.setRows(nil)
		return
	}

	scr.headersEditor.setRows(headersFromMap(scr.requests[scr.selectedRequestIndex].Headers))
}

func (scr *Screen) showSelectedRequestBody() {
	if len(scr.requests) == 0 {
		scr.body.SetValue("")
		return
	}

	scr.body.SetValue(scr.requests[scr.selectedRequestIndex].Body)
}

func (scr *Screen) showSelectedRequestMethod() {
	if len(scr.requests) == 0 {
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
	if len(scr.requests) == 0 {
		return
	}

	scr.requests[scr.selectedRequestIndex].Method = method
}

func (scr *Screen) syncBodyToSelectedRequest() {
	if len(scr.requests) == 0 {
		return
	}

	scr.requests[scr.selectedRequestIndex].Body = scr.body.Value()
}

func (scr *Screen) syncURLToSelectedRequest() {
	if len(scr.requests) == 0 {
		return
	}

	url, err := request.NewURL(scr.url.Value())
	if err != nil {
		return
	}
	scr.requests[scr.selectedRequestIndex].URL = url
}

func (scr *Screen) writeSelectedRequest() {
	if scr.requestWriter == nil || len(scr.requests) == 0 {
		return
	}

	_ = scr.requestWriter.WriteToFile(scr.requests[scr.selectedRequestIndex])
}

func (scr Screen) requestCommand() tea.Cmd {
	if scr.requestDoer == nil || len(scr.requests) == 0 {
		return nil
	}

	requestToDo := scr.requests[scr.selectedRequestIndex]
	return func() tea.Msg {
		response, err := scr.requestDoer.Do(requestToDo)
		return presentationRequest.RequestResultMessage{Response: response, Err: err}
	}
}

func (scr *Screen) syncHeadersToSelectedRequest() {
	if len(scr.requests) == 0 {
		return
	}

	headers := make(map[string]string)
	for _, row := range scr.headersEditor.rows {
		if row.key == "" {
			continue
		}
		headers[row.key] = row.value
	}
	scr.requests[scr.selectedRequestIndex].Headers = headers
}
