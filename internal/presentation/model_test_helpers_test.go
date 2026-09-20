package presentation

import (
	"context"
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jzes/reqman/internal/domain/request"
)

type fakeWriter struct {
	request request.Request
	calls   int
	err     error
}

func (w *fakeWriter) WriteToFile(req request.Request) error {
	w.request = req
	w.calls++
	return w.err
}

type fakeDoer struct {
	request  request.Request
	calls    int
	response request.Response
	err      error
	ctx      context.Context
}

func (d *fakeDoer) Do(ctx context.Context, req request.Request) (request.Response, error) {
	d.ctx = ctx
	d.request = req
	d.calls++
	return d.response, d.err
}

func enterHeaderInsertMode(t *testing.T, screen Screen) Screen {
	t.Helper()
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	return screen
}

func enterBodyTextInsertMode(t *testing.T, screen Screen) Screen {
	t.Helper()
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	return screen
}

func enterURLTextInsertMode(t *testing.T, screen Screen) Screen {
	t.Helper()
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	screen, _ = updateScreen(t, screen, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	return screen
}

func newScreen(requests []request.Request) Screen {
	return newScreenWithDeps(requests, nil, nil)
}

func newScreenWithDeps(requests []request.Request, writer *fakeWriter, doer *fakeDoer) Screen {
	paths := make([]string, len(requests))
	requestsByPath := make(map[string]request.Request, len(requests))
	for i, req := range requests {
		path := req.Path
		if path == "" {
			path = req.Name
		}
		if path == "" {
			path = fmt.Sprintf("req-%d.curl", i)
		}
		paths[i] = path
		requestsByPath[path] = req
	}
	return NewScreen(paths, writer, doer, fakeLoader{requests: requestsByPath})
}

type fakeLoader struct {
	requests map[string]request.Request
}

func (l fakeLoader) Load(path string) (request.Request, error) {
	return l.requests[path], nil
}

type countingLoader struct {
	requests map[string]request.Request
	calls    map[string]int
}

func (l *countingLoader) Load(path string) (request.Request, error) {
	if l.calls == nil {
		l.calls = make(map[string]int)
	}
	l.calls[path]++
	return l.requests[path], nil
}

func updateScreen(t *testing.T, screen Screen, msg tea.Msg) (Screen, tea.Cmd) {
	t.Helper()
	model, cmd := screen.Update(msg)
	updated, ok := model.(Screen)
	if !ok {
		t.Fatalf("updated model has type %T, want Screen", model)
	}
	return updated, cmd
}

func firstCommandMessage(t *testing.T, cmd tea.Cmd) tea.Msg {
	t.Helper()
	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		return msg
	}
	if len(batch) == 0 {
		t.Fatal("batch command is empty")
	}
	return batch[0]()
}
