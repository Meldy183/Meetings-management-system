// Package docx generates .docx (OOXML) files in memory without external dependencies.
package docx

import (
	"archive/zip"
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"

	domMeeting "meetings-editor/internal/domain/meeting"
	"meetings-editor/internal/domain/person"
)

// Generator implements handler.ExportService.
type Generator struct{}

func New() *Generator { return &Generator{} }

// Agenda generates the "Повестка" .docx document.
func (g *Generator) Agenda(m *domMeeting.Meeting) ([]byte, error) {
	var body strings.Builder

	// Title block
	header := tnrBold("ПОВЕСТКА", 28) +
		lineBreak() + tnrBold(lcFirst(m.Title)+" под председательством ", 28)
	if m.Chairperson != nil {
		if m.Chairperson.Info != "" {
			header += lineBreak() + tnrBold(m.Chairperson.Info, 28)
		}
		header += lineBreak() + tnrBold(shortName(*m.Chairperson), 28)
	}
	body.WriteString(para(pPrCenter() + header))
	body.WriteString(para(pPrRight() + tnrBold(formatDate(m.Date), 28)))
	if m.Place != "" {
		body.WriteString(para(pPrRight() + tnr(m.Place, 28)))
	}
	body.WriteString(para(pPrLeft())) // blank line between header and agenda items

	// Agenda items
	for i, item := range m.AgendaItems {
		roman := toRoman(i + 1)
		body.WriteString(para(pPrLeft() + tnrBold(roman+". "+item.Text, 27)))
		label := "Докладчик:"
		if len(item.Speakers) > 1 {
			label = "Докладчики:"
		}
		body.WriteString(para(pPrCenterSpaced() + tnrBoldUnderline(label, 27)))
		for _, spk := range item.Speakers {
			body.WriteString(agendaTable(spk))
		}
		body.WriteString(para(pPrLeft())) // spacer between items
	}

	return buildDocx(body.String())
}

// Participants generates the "Список участников" .docx document.
func (g *Generator) Participants(m *domMeeting.Meeting) ([]byte, error) {
	var body strings.Builder

	// Title block
	pHeader := tnrBold("СПИСОК", 28) +
		lineBreak() + tnrBold("участников "+lcFirst(m.Title)+" под председательством ", 28)
	if m.Chairperson != nil {
		if m.Chairperson.Info != "" {
			pHeader += lineBreak() + tnrBold(m.Chairperson.Info, 28)
		}
		pHeader += lineBreak() + tnrBold(shortName(*m.Chairperson), 28)
	}
	body.WriteString(para(pPrCenter() + pHeader))
	body.WriteString(para(pPrRight() + tnrBold(formatDate(m.Date), 28)))
	if m.Place != "" {
		body.WriteString(para(pPrRight() + tnr(m.Place, 28)))
	}
	body.WriteString(para(pPrLeft())) // blank line before table
	people := m.People
	if m.Chairperson != nil {
		people = make([]person.Person, 0, len(m.People))
		for _, p := range m.People {
			if p.ID != m.Chairperson.ID {
				people = append(people, p)
			}
		}
	}
	body.WriteString(participantsTable(people))

	return buildDocx(body.String())
}

// --- paragraph / run helpers ---

func para(content string) string {
	return `<w:p>` + content + `</w:p>`
}

// pPrCenter returns centered paragraph properties with standard line spacing.
func pPrCenter() string {
	return `<w:pPr><w:spacing w:after="0" w:line="240" w:lineRule="auto"/><w:jc w:val="center"/></w:pPr>`
}

// pPrRight returns right-aligned paragraph properties with standard line spacing.
func pPrRight() string {
	return `<w:pPr><w:spacing w:after="0" w:line="240" w:lineRule="auto"/><w:jc w:val="right"/></w:pPr>`
}

// pPrLeft returns left-aligned paragraph properties with standard line spacing.
func pPrLeft() string {
	return `<w:pPr><w:spacing w:after="0" w:line="240" w:lineRule="auto"/></w:pPr>`
}

