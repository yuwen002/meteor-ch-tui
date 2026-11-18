package ui

import (
	"strconv"

	"github.com/xuri/excelize/v2"
)

func colName(index int) string {
	name, _ := excelize.ColumnNumberToName(index + 1)
	return name
}

func writeXLSX(sheet string, columns []string, rows [][]any, path string) error {
	f := excelize.NewFile()
	defer f.Close()
	if err := f.SetSheetName("Sheet1", sheet); err != nil {
		return err
	}

	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: "Arial", Size: 11, Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"4472C4"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return err
	}
	dataStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Family: "Arial", Size: 11},
	})
	if err != nil {
		return err
	}

	for i, col := range columns {
		cell := colName(i) + "1"
		if err := f.SetCellValue(sheet, cell, col); err != nil {
			return err
		}
		if err := f.SetCellStyle(sheet, cell, cell, headerStyle); err != nil {
			return err
		}
		width := len([]rune(col)) + 4
		if width < 12 {
			width = 12
		}
		if width > 40 {
			width = 40
		}
		if err := f.SetColWidth(sheet, colName(i), colName(i), float64(width)); err != nil {
			return err
		}
	}
	if err := f.SetRowHeight(sheet, 1, 24); err != nil {
		return err
	}

	for r, row := range rows {
		for c := range row {
			cell := colName(c) + strconv.Itoa(r+2)
			if err := f.SetCellValue(sheet, cell, xlsxValue(row[c])); err != nil {
				return err
			}
			if err := f.SetCellStyle(sheet, cell, cell, dataStyle); err != nil {
				return err
			}
		}
	}
	return f.SaveAs(path)
}
