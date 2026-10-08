package registryexport

import (
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
		{"Фамилия", "Имя", "Отчество", "Информация", "ID"},
		{"Маркин", "Федор", "Сергеевич", "Студент\nВторой курс", "7"},
		{"Маркин", "Федор", "Сергеевич", "=SUM(1,2)", "9"},
		{"  Иванова  ", "Анна", "", "+literal @text", "11"},
	}
	if !reflect.DeepEqual(rows, want) {
		t.Fatalf("registry data changed: got %#v, want %#v", rows, want)
	}
	formula, err := f.GetCellFormula(sheetName, "D3")
	if err != nil || formula != "" {
		t.Fatalf("formula-like data must remain text: formula=%q, error=%v", formula, err)
	}
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
