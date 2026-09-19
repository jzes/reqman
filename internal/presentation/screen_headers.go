package presentation

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/jzes/reqman/internal/domain/request"
)

func newHeadersEditor() headersEditor {
	return headersEditor{}
}

func (e *headersEditor) setRows(rows []headerRow) {
	e.rows = rows
	e.clampSelection()
}

func (e *headersEditor) clampSelection() {
	if e.selectedRow < 0 {
		e.selectedRow = 0
	}
	if len(e.rows) == 0 {
		e.selectedRow = 0
		e.selectedCell = headerCellKey
		return
	}
	if e.selectedRow >= len(e.rows) {
		e.selectedRow = len(e.rows) - 1
	}
	if e.selectedCell != headerCellValue {
		e.selectedCell = headerCellKey
	}
}

func (scr *Screen) ensureEditableHeaderRow() {
	if len(scr.headersEditor.rows) == 0 {
		scr.headersEditor.rows = []headerRow{{}}
		scr.headersEditor.selectedRow = 0
		scr.headersEditor.selectedCell = headerCellKey
	}
}

func (scr *Screen) updateHeadersEditor(msg tea.KeyMsg) bool {
	if len(scr.headersEditor.rows) == 0 {
		scr.ensureEditableHeaderRow()
	}

	mutated := false
	switch msg.Type {
	case tea.KeyRunes:
		scr.headersEditor.appendToSelectedCell(string(msg.Runes))
		mutated = true
	case tea.KeyBackspace:
		mutated = scr.headersEditor.backspaceSelectedCell()
	case tea.KeyTab, tea.KeyRight:
		scr.headersEditor.selectedCell = headerCellValue
	case tea.KeyShiftTab, tea.KeyLeft:
		scr.headersEditor.selectedCell = headerCellKey
	case tea.KeyEnter:
		if scr.headersEditor.selectedCell == headerCellKey {
			scr.headersEditor.selectedCell = headerCellValue
		} else {
			scr.headersEditor.insertRowAfterSelection()
			mutated = true
		}
	case tea.KeyUp:
		scr.headersEditor.selectPreviousRow()
	case tea.KeyDown:
		scr.headersEditor.selectNextRow()
	case tea.KeyDelete, tea.KeyCtrlD:
		mutated = scr.headersEditor.deleteSelectedRow()
	}
	return mutated
}

func (e *headersEditor) appendToSelectedCell(text string) {
	if len(e.rows) == 0 {
		return
	}
	row := &e.rows[e.selectedRow]
	if e.selectedCell == headerCellValue {
		row.value += text
		return
	}
	row.key += text
}

func (e *headersEditor) backspaceSelectedCell() bool {
	if len(e.rows) == 0 {
		return false
	}
	row := &e.rows[e.selectedRow]
	if e.selectedCell == headerCellValue {
		newValue, changed := removeLastRune(row.value)
		row.value = newValue
		return changed
	}
	newKey, changed := removeLastRune(row.key)
	row.key = newKey
	return changed
}

func (e *headersEditor) insertRowAfterSelection() {
	newRowIndex := e.selectedRow + 1
	e.rows = append(e.rows, headerRow{})
	copy(e.rows[newRowIndex+1:], e.rows[newRowIndex:])
	e.rows[newRowIndex] = headerRow{}
	e.selectedRow = newRowIndex
	e.selectedCell = headerCellKey
}

func (e *headersEditor) deleteSelectedRow() bool {
	if len(e.rows) == 0 {
		return false
	}
	e.rows = append(e.rows[:e.selectedRow], e.rows[e.selectedRow+1:]...)
	e.clampSelection()
	return true
}

func (e *headersEditor) selectPreviousRow() {
	if e.selectedRow > 0 {
		e.selectedRow--
	}
}

func (e *headersEditor) selectNextRow() {
	if e.selectedRow < len(e.rows)-1 {
		e.selectedRow++
	}
}

func headersFromDomain(headers request.Headers) []headerRow {
	requestHeaders := headers.List()
	rows := make([]headerRow, 0, len(requestHeaders))
	for _, header := range requestHeaders {
		rows = append(rows, headerRow{key: header.Key, value: header.Value})
	}
	return rows
}
