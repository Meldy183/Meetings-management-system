package meeting_test

import (
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	domMeeting "meetings-editor/internal/domain/meeting"
	"meetings-editor/internal/domain/person"
	"meetings-editor/internal/mocks"
	svcMeeting "meetings-editor/internal/service/meeting"
	"meetings-editor/internal/testutil"
	"meetings-editor/pkg/errs"
)

const testMeetingID = "00000000-0000-0000-0000-000000000001"

var (
	alice = person.Person{ID: 1, LastName: "Иванова", FirstName: "Алиса"}
	bob   = person.Person{ID: 2, LastName: "Петров", FirstName: "Борис"}
	carol = person.Person{ID: 3, LastName: "Сидорова", FirstName: "Карина"}
)

func setup(t *testing.T) (*mocks.MockMeetingRepository, *mocks.MockPersonRepository, svcMeeting.Service) {
	t.Helper()
	ctrl := gomock.NewController(t)
	meetingRepo := mocks.NewMockMeetingRepository(ctrl)
	personRepo := mocks.NewMockPersonRepository(ctrl)
	svc := svcMeeting.New(meetingRepo, personRepo)
	return meetingRepo, personRepo, svc
}

func meetingWith(people []person.Person, chairperson *person.Person, items []domMeeting.AgendaItem) *domMeeting.Meeting {
	return &domMeeting.Meeting{
		ID:          testMeetingID,
		Title:       "Test Meeting",
		Date:        time.Now(),
		Chairperson: chairperson,
		People:      people,
		AgendaItems: items,
	}
}

// --- Create ---

func TestCreate_OK(t *testing.T) {
	repo, _, svc := setup(t)
	ctx := testutil.Ctx()

	req := &svcMeeting.CreateRequest{Title: "Test", Date: time.Now()}
	want := &domMeeting.Meeting{ID: testMeetingID, Title: "Test"}
	repo.EXPECT().Create(ctx, gomock.Any()).Return(want, nil)

	got, err := svc.Create(ctx, req)
	if err != nil || got.ID != testMeetingID {
		t.Errorf("unexpected: %v, %v", got, err)
	}
}

// --- SetChairperson ---

func TestSetChairperson_OK(t *testing.T) {
	repo, _, svc := setup(t)
	ctx := testutil.Ctx()

	m := meetingWith([]person.Person{alice, bob}, nil, nil)
	updated := meetingWith([]person.Person{alice, bob}, &alice, nil)

	repo.EXPECT().GetByID(ctx, testMeetingID).Return(m, nil)
	repo.EXPECT().SetChairperson(ctx, testMeetingID, alice.ID).Return(nil)
	repo.EXPECT().GetByID(ctx, testMeetingID).Return(updated, nil)

	got, err := svc.SetChairperson(ctx, testMeetingID, alice.ID)
	if err != nil || got.Chairperson == nil || got.Chairperson.ID != alice.ID {
		t.Errorf("unexpected: %v, %v", got, err)
	}
}

func TestSetChairperson_PersonNotInMeeting(t *testing.T) {
	repo, _, svc := setup(t)
	ctx := testutil.Ctx()

	m := meetingWith([]person.Person{alice}, nil, nil)
	repo.EXPECT().GetByID(ctx, testMeetingID).Return(m, nil)

	_, err := svc.SetChairperson(ctx, testMeetingID, bob.ID)
	var e *svcMeeting.ErrChairpersonNotInMeeting
	if !errors.As(err, &e) {
		t.Errorf("want ErrChairpersonNotInMeeting, got %v", err)
	}
}

