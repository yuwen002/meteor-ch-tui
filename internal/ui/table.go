package ui

import (
	"strings"

	"github.com/mattn/go-runewidth"

	"github.com/yuwen002/meteor-ch-tui/internal/ch"
)

const (
	minColWidth = 6
	maxColWidth = 40
	sampleRows  = 200
	sepChar     = "│"
)

// tableState 管理单个结果集的宽表状态（列宽、横向偏移、光标、顶部行）。
type tableState struct {
	res      *ch.Result
	widths   []int
	order    []int
	hidden   map[string]bool
	startCol int
	cursor   int
	topRow   int
}

func (t *tableState) setResult(res *ch.Result) {
	t.res = res
	t.widths = computeWidths(res)
	t.hidden = map[string]bool{}
	t.rebuildOrder()
	t.startCol = 0
	t.cursor = 0
	t.topRow = 0
}

func (t *tableState) rebuildOrder() {
	t.order = nil
	if t.res == nil {
		return
	}
	for i, name := range t.res.Columns {
		if !t.hidden[name] {
			t.order = append(t.order, i)
		}
	}
	if len(t.order) == 0 && len(t.res.Columns) > 0 {
		t.order = []int{0}
		delete(t.hidden, t.res.Columns[0])
	}
}

func (t *tableState) setHidden(names []string) {
	t.hidden = map[string]bool{}
	for _, name := range names {
		t.hidden[name] = true
	}
	t.rebuildOrder()
	t.startCol = 0
}

func (t *tableState) hiddenList() []string {
	if t.res == nil {
		return nil
	}
	var out []string
	for _, name := range t.res.Columns {
		if t.hidden[name] {
			out = append(out, name)
		}
	}
	return out
}

func (t *tableState) toggleColumn(name string) {
	if t.hidden == nil {
		t.hidden = map[string]bool{}
	}
	if t.hidden[name] {
		delete(t.hidden, name)
	} else {
		t.hidden[name] = true
	}
	t.rebuildOrder()
	t.startCol = clampInt(t.startCol, 0, maxInt(0, len(t.order)-1))
}

func (t *tableState) showAllColumns() {
	t.hidden = map[string]bool{}
	t.rebuildOrder()
	t.startCol = 0
}

func (t *tableState) widthMap() map[string]int {
	out := map[string]int{}
	if t.res == nil {
		return out
	}
	for i, name := range t.res.Columns {
		if i < len(t.widths) {
			out[name] = t.widths[i]
		}
	}
	return out
}

func (t *tableState) applyWidths(widths map[string]int) {
	if t.res == nil {
		return
	}
	for i, name := range t.res.Columns {
		if w, ok := widths[name]; ok && i < len(t.widths) {
			t.widths[i] = clampInt(w, minColWidth, maxColWidth)
		}
	}
}

func (t *tableState) adjustColumnWidth(name string, delta int) {
	if t.res == nil {
		return
	}
	for i, col := range t.res.Columns {
		if col == name && i < len(t.widths) {
			t.widths[i] = clampInt(t.widths[i]+delta, minColWidth, maxColWidth)
			return
		}
	}
}

func (t *tableState) resetColumnWidth(name string) {
	if t.res == nil {
		return
	}
	auto := computeWidths(t.res)
	for i, col := range t.res.Columns {
		if col == name && i < len(t.widths) && i < len(auto) {
			t.widths[i] = auto[i]
			return
		}
	}
}

func computeWidths(res *ch.Result) []int {
	if res == nil {
		return nil
	}
	w := make([]int, len(res.Columns))
	for i, c := range res.Columns {
		w[i] = runewidth.StringWidth(c)
	}
	n := len(res.Rows)
	if n > sampleRows {
		n = sampleRows
	}
	for r := 0; r < n; r++ {
		row := res.Rows[r]
		for i := range res.Columns {
			if i >= len(row) {
				continue
			}
			if ww := runewidth.StringWidth(ch.FormatValue(row[i])); ww > w[i] {
				w[i] = ww
			}
		}
	}
	for i := range w {
		w[i] = clampInt(w[i], minColWidth, maxColWidth)
	}
	return w
}

func (t *tableState) visibleColumns(contentWidth int) []int {
	var idx []int
	used := 0
	for pos := t.startCol; pos < len(t.order); pos++ {
		i := t.order[pos]
		if i < 0 || i >= len(t.widths) {
			continue
		}
		add := t.widths[i] + 2
		if len(idx) > 0 {
			add++
		}
		if len(idx) > 0 && used+add > contentWidth {
			break
		}
		used += add
		idx = append(idx, i)
		if used >= contentWidth {
			break
		}
	}
	return idx
}

func (t *tableState) scrollCols(delta int, contentWidth int) {
	if t.res == nil || len(t.order) == 0 {
		return
	}
	t.startCol = clampInt(t.startCol+delta, 0, t.maxStartCol(contentWidth))
}

// maxStartCol 返回横向滚动上限：从该列开始已经能看到最后一个可见列。
func (t *tableState) maxStartCol(contentWidth int) int {
	n := len(t.order)
	if n == 0 {
		return 0
	}
	for start := 0; start < n; start++ {
		probe := &tableState{widths: t.widths, order: t.order, startCol: start}
		cols := probe.visibleColumns(contentWidth)
		if len(cols) > 0 && cols[len(cols)-1] == t.order[n-1] {
			return start
		}
	}
	return n - 1
}

func (t *tableState) moveCursor(delta, rowsVisible int) {
	if t.res == nil || len(t.res.Rows) == 0 {
		return
	}
	t.cursor = clampInt(t.cursor+delta, 0, len(t.res.Rows)-1)
	t.ensureCursor(rowsVisible)
}

func (t *tableState) ensureCursor(rowsVisible int) {
	if t.res == nil || rowsVisible <= 0 {
		return
	}
	if t.cursor < t.topRow {
		t.topRow = t.cursor
	}
	if t.cursor >= t.topRow+rowsVisible {
		t.topRow = t.cursor - rowsVisible + 1
	}
	maxTop := len(t.res.Rows) - rowsVisible
	if maxTop < 0 {
		maxTop = 0
	}
	t.topRow = clampInt(t.topRow, 0, maxTop)
}

func (t *tableState) cell(rowIdx, col int) string {
	row := t.res.Rows[rowIdx]
	if col >= len(row) {
		return ""
	}
	return ch.FormatValue(row[col])
}

func (t *tableState) headerLine(idx []int) string {
	parts := make([]string, 0, len(idx))
	for _, c := range idx {
		parts = append(parts, " "+padCell(t.res.Columns[c], t.widths[c])+" ")
	}
	return strings.Join(parts, sepChar)
}

func (t *tableState) rowLine(rowIdx int, idx []int) string {
	parts := make([]string, 0, len(idx))
	for _, c := range idx {
		parts = append(parts, " "+padCell(t.cell(rowIdx, c), t.widths[c])+" ")
	}
	return strings.Join(parts, sepChar)
}
