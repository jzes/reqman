package presentation

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jzes/reqman/internal/adapter/bodyeditor"
	"github.com/jzes/reqman/internal/config"
)

type bodyEditedMsg struct {
	index int
	body  string
	err   error
}

func (scr Screen) startBodyEdit() (tea.Model, tea.Cmd) {
	if scr.editorCancel != nil || !scr.loadSelectedRequest() {
		return scr, nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	req := scr.requests[scr.selectedRequestIndex]
	session, err := bodyeditor.New(ctx, scr.config.EditorMode, os.Getenv("EDITOR"), req.Name, req.Body)
	if err != nil {
		cancel()
		scr.showResponseError(fmt.Sprintf("Failed to open body editor: %v", err))
		return scr, nil
	}
	scr.editorCancel = cancel
	index := scr.selectedRequestIndex
	finish := func(err error) tea.Msg {
		return bodyEditedMsg{index: index, body: session.Body, err: err}
	}
	if scr.config.EditorMode == config.InLine {
		return scr, tea.Exec(session, finish)
	}
	return scr, func() tea.Msg { return finish(session.Run()) }
}

func (scr Screen) finishBodyEdit(msg bodyEditedMsg) (tea.Model, tea.Cmd) {
	scr.cancelBodyEdit()
	if msg.err != nil {
		scr.showResponseError(fmt.Sprintf("Failed to edit body: %v", msg.err))
		return scr, nil
	}
	if msg.index < 0 || msg.index >= len(scr.requests) {
		return scr, nil
	}
	scr.requests[msg.index].Body = msg.body
	if scr.selectedRequestIndex == msg.index {
		scr.showSelectedRequestBody()
	}
	return scr, nil
}

func (scr *Screen) cancelBodyEdit() {
	if scr.editorCancel != nil {
		scr.editorCancel()
		scr.editorCancel = nil
	}
}
