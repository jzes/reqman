package presentation

import (
	"github.com/jzes/reqman/internal/domain/request"
	"github.com/jzes/reqman/internal/presentation/panels/headers"
)

func (scr *Screen) ensureEditableHeaderRow() {
	scr.headersEditor.EnsureEditableRow()
}

func headersFromDomain(requestHeaders request.Headers) []headers.Row {
	domainHeaders := requestHeaders.List()
	rows := make([]headers.Row, 0, len(domainHeaders))
	for _, header := range domainHeaders {
		rows = append(rows, headers.Row{Key: header.Key, Value: header.Value})
	}
	return rows
}