// pPrCenterSpaced returns centered paragraph properties with 120-twip spacing above and below.
func pPrCenterSpaced() string {
	return `<w:pPr><w:spacing w:before="120" w:after="120" w:line="240" w:lineRule="auto"/><w:jc w:val="center"/></w:pPr>`
}

// pPrLeftSpaced returns left-aligned paragraph properties with 120-twip spacing before.
func pPrLeftSpaced() string {
	return `<w:pPr><w:spacing w:before="120" w:after="0" w:line="240" w:lineRule="auto"/></w:pPr>`
}

// tnr produces a Times New Roman run at the given half-point size.
func tnr(s string, size int) string {
	return fmt.Sprintf(
		`<w:r><w:rPr><w:rFonts w:ascii="Times New Roman" w:hAnsi="Times New Roman" w:cs="Times New Roman"/><w:sz w:val="%d"/><w:szCs w:val="%d"/></w:rPr><w:t xml:space="preserve">%s</w:t></w:r>`,
		size, size, xmlEscape(s),
	)
}

// tnrBold produces a bold Times New Roman run.
func tnrBold(s string, size int) string {
	return fmt.Sprintf(
		`<w:r><w:rPr><w:rFonts w:ascii="Times New Roman" w:hAnsi="Times New Roman" w:cs="Times New Roman"/><w:b/><w:sz w:val="%d"/><w:szCs w:val="%d"/></w:rPr><w:t xml:space="preserve">%s</w:t></w:r>`,
		size, size, xmlEscape(s),
	)
}

// tnrBoldUnderline produces a bold, single-underlined Times New Roman run.
func tnrBoldUnderline(s string, size int) string {
	return fmt.Sprintf(
		`<w:r><w:rPr><w:rFonts w:ascii="Times New Roman" w:hAnsi="Times New Roman" w:cs="Times New Roman"/><w:b/><w:u w:val="single"/><w:sz w:val="%d"/><w:szCs w:val="%d"/></w:rPr><w:t xml:space="preserve">%s</w:t></w:r>`,
		size, size, xmlEscape(s),
	)
}

// lineBreak produces an inline line-break run (w:br inside w:r).
func lineBreak() string {
	return `<w:r><w:br/></w:r>`
}

// tnrCell produces a paragraph suitable for use inside a table cell.
func tnrCell(s string, size int) string {
	return `<w:p>` + pPrLeft() + tnr(s, size) + `</w:p>`
}

// tnrCellNameSplit renders LASTNAME on line 1 and "Firstname Patronymic" on line 2
// inside a single paragraph using a line break. Includes spacing before for participant rows.
func tnrCellNameSplit(p person.Person, size int) string {
	lastName := strings.ToUpper(p.LastName)
	firstMid := strings.TrimSpace(p.FirstName + " " + p.MiddleName)
	rProps := fmt.Sprintf(
		`<w:rPr><w:rFonts w:ascii="Times New Roman" w:hAnsi="Times New Roman" w:cs="Times New Roman"/><w:sz w:val="%d"/><w:szCs w:val="%d"/></w:rPr>`,
		size, size,
	)
	return fmt.Sprintf(
		`<w:p>%s<w:r>%s<w:t>%s</w:t></w:r><w:r>%s<w:br/><w:t>%s</w:t></w:r></w:p>`,
		pPrLeft(), rProps, xmlEscape(lastName), rProps, xmlEscape(firstMid),
	)
}

// --- table helpers ---

