package ui

import "strings"

func resultLayoutKey(columns []string) string {
	return strings.Join(columns, "\x1f")
}

func (m *Model) applyResultLayout(t *tableState) {
	if t.res == nil {
		return
	}
	key := resultLayoutKey(t.res.Columns)
	t.setHidden(m.uiState.ResultHiddenColumns[key])
	t.applyWidths(m.uiState.ResultColumnWidths[key])
}

func (m *Model) persistResultLayout(t *tableState) {
	if t.res == nil {
		return
	}
	key := resultLayoutKey(t.res.Columns)
	if m.uiState.ResultHiddenColumns == nil {
		m.uiState.ResultHiddenColumns = map[string][]string{}
	}
	if m.uiState.ResultColumnWidths == nil {
		m.uiState.ResultColumnWidths = map[string]map[string]int{}
	}
	m.uiState.ResultHiddenColumns[key] = t.hiddenList()
	m.uiState.ResultColumnWidths[key] = t.widthMap()
	m.persistUIState()
}
