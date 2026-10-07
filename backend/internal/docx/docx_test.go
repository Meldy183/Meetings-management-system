package docx

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"strings"
	"testing"
	"time"

	domMeeting "meetings-editor/internal/domain/meeting"
	"meetings-editor/internal/domain/person"
)

func participantTableText(t *testing.T, document string) []string {
	t.Helper()
	decoder := xml.NewDecoder(strings.NewReader(document))
	var tables []string
	depth := 0
	var current strings.Builder
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("parse document XML: %v", err)
		}
		switch token := token.(type) {
		case xml.StartElement:
			if token.Name.Local == "tbl" {
				depth++
				if depth == 1 {
					current.Reset()
				}
			}
		case xml.CharData:
			if depth > 0 {
				current.Write(token)
			}
		case xml.EndElement:
			if token.Name.Local == "tbl" && depth > 0 {
				depth--
				if depth == 0 {
					tables = append(tables, current.String())
				}
			}
		}
	}
	return tables
}

var (
	testPerson1 = person.Person{ID: 1, LastName: "Иванов", FirstName: "Иван", MiddleName: "Иванович", Info: "Директор"}
	testPerson2 = person.Person{ID: 2, LastName: "Петрова", FirstName: "Мария", MiddleName: "", Info: "Зам. директора"}
)

func completeMeeting() *domMeeting.Meeting {
	return &domMeeting.Meeting{
		ID:          "00000000-0000-0000-0000-000000000001",
		Title:       "совещание по тестам",
		Date:        time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC),
		Chairperson: &testPerson1,
		People:      []domMeeting.MeetingPerson{{Person: testPerson1, AttendanceMode: domMeeting.AttendanceModeInPerson}, {Person: testPerson2, AttendanceMode: domMeeting.AttendanceModeVCS}},
		AgendaItems: []domMeeting.AgendaItem{
			{ID: 1, Text: "Первый вопрос", Speakers: []person.Person{testPerson2}},
			{ID: 2, Text: "Второй вопрос", Speakers: []person.Person{testPerson1, testPerson2}},
		},
	}
}

// parseDocx opens the ZIP bytes and returns a map filename→content.
func parseDocx(t *testing.T, data []byte) map[string]string {
	t.Helper()
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("parse docx zip: %v", err)
	}
	files := make(map[string]string, len(r.File))
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open zip entry %q: %v", f.Name, err)
		}
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(rc)
		rc.Close()
		files[f.Name] = buf.String()
	}
	return files
}

// --- Agenda ---

func TestAgenda_ReturnsValidDocx(t *testing.T) {
	g := New()
	data, err := g.Agenda(completeMeeting())
	if err != nil {
		t.Fatalf("Agenda: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty bytes")
	}
	files := parseDocx(t, data)
	for _, required := range []string{
		"[Content_Types].xml",
		"_rels/.rels",
		"word/_rels/document.xml.rels",
		"word/document.xml",
	} {
		if _, ok := files[required]; !ok {
			t.Errorf("missing required entry %q in docx archive", required)
		}
	}
}

func TestAgenda_ContainsMeetingTitle(t *testing.T) {
	g := New()
	m := completeMeeting()
	data, _ := g.Agenda(m)
	files := parseDocx(t, data)
	doc := files["word/document.xml"]
	if !strings.Contains(doc, "совещание по тестам") {
		t.Error("document.xml should contain meeting title")
	}
}

func TestParticipants_GroupsAndNumbersAttendanceModes(t *testing.T) {
	g := New()
	m := completeMeeting()
	third := person.Person{ID: 3, LastName: "Сидоров", FirstName: "Семён"}
	m.People = append(m.People, domMeeting.MeetingPerson{Person: third, AttendanceMode: domMeeting.AttendanceModeInPerson})
	data, err := g.Participants(m)
	if err != nil {
		t.Fatal(err)
	}
	document := parseDocx(t, data)["word/document.xml"]
	tables := participantTableText(t, document)
	if len(tables) != 2 {
		t.Fatalf("want two participant tables, got %d", len(tables))
	}
	if strings.Index(document, ">Очно<") > strings.Index(document, ">ВКС<") {
		t.Fatal("attendance groups are in the wrong order")
	}
	if strings.Count(document, `<w:color w:val="0070C0"/>`) != 2 {
		t.Fatal("only the in-person and VCS headings should use blue text")
	}
	if !strings.Contains(document, `<w:spacing w:before="240" w:after="0" w:line="240" w:lineRule="auto"/><w:keepNext/>`) {
		t.Fatal("attendance group headings should have a larger gap and stay with their table")
	}
	if strings.Contains(tables[0], "ИВАНОВ") || !strings.Contains(tables[0], "СИДОРОВ") || !strings.Contains(tables[1], "ПЕТРОВА") {
		t.Fatalf("participants are assigned to unexpected tables: %#v", tables)
	}
	for i, table := range tables {
		if !strings.Contains(table, "1.") {
			t.Errorf("table %d should restart numbering at 1", i)
		}
	}
}

func TestParticipants_OmitsEmptyGroupsAndChairperson(t *testing.T) {
	g := New()
	m := completeMeeting()
	m.People = []domMeeting.MeetingPerson{{Person: testPerson1, AttendanceMode: domMeeting.AttendanceModeInPerson}}
	data, err := g.Participants(m)
	if err != nil {
		t.Fatal(err)
	}
	document := parseDocx(t, data)["word/document.xml"]
	if tables := participantTableText(t, document); len(tables) != 0 {
		t.Fatalf("chairperson should not be printed in participant tables: %v", tables)
	}
	if strings.Contains(document, ">Очно<") || strings.Contains(document, ">ВКС<") {
		t.Fatal("empty attendance group headings should be omitted")
	}
}

