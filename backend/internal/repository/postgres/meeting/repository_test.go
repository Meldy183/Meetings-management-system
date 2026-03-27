package meeting

import (
	"testing"
	"time"

	"meetings-editor/internal/domain/meeting"
	"meetings-editor/internal/domain/person"
	"meetings-editor/internal/testutil"
	"meetings-editor/pkg/errs"

	personRepo "meetings-editor/internal/repository/postgres/person"
)

// seedPerson inserts a participant and returns it.
func seedPerson(t *testing.T, repo person.Repository, last, first string) person.Person {
	t.Helper()
	p, err := repo.Create(testutil.Ctx(), &person.Person{LastName: last, FirstName: first})
	if err != nil {
		t.Fatalf("seed person %s %s: %v", last, first, err)
	}
	return *p
}

// seedMeeting inserts a bare meeting and returns it.
func seedMeeting(t *testing.T, repo meeting.Repository, title string) *meeting.Meeting {
	t.Helper()
	m := &meeting.Meeting{Title: title, Date: time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)}
	created, err := repo.Create(testutil.Ctx(), m)
	if err != nil {
		t.Fatalf("seed meeting %q: %v", title, err)
	}
	return created
}

func TestMeetingRepo_CreateAndGetByID(t *testing.T) {
	pool := testutil.NewDB(t)
	testutil.TruncateTables(t, pool)
	repo := New(pool)
	ctx := testutil.Ctx()

	m := seedMeeting(t, repo, "Тест")

	if m.ID == "" {
		t.Fatal("expected non-empty ID")
	}

	got, err := repo.GetByID(ctx, m.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Title != "Тест" {
		t.Errorf("want title 'Тест', got %q", got.Title)
	}
	if got.Chairperson != nil {
		t.Error("expected nil chairperson")
	}
}

func TestMeetingRepo_GetByID_NotFound(t *testing.T) {
	pool := testutil.NewDB(t)
	testutil.TruncateTables(t, pool)
	repo := New(pool)
	ctx := testutil.Ctx()

	_, err := repo.GetByID(ctx, "00000000-0000-0000-0000-000000000000")
	if err != errs.ErrNotFound {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestMeetingRepo_Update(t *testing.T) {
	pool := testutil.NewDB(t)
	testutil.TruncateTables(t, pool)
	repo := New(pool)
	ctx := testutil.Ctx()

	m := seedMeeting(t, repo, "Оригинал")
	newDate := time.Date(2027, 1, 1, 9, 0, 0, 0, time.UTC)

	if err := repo.Update(ctx, m.ID, "Изменён", newDate, ""); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, _ := repo.GetByID(ctx, m.ID)
	if got.Title != "Изменён" {
		t.Errorf("want title 'Изменён', got %q", got.Title)
	}
}

func TestMeetingRepo_Update_NotFound(t *testing.T) {
	pool := testutil.NewDB(t)
	testutil.TruncateTables(t, pool)
	repo := New(pool)
	ctx := testutil.Ctx()

	err := repo.Update(ctx, "00000000-0000-0000-0000-000000000000", "X", time.Now(), "")
	if err != errs.ErrNotFound {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestMeetingRepo_Delete(t *testing.T) {
	pool := testutil.NewDB(t)
	testutil.TruncateTables(t, pool)
	repo := New(pool)
	ctx := testutil.Ctx()

	m := seedMeeting(t, repo, "Удалить")

	if err := repo.Delete(ctx, m.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.GetByID(ctx, m.ID); err != errs.ErrNotFound {
		t.Errorf("after delete: want ErrNotFound, got %v", err)
	}
}

func TestMeetingRepo_GetAll_Pagination(t *testing.T) {
	pool := testutil.NewDB(t)
	testutil.TruncateTables(t, pool)
	repo := New(pool)
	ctx := testutil.Ctx()

	for i := 0; i < 3; i++ {
		seedMeeting(t, repo, "m")
	}

	meetings, total, err := repo.GetAll(ctx, 2, 0, "")
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if total != 3 {
		t.Errorf("want total=3, got %d", total)
	}
	if len(meetings) != 2 {
		t.Errorf("want 2 items (limit), got %d", len(meetings))
	}

	page2, _, _ := repo.GetAll(ctx, 2, 2, "")
	if len(page2) != 1 {
		t.Errorf("want 1 item on page 2, got %d", len(page2))
	}
}

func TestMeetingRepo_SetChairperson_And_Load(t *testing.T) {
	pool := testutil.NewDB(t)
	testutil.TruncateTables(t, pool)
	mRepo := New(pool)
	pRepo := personRepo.New(pool)
	ctx := testutil.Ctx()

	alice := seedPerson(t, pRepo, "Иванова", "Алиса")
	m := seedMeeting(t, mRepo, "Совещание")

	// Add alice to meeting first
	if err := mRepo.AddPerson(ctx, m.ID, alice.ID); err != nil {
		t.Fatalf("AddPerson: %v", err)
	}

	if err := mRepo.SetChairperson(ctx, m.ID, alice.ID); err != nil {
		t.Fatalf("SetChairperson: %v", err)
	}

	got, _ := mRepo.GetByID(ctx, m.ID)
	if got.Chairperson == nil || got.Chairperson.ID != alice.ID {
		t.Errorf("expected chairperson alice, got %+v", got.Chairperson)
	}
}

func TestMeetingRepo_AddAndRemovePerson(t *testing.T) {
	pool := testutil.NewDB(t)
	testutil.TruncateTables(t, pool)
	mRepo := New(pool)
	pRepo := personRepo.New(pool)
	ctx := testutil.Ctx()

	alice := seedPerson(t, pRepo, "Алиса", "А")
	bob := seedPerson(t, pRepo, "Боб", "Б")
	m := seedMeeting(t, mRepo, "Meeting")

	if err := mRepo.AddPerson(ctx, m.ID, alice.ID); err != nil {
		t.Fatalf("AddPerson alice: %v", err)
	}
	if err := mRepo.AddPerson(ctx, m.ID, bob.ID); err != nil {
		t.Fatalf("AddPerson bob: %v", err)
	}

	got, _ := mRepo.GetByID(ctx, m.ID)
	if len(got.People) != 2 {
		t.Fatalf("want 2 people, got %d", len(got.People))
	}

	if err := mRepo.RemovePerson(ctx, m.ID, alice.ID); err != nil {
		t.Fatalf("RemovePerson: %v", err)
	}

	got, _ = mRepo.GetByID(ctx, m.ID)
	if len(got.People) != 1 || got.People[0].ID != bob.ID {
		t.Errorf("unexpected people after remove: %v", got.People)
	}
}

func TestMeetingRepo_ReorderPeople(t *testing.T) {
	pool := testutil.NewDB(t)
	testutil.TruncateTables(t, pool)
	mRepo := New(pool)
	pRepo := personRepo.New(pool)
	ctx := testutil.Ctx()

	alice := seedPerson(t, pRepo, "Алиса", "А")
	bob := seedPerson(t, pRepo, "Боб", "Б")
	m := seedMeeting(t, mRepo, "Meeting")
	_ = mRepo.AddPerson(ctx, m.ID, alice.ID)
	_ = mRepo.AddPerson(ctx, m.ID, bob.ID)

	// Reverse order
	if err := mRepo.ReorderPeople(ctx, m.ID, []int{bob.ID, alice.ID}); err != nil {
		t.Fatalf("ReorderPeople: %v", err)
	}

	got, _ := mRepo.GetByID(ctx, m.ID)
	if got.People[0].ID != bob.ID || got.People[1].ID != alice.ID {
		t.Errorf("unexpected order after reorder: %v", got.People)
	}
}

func TestMeetingRepo_AddAgendaItem_And_Speakers(t *testing.T) {
	pool := testutil.NewDB(t)
	testutil.TruncateTables(t, pool)
	mRepo := New(pool)
	pRepo := personRepo.New(pool)
	ctx := testutil.Ctx()

	alice := seedPerson(t, pRepo, "Алиса", "А")
	bob := seedPerson(t, pRepo, "Боб", "Б")
	m := seedMeeting(t, mRepo, "Meeting")
	_ = mRepo.AddPerson(ctx, m.ID, alice.ID)
	_ = mRepo.AddPerson(ctx, m.ID, bob.ID)

	itemID, err := mRepo.AddAgendaItem(ctx, m.ID, "Первый вопрос", []int{alice.ID})
	if err != nil {
		t.Fatalf("AddAgendaItem: %v", err)
	}
	if itemID == 0 {
		t.Fatal("expected non-zero item ID")
	}

	got, _ := mRepo.GetByID(ctx, m.ID)
	if len(got.AgendaItems) != 1 {
		t.Fatalf("want 1 agenda item, got %d", len(got.AgendaItems))
	}
	if got.AgendaItems[0].Text != "Первый вопрос" {
		t.Errorf("unexpected item text: %q", got.AgendaItems[0].Text)
	}
	if len(got.AgendaItems[0].Speakers) != 1 || got.AgendaItems[0].Speakers[0].ID != alice.ID {
		t.Errorf("unexpected speakers: %v", got.AgendaItems[0].Speakers)
	}

	// Add second speaker
	if err := mRepo.AddAgendaItemSpeaker(ctx, m.ID, itemID, bob.ID); err != nil {
		t.Fatalf("AddAgendaItemSpeaker: %v", err)
	}
	got, _ = mRepo.GetByID(ctx, m.ID)
	if len(got.AgendaItems[0].Speakers) != 2 {
		t.Errorf("want 2 speakers, got %d", len(got.AgendaItems[0].Speakers))
	}
}

func TestMeetingRepo_UpdateAgendaItem(t *testing.T) {
	pool := testutil.NewDB(t)
	testutil.TruncateTables(t, pool)
	mRepo := New(pool)
	pRepo := personRepo.New(pool)
	ctx := testutil.Ctx()

	alice := seedPerson(t, pRepo, "Алиса", "А")
	bob := seedPerson(t, pRepo, "Боб", "Б")
	m := seedMeeting(t, mRepo, "Meeting")
	_ = mRepo.AddPerson(ctx, m.ID, alice.ID)
	_ = mRepo.AddPerson(ctx, m.ID, bob.ID)

	itemID, _ := mRepo.AddAgendaItem(ctx, m.ID, "Старый текст", []int{alice.ID})

	if err := mRepo.UpdateAgendaItem(ctx, m.ID, itemID, "Новый текст", []int{bob.ID}); err != nil {
		t.Fatalf("UpdateAgendaItem: %v", err)
	}

	got, _ := mRepo.GetByID(ctx, m.ID)
	item := got.AgendaItems[0]
	if item.Text != "Новый текст" {
		t.Errorf("want 'Новый текст', got %q", item.Text)
	}
	if len(item.Speakers) != 1 || item.Speakers[0].ID != bob.ID {
		t.Errorf("unexpected speakers after update: %v", item.Speakers)
	}
}

func TestMeetingRepo_DeleteAgendaItem(t *testing.T) {
	pool := testutil.NewDB(t)
	testutil.TruncateTables(t, pool)
	mRepo := New(pool)
	pRepo := personRepo.New(pool)
	ctx := testutil.Ctx()

	alice := seedPerson(t, pRepo, "Алиса", "А")
	m := seedMeeting(t, mRepo, "Meeting")
	_ = mRepo.AddPerson(ctx, m.ID, alice.ID)
	itemID, _ := mRepo.AddAgendaItem(ctx, m.ID, "Удалить", []int{alice.ID})

	if err := mRepo.DeleteAgendaItem(ctx, m.ID, itemID); err != nil {
		t.Fatalf("DeleteAgendaItem: %v", err)
	}

	got, _ := mRepo.GetByID(ctx, m.ID)
	if len(got.AgendaItems) != 0 {
		t.Errorf("want 0 agenda items after delete, got %d", len(got.AgendaItems))
	}
}

func TestMeetingRepo_ReorderAgendaItems(t *testing.T) {
	pool := testutil.NewDB(t)
	testutil.TruncateTables(t, pool)
	mRepo := New(pool)
	pRepo := personRepo.New(pool)
	ctx := testutil.Ctx()

	alice := seedPerson(t, pRepo, "Алиса", "А")
	m := seedMeeting(t, mRepo, "Meeting")
	_ = mRepo.AddPerson(ctx, m.ID, alice.ID)
	id1, _ := mRepo.AddAgendaItem(ctx, m.ID, "Первый", []int{alice.ID})
	id2, _ := mRepo.AddAgendaItem(ctx, m.ID, "Второй", []int{alice.ID})

	if err := mRepo.ReorderAgendaItems(ctx, m.ID, []int{id2, id1}); err != nil {
		t.Fatalf("ReorderAgendaItems: %v", err)
	}

	got, _ := mRepo.GetByID(ctx, m.ID)
	if got.AgendaItems[0].ID != id2 || got.AgendaItems[1].ID != id1 {
		t.Errorf("unexpected order: %v", got.AgendaItems)
	}
}

func TestMeetingRepo_RemoveAgendaItemSpeaker(t *testing.T) {
	pool := testutil.NewDB(t)
	testutil.TruncateTables(t, pool)
	mRepo := New(pool)
	pRepo := personRepo.New(pool)
	ctx := testutil.Ctx()

	alice := seedPerson(t, pRepo, "Алиса", "А")
	bob := seedPerson(t, pRepo, "Боб", "Б")
	m := seedMeeting(t, mRepo, "Meeting")
	_ = mRepo.AddPerson(ctx, m.ID, alice.ID)
	_ = mRepo.AddPerson(ctx, m.ID, bob.ID)
	itemID, _ := mRepo.AddAgendaItem(ctx, m.ID, "item", []int{alice.ID, bob.ID})

	if err := mRepo.RemoveAgendaItemSpeaker(ctx, m.ID, itemID, alice.ID); err != nil {
		t.Fatalf("RemoveAgendaItemSpeaker: %v", err)
	}

	got, _ := mRepo.GetByID(ctx, m.ID)
	spk := got.AgendaItems[0].Speakers
	if len(spk) != 1 || spk[0].ID != bob.ID {
		t.Errorf("unexpected speakers after remove: %v", spk)
	}
}

func TestMeetingRepo_ReorderAgendaItemSpeakers(t *testing.T) {
	pool := testutil.NewDB(t)
	testutil.TruncateTables(t, pool)
	mRepo := New(pool)
	pRepo := personRepo.New(pool)
	ctx := testutil.Ctx()

	alice := seedPerson(t, pRepo, "Алиса", "А")
	bob := seedPerson(t, pRepo, "Боб", "Б")
	m := seedMeeting(t, mRepo, "Meeting")
	_ = mRepo.AddPerson(ctx, m.ID, alice.ID)
	_ = mRepo.AddPerson(ctx, m.ID, bob.ID)
	itemID, _ := mRepo.AddAgendaItem(ctx, m.ID, "item", []int{alice.ID, bob.ID})

	// Reverse speaker order
	if err := mRepo.ReorderAgendaItemSpeakers(ctx, m.ID, itemID, []int{bob.ID, alice.ID}); err != nil {
		t.Fatalf("ReorderAgendaItemSpeakers: %v", err)
	}

	got, _ := mRepo.GetByID(ctx, m.ID)
	spk := got.AgendaItems[0].Speakers
	if spk[0].ID != bob.ID || spk[1].ID != alice.ID {
		t.Errorf("unexpected speaker order: %v", spk)
	}
}

func TestMeetingRepo_GetByID_FullNestedStructure(t *testing.T) {
	pool := testutil.NewDB(t)
	testutil.TruncateTables(t, pool)
	mRepo := New(pool)
	pRepo := personRepo.New(pool)
	ctx := testutil.Ctx()

	alice := seedPerson(t, pRepo, "Алиса", "А")
	bob := seedPerson(t, pRepo, "Боб", "Б")
	carol := seedPerson(t, pRepo, "Кэрол", "К")

	m := seedMeeting(t, mRepo, "Полное совещание")
	_ = mRepo.AddPerson(ctx, m.ID, alice.ID)
	_ = mRepo.AddPerson(ctx, m.ID, bob.ID)
	_ = mRepo.AddPerson(ctx, m.ID, carol.ID)
	_ = mRepo.SetChairperson(ctx, m.ID, alice.ID)
	id1, _ := mRepo.AddAgendaItem(ctx, m.ID, "Вопрос 1", []int{bob.ID})
	id2, _ := mRepo.AddAgendaItem(ctx, m.ID, "Вопрос 2", []int{bob.ID, carol.ID})

	got, err := mRepo.GetByID(ctx, m.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	// Chairperson
	if got.Chairperson == nil || got.Chairperson.ID != alice.ID {
		t.Errorf("wrong chairperson: %+v", got.Chairperson)
	}
	// People
	if len(got.People) != 3 {
		t.Errorf("want 3 people, got %d", len(got.People))
	}
	// Agenda items
	if len(got.AgendaItems) != 2 {
		t.Fatalf("want 2 agenda items, got %d", len(got.AgendaItems))
	}
	// Find items by ID to avoid ordering assumption
	itemMap := make(map[int]meeting.AgendaItem)
	for _, item := range got.AgendaItems {
		itemMap[item.ID] = item
	}
	if len(itemMap[id1].Speakers) != 1 {
		t.Errorf("item1: want 1 speaker, got %d", len(itemMap[id1].Speakers))
	}
	if len(itemMap[id2].Speakers) != 2 {
		t.Errorf("item2: want 2 speakers, got %d", len(itemMap[id2].Speakers))
	}
}
