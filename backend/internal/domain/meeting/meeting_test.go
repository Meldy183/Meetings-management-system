package meeting

import (
	"testing"
	"time"

	"meetings-editor/internal/domain/person"
)

var (
	alice = person.Person{ID: 1, LastName: "Иванова", FirstName: "Алиса"}
	bob   = person.Person{ID: 2, LastName: "Петров", FirstName: "Борис"}
)

func baseMeeting() *Meeting {
	return &Meeting{
		ID:    "00000000-0000-0000-0000-000000000001",
		Title: "Test",
		Date:  time.Now(),
		Chairperson: &alice,
		People:      []person.Person{alice, bob},
		AgendaItems: []AgendaItem{
			{ID: 1, Text: "Item 1", Speakers: []person.Person{bob}},
		},
		CreatedAt: time.Now(),
	}
}

func TestStatus_Complete(t *testing.T) {
	m := baseMeeting()
	if got := m.Status(); got != "complete" {
		t.Errorf("want complete, got %q", got)
	}
}

func TestStatus_NoChairperson(t *testing.T) {
	m := baseMeeting()
	m.Chairperson = nil
	if got := m.Status(); got != "incomplete" {
		t.Errorf("want incomplete, got %q", got)
	}
}

func TestStatus_NoPeople(t *testing.T) {
	m := baseMeeting()
	m.People = nil
	if got := m.Status(); got != "incomplete" {
		t.Errorf("want incomplete, got %q", got)
	}
}

func TestStatus_EmptyPeopleSlice(t *testing.T) {
	m := baseMeeting()
	m.People = []person.Person{}
	if got := m.Status(); got != "incomplete" {
		t.Errorf("want incomplete, got %q", got)
	}
}

func TestStatus_NoAgendaItems(t *testing.T) {
	m := baseMeeting()
	m.AgendaItems = nil
	if got := m.Status(); got != "incomplete" {
		t.Errorf("want incomplete, got %q", got)
	}
}

func TestStatus_EmptyAgendaItemsSlice(t *testing.T) {
	m := baseMeeting()
	m.AgendaItems = []AgendaItem{}
	if got := m.Status(); got != "incomplete" {
		t.Errorf("want incomplete, got %q", got)
	}
}
