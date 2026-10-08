package registryexport

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"

	"meetings-editor/internal/domain/person"
)

func TestWriteXLSXPreservesAllRecordsAndLiteralText(t *testing.T) {
	people := []person.Person{
		{ID: 7, LastName: "Маркин", FirstName: "Федор", MiddleName: "Сергеевич", Info: "Студент\nВторой курс"},
		{ID: 9, LastName: "Маркин", FirstName: "Федор", MiddleName: "Сергеевич", Info: "=SUM(1,2)"},
		{ID: 11, LastName: "  Иванова  ", FirstName: "Анна", Info: "+literal @text"},
	}
	path := filepath.Join(t.TempDir(), "participants.xlsx")
	if err := WriteXLSX(people, path); err != nil {
		t.Fatal(err)
	}
	f, err := excelize.OpenFile(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if got := f.GetSheetList(); !reflect.DeepEqual(got, []string{sheetName}) {
		t.Fatalf("unexpected sheets: %v", got)
	}
	rows, err := f.GetRows(sheetName)
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{
		{"Фамилия", "Имя", "Отчество", "Должность"},
		{"Маркин", "Федор", "Сергеевич", "Студент\nВторой курс"},
		{"Маркин", "Федор", "Сергеевич", "=SUM(1,2)"},
		{"  Иванова  ", "Анна", "", "+literal @text"},
	}
	if !reflect.DeepEqual(rows, want) {
		t.Fatalf("registry data changed: got %#v, want %#v", rows, want)
	}
	formula, err := f.GetCellFormula(sheetName, "D3")
	if err != nil || formula != "" {
		t.Fatalf("formula-like data must remain text: formula=%q, error=%v", formula, err)
	}
}

func TestWriteXLSXMatchesExampleLayout(t *testing.T) {
	// Keep a copy of the example in testdata so this test also runs in Docker.
	examplePath := filepath.Join("testdata", "participants_example.xlsx")
	example, err := excelize.OpenFile(examplePath)
	if err != nil {
		t.Fatal(err)
	}
	defer example.Close()
	exampleSheet := example.GetSheetName(0)
	exampleRows, err := example.GetRows(exampleSheet)
	if err != nil {
		t.Fatal(err)
	}
	var people []person.Person
	for i, row := range exampleRows[1:] {
		people = append(people, person.Person{ID: i + 1, LastName: row[0], FirstName: row[1], MiddleName: row[2], Info: row[3]})
	}
	path := filepath.Join(t.TempDir(), "participants.xlsx")
	if err := WriteXLSX(people, path); err != nil {
		t.Fatal(err)
	}
	f, err := excelize.OpenFile(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if got := f.GetSheetList(); !reflect.DeepEqual(got, example.GetSheetList()) {
		t.Fatalf("sheet should match the example, got %v", got)
	}
	rows, err := f.GetRows(sheetName)
	if err != nil || !reflect.DeepEqual(rows, exampleRows) {
		t.Fatalf("export should match the example's four-column format: rows=%v, error=%v", rows, err)
	}
	for _, col := range []string{"A", "B", "C", "D"} {
		want, err := example.GetColWidth(exampleSheet, col)
		if err != nil {
			t.Fatal(err)
		}
		got, err := f.GetColWidth(sheetName, col)
		if err != nil || got != want {
			t.Errorf("column %s width=%v, want %v, error=%v", col, got, want, err)
		}
	}
	for row := 1; row <= len(exampleRows); row++ {
		want, err := example.GetRowHeight(exampleSheet, row)
		if err != nil {
			t.Fatal(err)
		}
		got, err := f.GetRowHeight(sheetName, row)
		if err != nil || got != want {
			t.Errorf("row %d height=%v, want %v, error=%v", row, got, want, err)
		}
	}
	exampleStyleID, err := example.GetCellStyle(exampleSheet, "A1")
	if err != nil {
		t.Fatal(err)
	}
	exampleStyle, err := example.GetStyle(exampleStyleID)
	if err != nil {
		t.Fatal(err)
	}
	styleID, err := f.GetCellStyle(sheetName, "A1")
	if err != nil {
		t.Fatal(err)
	}
	style, err := f.GetStyle(styleID)
	if err != nil {
		t.Fatal(err)
	}
	font, err := f.GetDefaultFont()
	if err != nil {
		t.Fatal(err)
	}
	exampleFont, err := example.GetDefaultFont()
	if err != nil {
		t.Fatal(err)
	}
	if font != exampleFont {
		t.Fatalf("default font should match the example: got %q, want %q", font, exampleFont)
	}
	// The example inherits its normal font; Excelize may return nil for it.
	bold := style.Font != nil && style.Font.Bold
	exampleBold := exampleStyle.Font != nil && exampleStyle.Font.Bold
	if bold != exampleBold {
		t.Fatal("header font weight should match the example")
	}
	if style.Fill.Pattern != exampleStyle.Fill.Pattern || !reflect.DeepEqual(style.Fill.Color, exampleStyle.Fill.Color) {
		t.Fatalf("header fill should match the plain example: %#v", style.Fill)
	}
	wrap := style.Alignment != nil && style.Alignment.WrapText
	exampleWrap := exampleStyle.Alignment != nil && exampleStyle.Alignment.WrapText
	if wrap != exampleWrap {
		t.Fatal("text wrapping should match the example")
	}
	for _, element := range []string{"autoFilter", "pane"} {
		if sheetContainsElement(t, path, element) != sheetContainsElement(t, examplePath, element) {
			t.Errorf("worksheet element %s should match the example", element)
		}
	}
}

func sheetContainsElement(t *testing.T, path, name string) bool {
	t.Helper()
	z, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	for _, entry := range z.File {
		if entry.Name != "xl/worksheets/sheet1.xml" {
			continue
		}
		r, err := entry.Open()
		if err != nil {
			t.Fatal(err)
		}
		document, err := io.ReadAll(r)
		r.Close()
		if err != nil {
			t.Fatal(err)
		}
		return strings.Contains(string(document), "<"+name)
	}
	t.Fatal("worksheet XML not found")
	return false
}

func TestWriteXLSXEmptyRegistry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.xlsx")
	if err := WriteXLSX(nil, path); err != nil {
		t.Fatal(err)
	}
	f, err := excelize.OpenFile(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rows, err := f.GetRows(sheetName)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rows, [][]string{headers}) {
		t.Fatalf("empty registry should contain only headers, got %v", rows)
	}
}

func TestWriteXLSXDoesNotOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "existing.xlsx")
	original := []byte("existing data")
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	if err := WriteXLSX(nil, path); err == nil {
		t.Fatal("expected refusal to overwrite existing output")
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != string(original) {
		t.Fatalf("existing file changed: contents=%q error=%v", got, err)
	}
}

func TestWriteXLSXRejectsTruncationAndWrongExtension(t *testing.T) {
	for _, tc := range []struct {
		name   string
		people []person.Person
	}{
		{"too-long.xlsx", []person.Person{{ID: 1, Info: strings.Repeat("я", 32768)}}},
		{"wrong.csv", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), tc.name)
			if err := WriteXLSX(tc.people, path); err == nil {
				t.Fatal("expected validation error")
			}
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Fatalf("failed export should not leave a file: %v", err)
			}
		})
	}
}
