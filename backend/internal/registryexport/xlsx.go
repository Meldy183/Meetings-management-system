package registryexport

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/xuri/excelize/v2"

	"meetings-editor/internal/domain/person"
)

const sheetName = "Реестр"

var headers = []string{"Фамилия", "Имя", "Отчество", "Информация", "ID"}

// WriteXLSX creates a new file. Existing files are never overwritten.
// The first four columns match the application's Excel import format.
func WriteXLSX(people []person.Person, path string) (err error) {
	if !strings.EqualFold(filepath.Ext(path), ".xlsx") {
		return errors.New("output file must have the .xlsx extension")
	}
	if len(people) > 1048575 {
		return errors.New("registry exceeds the XLSX limit of 1048575 participant rows")
	}
	f := excelize.NewFile()
	defer func() { err = errors.Join(err, f.Close()) }()
	if err := f.SetSheetName("Sheet1", sheetName); err != nil {
		return err
	}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellStr(sheetName, cell, header); err != nil {
			return err
		}
	}
	for i, p := range people {
		values := []string{p.LastName, p.FirstName, p.MiddleName, p.Info}
		for j, value := range values {
			if utf8.RuneCountInString(value) > 32767 {
				return fmt.Errorf("participant ID %d: column %q exceeds the XLSX cell limit", p.ID, headers[j])
			}
			cell, _ := excelize.CoordinatesToCellName(j+1, i+2)
			// SetCellStr preserves formula-like input as literal text.
			if err := f.SetCellStr(sheetName, cell, value); err != nil {
				return err
			}
		}
		if err := f.SetCellInt(sheetName, fmt.Sprintf("E%d", i+2), int64(p.ID)); err != nil {
			return err
		}
	}
	if err := formatWorkbook(f, len(people)+1); err != nil {
		return err
	}

	output, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("create output file (existing files are not overwritten): %w", err)
	}
	defer func() {
		err = errors.Join(err, output.Close())
		if err != nil {
			_ = os.Remove(path)
		}
	}()
	if err := f.Write(output); err != nil {
		return fmt.Errorf("write XLSX: %w", err)
	}
	return output.Sync()
}

func formatWorkbook(f *excelize.File, lastRow int) error {
	bodyStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: "Calibri", Size: 11},
		Alignment: &excelize.Alignment{Vertical: "top", WrapText: true},
	})
	if err != nil {
		return err
	}
	if err := f.SetCellStyle(sheetName, "A1", fmt.Sprintf("E%d", lastRow), bodyStyle); err != nil {
		return err
	}
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Family: "Calibri", Size: 11, Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"1F4E78"}},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	if err != nil {
		return err
	}
	if err := f.SetCellStyle(sheetName, "A1", "E1", headerStyle); err != nil {
		return err
	}
	if err := f.SetRowHeight(sheetName, 1, 24); err != nil {
		return err
	}
	for _, col := range []struct {
		name  string
		width float64
	}{{"A", 26}, {"B", 22}, {"C", 26}, {"D", 72}, {"E", 12}} {
		if err := f.SetColWidth(sheetName, col.name, col.name, col.width); err != nil {
			return err
		}
	}
	if err := f.SetPanes(sheetName, &excelize.Panes{Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"}); err != nil {
		return err
	}
	return f.AutoFilter(sheetName, fmt.Sprintf("A1:E%d", lastRow), nil)
}