// agendaTable renders a borderless 3-column table: name | "–" | info.
func agendaTable(sp person.Person) string {
	// Column widths (dxa): name=4000, dash=300, info=5054. Total≈9354 (A4 text width).
	name := strings.ToUpper(sp.LastName) + "\n" + strings.TrimSpace(sp.FirstName+" "+sp.MiddleName)
	nameRProps := `<w:rPr><w:rFonts w:ascii="Times New Roman" w:hAnsi="Times New Roman" w:cs="Times New Roman"/><w:sz w:val="28"/><w:szCs w:val="28"/></w:rPr>`
	nameParts := strings.SplitN(name, "\n", 2)
	nameCell := fmt.Sprintf(
		`<w:p>%s<w:r>%s<w:t>%s</w:t></w:r><w:r>%s<w:br/><w:t>%s</w:t></w:r></w:p>`,
		pPrLeft(), nameRProps, xmlEscape(nameParts[0]), nameRProps, xmlEscape(nameParts[1]),
	)

	return fmt.Sprintf(`
<w:tbl>
  <w:tblPr>
    <w:tblW w:w="9354" w:type="dxa"/>
    <w:tblBorders>
      <w:top w:val="none" w:sz="0" w:space="0" w:color="auto"/>
      <w:left w:val="none" w:sz="0" w:space="0" w:color="auto"/>
      <w:bottom w:val="none" w:sz="0" w:space="0" w:color="auto"/>
      <w:right w:val="none" w:sz="0" w:space="0" w:color="auto"/>
      <w:insideH w:val="none" w:sz="0" w:space="0" w:color="auto"/>
      <w:insideV w:val="none" w:sz="0" w:space="0" w:color="auto"/>
    </w:tblBorders>
  </w:tblPr>
  <w:tr>
    <w:tc><w:tcPr><w:tcW w:w="4000" w:type="dxa"/><w:tcMar><w:top w:w="80" w:type="dxa"/><w:bottom w:w="80" w:type="dxa"/></w:tcMar></w:tcPr>%s</w:tc>
    <w:tc><w:tcPr><w:tcW w:w="300" w:type="dxa"/><w:tcMar><w:top w:w="80" w:type="dxa"/><w:bottom w:w="80" w:type="dxa"/></w:tcMar></w:tcPr>%s</w:tc>
    <w:tc><w:tcPr><w:tcW w:w="5054" w:type="dxa"/><w:tcMar><w:top w:w="80" w:type="dxa"/><w:bottom w:w="80" w:type="dxa"/></w:tcMar></w:tcPr>%s</w:tc>
  </w:tr>
</w:tbl>`, nameCell, tnrCell("—", 28), tnrCell(sp.Info, 28))
}

// participantsTable renders a borderless 4-column table: № | name | "-" | info.
func participantsTable(participants []person.Person) string {
	var sb strings.Builder
	sb.WriteString(`
<w:tbl>
  <w:tblPr>
    <w:tblW w:w="9810" w:type="dxa"/>
    <w:tblBorders>
      <w:top w:val="none" w:sz="0" w:space="0" w:color="auto"/>
      <w:left w:val="none" w:sz="0" w:space="0" w:color="auto"/>
      <w:bottom w:val="none" w:sz="0" w:space="0" w:color="auto"/>
      <w:right w:val="none" w:sz="0" w:space="0" w:color="auto"/>
      <w:insideH w:val="none" w:sz="0" w:space="0" w:color="auto"/>
      <w:insideV w:val="none" w:sz="0" w:space="0" w:color="auto"/>
    </w:tblBorders>
  </w:tblPr>`)

	for i, p := range participants {
		numCell := `<w:p>` + pPrLeft() + tnr(fmt.Sprintf("%d.", i+1), 28) + `</w:p>`
		dashCell := `<w:p>` + pPrLeft() + tnr("—", 28) + `</w:p>`
		infoCell := `<w:p>` + pPrLeft() + tnr(p.Info, 28) + `</w:p>`
		sb.WriteString(fmt.Sprintf(`
  <w:tr>
    <w:tc><w:tcPr><w:tcW w:w="566" w:type="dxa"/><w:tcMar><w:top w:w="80" w:type="dxa"/><w:bottom w:w="80" w:type="dxa"/></w:tcMar></w:tcPr>%s</w:tc>
    <w:tc><w:tcPr><w:tcW w:w="3742" w:type="dxa"/><w:tcMar><w:top w:w="80" w:type="dxa"/><w:bottom w:w="80" w:type="dxa"/></w:tcMar></w:tcPr>%s</w:tc>
    <w:tc><w:tcPr><w:tcW w:w="323" w:type="dxa"/><w:tcMar><w:top w:w="80" w:type="dxa"/><w:bottom w:w="80" w:type="dxa"/></w:tcMar></w:tcPr>%s</w:tc>
    <w:tc><w:tcPr><w:tcW w:w="5179" w:type="dxa"/><w:tcMar><w:top w:w="80" w:type="dxa"/><w:bottom w:w="80" w:type="dxa"/></w:tcMar></w:tcPr>%s</w:tc>
  </w:tr>`,
			numCell,
			tnrCellNameSplit(p, 27),
			dashCell,
			infoCell,
		))
	}

	sb.WriteString(`</w:tbl>`)
	return sb.String()
}