func TestSetChairperson_MeetingNotFound(t *testing.T) {
	repo, _, svc := setup(t)
	ctx := testutil.Ctx()

	repo.EXPECT().GetByID(ctx, testMeetingID).Return(nil, errs.ErrNotFound)

	_, err := svc.SetChairperson(ctx, testMeetingID, alice.ID)
	if !errors.Is(err, errs.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

// --- AddPerson ---

func TestAddPerson_OK(t *testing.T) {
	repo, personRepo, svc := setup(t)
	ctx := testutil.Ctx()

	m := meetingWith([]person.Person{alice}, nil, nil)
	updated := meetingWith([]person.Person{alice, bob}, nil, nil)

	personRepo.EXPECT().GetByIDs(ctx, []int{bob.ID}).Return([]person.Person{bob}, nil)
	repo.EXPECT().GetByID(ctx, testMeetingID).Return(m, nil)
	repo.EXPECT().AddPerson(ctx, testMeetingID, bob.ID).Return(nil)
	repo.EXPECT().GetByID(ctx, testMeetingID).Return(updated, nil)

	got, err := svc.AddPerson(ctx, testMeetingID, bob.ID)
	if err != nil || len(got.People) != 2 {
		t.Errorf("unexpected: %v, %v", got, err)
	}
}

func TestAddPerson_PersonNotExists(t *testing.T) {
	repo, personRepo, svc := setup(t)
	ctx := testutil.Ctx()
	_ = repo

	personRepo.EXPECT().GetByIDs(ctx, []int{bob.ID}).Return([]person.Person{}, nil)

	_, err := svc.AddPerson(ctx, testMeetingID, bob.ID)
	var e *svcMeeting.ErrInvalidIDs
	if !errors.As(err, &e) {
		t.Errorf("want ErrInvalidIDs, got %v", err)
	}
}

func TestAddPerson_AlreadyInMeeting(t *testing.T) {
	repo, personRepo, svc := setup(t)
	ctx := testutil.Ctx()

	m := meetingWith([]person.Person{alice, bob}, nil, nil)
	personRepo.EXPECT().GetByIDs(ctx, []int{bob.ID}).Return([]person.Person{bob}, nil)
	repo.EXPECT().GetByID(ctx, testMeetingID).Return(m, nil)

	_, err := svc.AddPerson(ctx, testMeetingID, bob.ID)
	var e *svcMeeting.ErrPersonAlreadyInMeeting
	if !errors.As(err, &e) {
		t.Errorf("want ErrPersonAlreadyInMeeting, got %v", err)
	}
}

// --- RemovePerson ---

func TestRemovePerson_OK(t *testing.T) {
	repo, _, svc := setup(t)
	ctx := testutil.Ctx()

	m := meetingWith([]person.Person{alice, bob}, &alice, nil)
	updated := meetingWith([]person.Person{alice}, &alice, nil)

	repo.EXPECT().GetByID(ctx, testMeetingID).Return(m, nil)
	repo.EXPECT().RemovePerson(ctx, testMeetingID, bob.ID).Return(nil)
	repo.EXPECT().GetByID(ctx, testMeetingID).Return(updated, nil)

	got, err := svc.RemovePerson(ctx, testMeetingID, bob.ID)
	if err != nil || len(got.People) != 1 {
		t.Errorf("unexpected: %v, %v", got, err)
	}
}

func TestRemovePerson_IsChairperson(t *testing.T) {
	repo, _, svc := setup(t)
	ctx := testutil.Ctx()

	m := meetingWith([]person.Person{alice, bob}, &alice, nil)
	repo.EXPECT().GetByID(ctx, testMeetingID).Return(m, nil)

	_, err := svc.RemovePerson(ctx, testMeetingID, alice.ID)
	var e *svcMeeting.ErrChairpersonRemoval
	if !errors.As(err, &e) {
		t.Errorf("want ErrChairpersonRemoval, got %v", err)
	}
}

func TestRemovePerson_IsSpeaker(t *testing.T) {
	repo, _, svc := setup(t)
	ctx := testutil.Ctx()

	items := []domMeeting.AgendaItem{{ID: 1, Text: "item", Speakers: []person.Person{bob}}}
	m := meetingWith([]person.Person{alice, bob}, &alice, items)
	repo.EXPECT().GetByID(ctx, testMeetingID).Return(m, nil)

	_, err := svc.RemovePerson(ctx, testMeetingID, bob.ID)
	var e *svcMeeting.ErrSpeakerRemoval
	if !errors.As(err, &e) {
		t.Errorf("want ErrSpeakerRemoval, got %v", err)
	}
}

// --- AddAgendaItem ---

func TestAddAgendaItem_OK(t *testing.T) {
	repo, _, svc := setup(t)
	ctx := testutil.Ctx()

	m := meetingWith([]person.Person{alice, bob}, &alice, nil)
	updated := meetingWith([]person.Person{alice, bob}, &alice, []domMeeting.AgendaItem{
		{ID: 1, Text: "new item", Speakers: []person.Person{bob}},
	})

	repo.EXPECT().GetByID(ctx, testMeetingID).Return(m, nil)
	repo.EXPECT().AddAgendaItem(ctx, testMeetingID, "new item", []int{bob.ID}).Return(1, nil)
	repo.EXPECT().GetByID(ctx, testMeetingID).Return(updated, nil)

	got, err := svc.AddAgendaItem(ctx, testMeetingID, "new item", []int{bob.ID})
	if err != nil || len(got.AgendaItems) != 1 {
		t.Errorf("unexpected: %v, %v", got, err)
	}
}

func TestAddAgendaItem_EmptySpeakerIDs(t *testing.T) {
	_, _, svc := setup(t)
	ctx := testutil.Ctx()

	_, err := svc.AddAgendaItem(ctx, testMeetingID, "text", []int{})
	var e *svcMeeting.ErrLastSpeaker
	if !errors.As(err, &e) {
		t.Errorf("want ErrLastSpeaker, got %v", err)
	}
}

func TestAddAgendaItem_SpeakerNotInMeeting(t *testing.T) {
	repo, _, svc := setup(t)
	ctx := testutil.Ctx()

	m := meetingWith([]person.Person{alice}, &alice, nil)
	repo.EXPECT().GetByID(ctx, testMeetingID).Return(m, nil)

	_, err := svc.AddAgendaItem(ctx, testMeetingID, "text", []int{carol.ID})
	var e *svcMeeting.ErrSpeakerNotInMeeting
	if !errors.As(err, &e) {
		t.Errorf("want ErrSpeakerNotInMeeting, got %v", err)
	}
}

// --- UpdateAgendaItem ---

func TestUpdateAgendaItem_OK(t *testing.T) {
	repo, _, svc := setup(t)
	ctx := testutil.Ctx()

	items := []domMeeting.AgendaItem{{ID: 1, Text: "old", Speakers: []person.Person{alice}}}
	m := meetingWith([]person.Person{alice, bob}, &alice, items)
	updated := meetingWith([]person.Person{alice, bob}, &alice, []domMeeting.AgendaItem{
		{ID: 1, Text: "new", Speakers: []person.Person{bob}},
	})

	repo.EXPECT().GetByID(ctx, testMeetingID).Return(m, nil)
	repo.EXPECT().UpdateAgendaItem(ctx, testMeetingID, 1, "new", []int{bob.ID}).Return(nil)
	repo.EXPECT().GetByID(ctx, testMeetingID).Return(updated, nil)

	got, err := svc.UpdateAgendaItem(ctx, testMeetingID, 1, "new", []int{bob.ID})
	if err != nil || got.AgendaItems[0].Text != "new" {
		t.Errorf("unexpected: %v, %v", got, err)
	}
}

func TestUpdateAgendaItem_ItemNotFound(t *testing.T) {
	repo, _, svc := setup(t)
	ctx := testutil.Ctx()

	m := meetingWith([]person.Person{alice}, &alice, []domMeeting.AgendaItem{{ID: 1, Text: "x", Speakers: []person.Person{alice}}})
	repo.EXPECT().GetByID(ctx, testMeetingID).Return(m, nil)

	_, err := svc.UpdateAgendaItem(ctx, testMeetingID, 999, "y", []int{alice.ID})
	if !errors.Is(err, errs.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestUpdateAgendaItem_EmptySpeakerIDs(t *testing.T) {
	_, _, svc := setup(t)
	ctx := testutil.Ctx()

	_, err := svc.UpdateAgendaItem(ctx, testMeetingID, 1, "text", []int{})
	var e *svcMeeting.ErrLastSpeaker
	if !errors.As(err, &e) {
		t.Errorf("want ErrLastSpeaker, got %v", err)
	}
}

func TestUpdateAgendaItem_SpeakerNotInMeeting(t *testing.T) {
	repo, _, svc := setup(t)
	ctx := testutil.Ctx()

	items := []domMeeting.AgendaItem{{ID: 1, Text: "x", Speakers: []person.Person{alice}}}
	m := meetingWith([]person.Person{alice}, &alice, items)
	repo.EXPECT().GetByID(ctx, testMeetingID).Return(m, nil)

	_, err := svc.UpdateAgendaItem(ctx, testMeetingID, 1, "text", []int{carol.ID})
	var e *svcMeeting.ErrSpeakerNotInMeeting
	if !errors.As(err, &e) {
		t.Errorf("want ErrSpeakerNotInMeeting, got %v", err)
	}
}

// --- DeleteAgendaItem ---

func TestDeleteAgendaItem_OK(t *testing.T) {
	repo, _, svc := setup(t)
	ctx := testutil.Ctx()

	updated := meetingWith([]person.Person{alice}, &alice, nil)
	repo.EXPECT().DeleteAgendaItem(ctx, testMeetingID, 1).Return(nil)
	repo.EXPECT().GetByID(ctx, testMeetingID).Return(updated, nil)

	got, err := svc.DeleteAgendaItem(ctx, testMeetingID, 1)
	if err != nil || len(got.AgendaItems) != 0 {
		t.Errorf("unexpected: %v, %v", got, err)
	}
}

func TestDeleteAgendaItem_NotFound(t *testing.T) {
	repo, _, svc := setup(t)
	ctx := testutil.Ctx()

	repo.EXPECT().DeleteAgendaItem(ctx, testMeetingID, 99).Return(errs.ErrNotFound)

	_, err := svc.DeleteAgendaItem(ctx, testMeetingID, 99)
	if !errors.Is(err, errs.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

// --- AddAgendaItemSpeaker ---

func TestAddAgendaItemSpeaker_OK(t *testing.T) {
	repo, _, svc := setup(t)
	ctx := testutil.Ctx()

	items := []domMeeting.AgendaItem{{ID: 1, Text: "x", Speakers: []person.Person{alice}}}
	m := meetingWith([]person.Person{alice, bob}, &alice, items)
	updated := meetingWith([]person.Person{alice, bob}, &alice, []domMeeting.AgendaItem{
		{ID: 1, Text: "x", Speakers: []person.Person{alice, bob}},
	})

	repo.EXPECT().GetByID(ctx, testMeetingID).Return(m, nil)
	repo.EXPECT().AddAgendaItemSpeaker(ctx, testMeetingID, 1, bob.ID).Return(nil)
	repo.EXPECT().GetByID(ctx, testMeetingID).Return(updated, nil)

	got, err := svc.AddAgendaItemSpeaker(ctx, testMeetingID, 1, bob.ID)
	if err != nil || len(got.AgendaItems[0].Speakers) != 2 {
		t.Errorf("unexpected: %v, %v", got, err)
	}
}

func TestAddAgendaItemSpeaker_SpeakerNotInMeeting(t *testing.T) {
	repo, _, svc := setup(t)
	ctx := testutil.Ctx()

	items := []domMeeting.AgendaItem{{ID: 1, Text: "x", Speakers: []person.Person{alice}}}
	m := meetingWith([]person.Person{alice}, &alice, items)
	repo.EXPECT().GetByID(ctx, testMeetingID).Return(m, nil)

	_, err := svc.AddAgendaItemSpeaker(ctx, testMeetingID, 1, carol.ID)
	var e *svcMeeting.ErrSpeakerNotInMeeting
	if !errors.As(err, &e) {
		t.Errorf("want ErrSpeakerNotInMeeting, got %v", err)
	}
}

func TestAddAgendaItemSpeaker_AlreadySpeaker(t *testing.T) {
	repo, _, svc := setup(t)
	ctx := testutil.Ctx()

	items := []domMeeting.AgendaItem{{ID: 1, Text: "x", Speakers: []person.Person{alice}}}
	m := meetingWith([]person.Person{alice}, &alice, items)
	repo.EXPECT().GetByID(ctx, testMeetingID).Return(m, nil)

	_, err := svc.AddAgendaItemSpeaker(ctx, testMeetingID, 1, alice.ID)
	var e *svcMeeting.ErrSpeakerAlreadyOnItem
	if !errors.As(err, &e) {
		t.Errorf("want ErrSpeakerAlreadyOnItem, got %v", err)
	}
}

// --- RemoveAgendaItemSpeaker ---

func TestRemoveAgendaItemSpeaker_OK(t *testing.T) {
	repo, _, svc := setup(t)
	ctx := testutil.Ctx()

	items := []domMeeting.AgendaItem{{ID: 1, Text: "x", Speakers: []person.Person{alice, bob}}}
	m := meetingWith([]person.Person{alice, bob}, &alice, items)
	updated := meetingWith([]person.Person{alice, bob}, &alice, []domMeeting.AgendaItem{
		{ID: 1, Text: "x", Speakers: []person.Person{alice}},
	})

	repo.EXPECT().GetByID(ctx, testMeetingID).Return(m, nil)
	repo.EXPECT().RemoveAgendaItemSpeaker(ctx, testMeetingID, 1, bob.ID).Return(nil)
	repo.EXPECT().GetByID(ctx, testMeetingID).Return(updated, nil)

	got, err := svc.RemoveAgendaItemSpeaker(ctx, testMeetingID, 1, bob.ID)
	if err != nil || len(got.AgendaItems[0].Speakers) != 1 {
		t.Errorf("unexpected: %v, %v", got, err)
	}
}

func TestRemoveAgendaItemSpeaker_LastSpeaker(t *testing.T) {
	repo, _, svc := setup(t)
	ctx := testutil.Ctx()

	items := []domMeeting.AgendaItem{{ID: 1, Text: "x", Speakers: []person.Person{alice}}}
	m := meetingWith([]person.Person{alice}, &alice, items)
	repo.EXPECT().GetByID(ctx, testMeetingID).Return(m, nil)

	_, err := svc.RemoveAgendaItemSpeaker(ctx, testMeetingID, 1, alice.ID)
	var e *svcMeeting.ErrLastSpeaker
	if !errors.As(err, &e) {
		t.Errorf("want ErrLastSpeaker, got %v", err)
	}
}

// --- ReorderPeople ---

func TestReorderPeople_OK(t *testing.T) {
	repo, _, svc := setup(t)
	ctx := testutil.Ctx()

	m := meetingWith([]person.Person{alice, bob}, &alice, nil)
	repo.EXPECT().GetByID(ctx, testMeetingID).Return(m, nil)
	repo.EXPECT().ReorderPeople(ctx, testMeetingID, []int{bob.ID, alice.ID}).Return(nil)

	if err := svc.ReorderPeople(ctx, testMeetingID, []int{bob.ID, alice.ID}); err != nil {
		t.Fatal(err)
	}
}

func TestReorderPeople_WrongLength(t *testing.T) {
	repo, _, svc := setup(t)
	ctx := testutil.Ctx()

	m := meetingWith([]person.Person{alice, bob}, nil, nil)
	repo.EXPECT().GetByID(ctx, testMeetingID).Return(m, nil)

	err := svc.ReorderPeople(ctx, testMeetingID, []int{alice.ID})
	var e *svcMeeting.ErrPersonSetMismatch
	if !errors.As(err, &e) {
		t.Errorf("want ErrPersonSetMismatch, got %v", err)
	}
}

func TestReorderPeople_UnknownID(t *testing.T) {
	repo, _, svc := setup(t)
	ctx := testutil.Ctx()

	m := meetingWith([]person.Person{alice, bob}, nil, nil)
	repo.EXPECT().GetByID(ctx, testMeetingID).Return(m, nil)

	err := svc.ReorderPeople(ctx, testMeetingID, []int{alice.ID, carol.ID})
	var e *svcMeeting.ErrPersonSetMismatch
	if !errors.As(err, &e) {
		t.Errorf("want ErrPersonSetMismatch, got %v", err)
	}
}

// --- ReorderAgendaItems ---

func TestReorderAgendaItems_OK(t *testing.T) {
	repo, _, svc := setup(t)
	ctx := testutil.Ctx()

	items := []domMeeting.AgendaItem{
		{ID: 1, Text: "a", Speakers: []person.Person{alice}},
		{ID: 2, Text: "b", Speakers: []person.Person{bob}},
	}
	m := meetingWith([]person.Person{alice, bob}, &alice, items)
	repo.EXPECT().GetByID(ctx, testMeetingID).Return(m, nil)
	repo.EXPECT().ReorderAgendaItems(ctx, testMeetingID, []int{2, 1}).Return(nil)

	if err := svc.ReorderAgendaItems(ctx, testMeetingID, []int{2, 1}); err != nil {
		t.Fatal(err)
	}
}

func TestReorderAgendaItems_Mismatch(t *testing.T) {
	repo, _, svc := setup(t)
	ctx := testutil.Ctx()

	items := []domMeeting.AgendaItem{{ID: 1, Text: "a", Speakers: []person.Person{alice}}}
	m := meetingWith([]person.Person{alice}, &alice, items)
	repo.EXPECT().GetByID(ctx, testMeetingID).Return(m, nil)

	err := svc.ReorderAgendaItems(ctx, testMeetingID, []int{99})
	var e *svcMeeting.ErrAgendaItemSetMismatch
	if !errors.As(err, &e) {
		t.Errorf("want ErrAgendaItemSetMismatch, got %v", err)
	}
}

// --- ReorderAgendaItemSpeakers ---

func TestReorderAgendaItemSpeakers_OK(t *testing.T) {
	repo, _, svc := setup(t)
	ctx := testutil.Ctx()

	items := []domMeeting.AgendaItem{{ID: 1, Text: "x", Speakers: []person.Person{alice, bob}}}
	m := meetingWith([]person.Person{alice, bob}, &alice, items)
	repo.EXPECT().GetByID(ctx, testMeetingID).Return(m, nil)
	repo.EXPECT().ReorderAgendaItemSpeakers(ctx, testMeetingID, 1, []int{bob.ID, alice.ID}).Return(nil)

	if err := svc.ReorderAgendaItemSpeakers(ctx, testMeetingID, 1, []int{bob.ID, alice.ID}); err != nil {
		t.Fatal(err)
	}
}

func TestReorderAgendaItemSpeakers_Mismatch(t *testing.T) {
	repo, _, svc := setup(t)
	ctx := testutil.Ctx()

	items := []domMeeting.AgendaItem{{ID: 1, Text: "x", Speakers: []person.Person{alice, bob}}}
	m := meetingWith([]person.Person{alice, bob}, &alice, items)
	repo.EXPECT().GetByID(ctx, testMeetingID).Return(m, nil)

	err := svc.ReorderAgendaItemSpeakers(ctx, testMeetingID, 1, []int{alice.ID, carol.ID})
	var e *svcMeeting.ErrAgendaItemSetMismatch
	if !errors.As(err, &e) {
		t.Errorf("want ErrAgendaItemSetMismatch, got %v", err)
	}
}