func TestAgenda_ContainsChairpersonName(t *testing.T) {
	g := New()
	data, _ := g.Agenda(completeMeeting())
	files := parseDocx(t, data)
	doc := files["word/document.xml"]
	if !strings.Contains(doc, "Иванов") {
		t.Error("document.xml should contain chairperson last name")
	}
}

func TestAgenda_NilChairperson_NoError(t *testing.T) {
	g := New()
	m := completeMeeting()
	m.Chairperson = nil
	_, err := g.Agenda(m)
	if err != nil {
		t.Fatalf("Agenda with nil chairperson should not fail: %v", err)
	}
}

func TestAgenda_SingleSpeakerLabel(t *testing.T) {
	g := New()
	m := completeMeeting()
	// Item 1 has one speaker
	data, _ := g.Agenda(m)
	files := parseDocx(t, data)
	doc := files["word/document.xml"]
	if !strings.Contains(doc, "Докладчик:") {
		t.Error("should contain singular label for item with one speaker")
	}
}

func TestAgenda_MultipleSpeakersLabel(t *testing.T) {
	g := New()
	m := completeMeeting()
	// Item 2 has two speakers
	data, _ := g.Agenda(m)
	files := parseDocx(t, data)
	doc := files["word/document.xml"]
	if !strings.Contains(doc, "Докладчики:") {
		t.Error("should contain plural label for item with multiple speakers")
	}
}

func TestAgenda_RomanNumerals(t *testing.T) {
	g := New()
	data, _ := g.Agenda(completeMeeting())
	files := parseDocx(t, data)
	doc := files["word/document.xml"]
	if !strings.Contains(doc, "I.") || !strings.Contains(doc, "II.") {
		t.Error("agenda items should be prefixed with Roman numerals")
	}
}

// --- Participants ---

func TestParticipants_ReturnsValidDocx(t *testing.T) {
	g := New()
	data, err := g.Participants(completeMeeting())
	if err != nil {
		t.Fatalf("Participants: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty bytes")
	}
	files := parseDocx(t, data)
	if _, ok := files["word/document.xml"]; !ok {
		t.Error("missing word/document.xml")
	}
}

func TestParticipants_ContainsAllPeople(t *testing.T) {
	g := New()
	data, _ := g.Participants(completeMeeting())
	files := parseDocx(t, data)
	doc := files["word/document.xml"]
	for _, name := range []string{"ПЕТРОВА"} {
		if !strings.Contains(doc, name) {
			t.Errorf("document.xml should contain participant %q", name)
		}
	}
}

func TestParticipants_NilChairperson_NoError(t *testing.T) {
	g := New()
	m := completeMeeting()
	m.Chairperson = nil
	_, err := g.Participants(m)
	if err != nil {
		t.Fatalf("Participants with nil chairperson should not fail: %v", err)
	}
}

// --- toRoman ---

func TestToRoman(t *testing.T) {
	cases := []struct {
		n    int
		want string
	}{
		{1, "I"},
		{2, "II"},
		{3, "III"},
		{4, "IV"},
		{5, "V"},
		{9, "IX"},
		{10, "X"},
		{14, "XIV"},
		{40, "XL"},
		{50, "L"},
		{90, "XC"},
		{100, "C"},
		{400, "CD"},
		{500, "D"},
		{900, "CM"},
		{1000, "M"},
		{1994, "MCMXCIV"},
	}
	for _, tc := range cases {
		if got := toRoman(tc.n); got != tc.want {
			t.Errorf("toRoman(%d) = %q, want %q", tc.n, got, tc.want)
		}
	}
}

// --- fullName ---

func TestFullName_WithMiddleName(t *testing.T) {
	p := person.Person{LastName: "Иванов", FirstName: "Иван", MiddleName: "Иванович"}
	want := "Иванов Иван Иванович"
	if got := fullName(p); got != want {
		t.Errorf("fullName = %q, want %q", got, want)
	}
}

func TestFullName_WithoutMiddleName(t *testing.T) {
	p := person.Person{LastName: "Петров", FirstName: "Пётр", MiddleName: ""}
	want := "Петров Пётр"
	if got := fullName(p); got != want {
		t.Errorf("fullName = %q, want %q", got, want)
	}
}

// --- formatDate ---

func TestFormatDate(t *testing.T) {
	t1 := time.Date(2026, 3, 21, 9, 30, 0, 0, time.UTC)
	got := formatDate(t1)
	if !strings.Contains(got, "21") || !strings.Contains(got, "марта") || !strings.Contains(got, "2026") {
		t.Errorf("formatDate(%v) = %q, expected day/month/year", t1, got)
	}
	if !strings.Contains(got, "09") || !strings.Contains(got, "30") {
		t.Errorf("formatDate(%v) = %q, expected time HH:MM", t1, got)
	}
}

// --- xmlEscape ---

func TestXmlEscape(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"hello", "hello"},
		{"a&b", "a&amp;b"},
		{"<tag>", "&lt;tag&gt;"},
		{`a"b`, "a&quot;b"},
	}
	for _, tc := range cases {
		if got := xmlEscape(tc.in); got != tc.want {
			t.Errorf("xmlEscape(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
