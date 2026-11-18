package ui

func (m *Model) sidebarWidth() int {
	w := sidebarOuterWidth
	if m.width > 0 && w > m.width/2 {
		w = m.width / 2
	}
	if w < 16 {
		w = 16
	}
	return w
}

func (m *Model) bodyHeight() int {
	h := m.height - 2
	if h < 6 {
		h = 6
	}
	return h
}

func (m *Model) rightWidth() int {
	w := m.width - m.sidebarWidth()
	if w < 20 {
		w = 20
	}
	return w
}

func (m *Model) tableContentWidth() int {
	return m.rightWidth() - 2
}

func (m *Model) tablePaneHeight() int {
	return m.bodyHeight() - editorPaneHeight
}

func (m *Model) tableRowsVisible() int {
	h := m.tablePaneHeight() - 4
	if h < 1 {
		h = 1
	}
	return h
}

func (m *Model) layout() {
	if m.width == 0 || m.height == 0 {
		return
	}
	m.editor.SetWidth(m.rightWidth() - 2)
	m.table.ensureCursor(m.tableRowsVisible())
	m.slow.ensureCursor(m.tableRowsVisible())
}

func (m *Model) modalPageSize() int {
	h := m.height - 6
	if h < 3 {
		h = 3
	}
	return h
}

func (m *Model) rowModalLines() int {
	t := &m.table
	if t.res == nil || t.cursor >= len(t.res.Rows) {
		return 0
	}
	return len(t.res.Columns)
}