// --- DOCX zip assembly ---

var documentTmpl = template.Must(template.New("doc").Parse(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:wpc="http://schemas.microsoft.com/office/word/2010/wordprocessingCanvas"
            xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"
            xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <w:body>
    {{.Body}}
    <w:sectPr>
      <w:pgSz w:w="11906" w:h="16838"/>
      <w:pgMar w:top="567" w:right="851" w:bottom="567" w:left="1701" w:header="709" w:footer="709" w:gutter="0"/>
    </w:sectPr>
  </w:body>
</w:document>`))

const contentTypes = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml"
    ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`

const rootRels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1"
    Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument"
    Target="word/document.xml"/>
</Relationships>`

const wordRels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"/>`

func buildDocx(body string) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	files := map[string]string{
		"[Content_Types].xml":          contentTypes,
		"_rels/.rels":                  rootRels,
		"word/_rels/document.xml.rels": wordRels,
	}

	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			return nil, err
		}
		if _, err := fmt.Fprint(w, content); err != nil {
			return nil, err
		}
	}

	docWriter, err := zw.Create("word/document.xml")
	if err != nil {
		return nil, err
	}
	if err := documentTmpl.Execute(docWriter, struct{ Body string }{Body: body}); err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// --- utilities ---

func fullName(p person.Person) string {
	name := p.LastName + " " + p.FirstName
	if p.MiddleName != "" {
		name += " " + p.MiddleName
	}
	return name
}

// shortName returns "И.О. Фамилия" — initials (no space between them) then full surname.
func shortName(p person.Person) string {
	var initials string
	if r := []rune(p.FirstName); len(r) > 0 {
		initials += string(r[0]) + "."
	}
	if r := []rune(p.MiddleName); len(r) > 0 {
		initials += string(r[0]) + "."
	}
	if initials == "" {
		return p.LastName
	}
	return initials + " " + p.LastName
}

func formatDate(t time.Time) string {
	months := []string{
		"января", "февраля", "марта", "апреля", "мая", "июня",
		"июля", "августа", "сентября", "октября", "ноября", "декабря",
	}
	return fmt.Sprintf("%d %s %d г., %02d:%02d",
		t.Day(), months[t.Month()-1], t.Year(), t.Hour(), t.Minute())
}

// lcFirst lowercases the first rune of s, leaving the rest unchanged.
func lcFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = []rune(strings.ToLower(string(r[0])))[0]
	return string(r)
}

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}

// toRoman converts a positive integer to an uppercase Roman numeral string.
func toRoman(n int) string {
	vals := []int{1000, 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1}
	syms := []string{"M", "CM", "D", "CD", "C", "XC", "L", "XL", "X", "IX", "V", "IV", "I"}
	var result strings.Builder
	for i, v := range vals {
		for n >= v {
			result.WriteString(syms[i])
			n -= v
		}
	}
	return result.String()
}
