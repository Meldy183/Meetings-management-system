package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	domMeeting "meetings-editor/internal/domain/meeting"
	"meetings-editor/internal/domain/person"
	"meetings-editor/internal/mocks"
	svcMeeting "meetings-editor/internal/service/meeting"
	"meetings-editor/internal/transport/http/model"
	"meetings-editor/pkg/errs"
)

const testID = "00000000-0000-0000-0000-000000000001"

var (
	alice = person.Person{ID: 1, LastName: "Иванова", FirstName: "Алиса"}
	bob   = person.Person{ID: 2, LastName: "Петров", FirstName: "Борис"}
)

func newMeetingHandler(t *testing.T) (*mocks.MockMeetingService, *mocks.MockExportService, *MeetingHandler) {
	t.Helper()
	ctrl := gomock.NewController(t)
	svc := mocks.NewMockMeetingService(ctrl)
	export := mocks.NewMockExportService(ctrl)
	return svc, export, NewMeetingHandler(svc, export)
}

func completeMeetingDomain() *domMeeting.Meeting {
	return &domMeeting.Meeting{
		ID:          testID,
		Title:       "Test",
		Date:        time.Now(),
		Chairperson: &alice,
		People:      []person.Person{alice, bob},
		AgendaItems: []domMeeting.AgendaItem{
			{ID: 1, Text: "item", Speakers: []person.Person{bob}},
		},
	}
}

// --- List ---

func TestMeetingList_OK(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	svc.EXPECT().GetAll(gomock.Any(), 20, 0).Return(
		[]domMeeting.Meeting{*completeMeetingDomain()}, 1, nil,
	)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /meetings", h.List)
	w := doRequest(mux, "GET", "/meetings", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp model.MeetingListResponse
	decodeJSON(t, w, &resp)
	if resp.Total != 1 || len(resp.Items) != 1 {
		t.Errorf("unexpected list response: %+v", resp)
	}
}

// --- Create ---

func TestMeetingCreate_OK(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	want := &domMeeting.Meeting{ID: testID, Title: "New Meeting", Date: time.Now()}
	svc.EXPECT().Create(gomock.Any(), gomock.Any()).Return(want, nil)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /meetings", h.Create)
	w := doRequest(mux, "POST", "/meetings", map[string]any{
		"title": "New Meeting",
		"date":  time.Now(),
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d: %s", w.Code, w.Body.String())
	}
	var resp model.MeetingResponse
	decodeJSON(t, w, &resp)
	if resp.ID != testID {
		t.Errorf("want id=%s, got %s", testID, resp.ID)
	}
}

func TestMeetingCreate_MissingTitle(t *testing.T) {
	_, export, h := newMeetingHandler(t)
	_ = export

	mux := http.NewServeMux()
	mux.HandleFunc("POST /meetings", h.Create)
	w := doRequest(mux, "POST", "/meetings", map[string]any{
		"date": time.Now(),
	})

	if w.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", w.Code)
	}
}

// --- GetByID ---

func TestMeetingGetByID_OK(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	svc.EXPECT().GetByID(gomock.Any(), testID).Return(completeMeetingDomain(), nil)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /meetings/{id}", h.GetByID)
	w := doRequest(mux, "GET", "/meetings/"+testID, nil)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	var resp model.MeetingResponse
	decodeJSON(t, w, &resp)
	if resp.ID != testID {
		t.Errorf("want id=%s, got %s", testID, resp.ID)
	}
}

func TestMeetingGetByID_NotFound(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	svc.EXPECT().GetByID(gomock.Any(), testID).Return(nil, errs.ErrNotFound)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /meetings/{id}", h.GetByID)
	w := doRequest(mux, "GET", "/meetings/"+testID, nil)

	if w.Code != http.StatusNotFound {
		t.Errorf("want 404, got %d", w.Code)
	}
}

// --- Update ---

func TestMeetingUpdate_OK(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	svc.EXPECT().Update(gomock.Any(), testID, gomock.Any()).Return(completeMeetingDomain(), nil)

	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /meetings/{id}", h.Update)
	w := doRequest(mux, "PATCH", "/meetings/"+testID, map[string]any{
		"title": "Updated", "date": time.Now(),
	})

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMeetingUpdate_NotFound(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	svc.EXPECT().Update(gomock.Any(), testID, gomock.Any()).Return(nil, errs.ErrNotFound)

	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /meetings/{id}", h.Update)
	w := doRequest(mux, "PATCH", "/meetings/"+testID, map[string]any{
		"title": "X", "date": time.Now(),
	})

	if w.Code != http.StatusNotFound {
		t.Errorf("want 404, got %d", w.Code)
	}
}

// --- Delete ---

func TestMeetingDelete_OK(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	svc.EXPECT().Delete(gomock.Any(), testID).Return(nil)

	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /meetings/{id}", h.Delete)
	w := doRequest(mux, "DELETE", "/meetings/"+testID, nil)

	if w.Code != http.StatusNoContent {
		t.Errorf("want 204, got %d", w.Code)
	}
}

// --- SetChairperson ---

func TestSetChairperson_OK(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	svc.EXPECT().SetChairperson(gomock.Any(), testID, alice.ID).Return(completeMeetingDomain(), nil)

	mux := http.NewServeMux()
	mux.HandleFunc("PUT /meetings/{id}/chairperson", h.SetChairperson)
	w := doRequest(mux, "PUT", "/meetings/"+testID+"/chairperson", map[string]int{
		"person_id": alice.ID,
	})

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSetChairperson_NotInMeeting(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	svc.EXPECT().SetChairperson(gomock.Any(), testID, alice.ID).
		Return(nil, &svcMeeting.ErrChairpersonNotInMeeting{})

	mux := http.NewServeMux()
	mux.HandleFunc("PUT /meetings/{id}/chairperson", h.SetChairperson)
	w := doRequest(mux, "PUT", "/meetings/"+testID+"/chairperson", map[string]int{
		"person_id": alice.ID,
	})

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("want 422, got %d", w.Code)
	}
}

func TestSetChairperson_ZeroPersonID(t *testing.T) {
	_, export, h := newMeetingHandler(t)
	_ = export

	r := httptest.NewRequest("PUT", "/meetings/"+testID+"/chairperson", nil)
	r.SetPathValue("id", testID)
	w := httptest.NewRecorder()
	h.SetChairperson(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", w.Code)
	}
}

// --- AddPerson ---

func TestAddPerson_OK(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	svc.EXPECT().AddPerson(gomock.Any(), testID, bob.ID).Return(completeMeetingDomain(), nil)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /meetings/{id}/people", h.AddPerson)
	w := doRequest(mux, "POST", "/meetings/"+testID+"/people", map[string]int{
		"person_id": bob.ID,
	})

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAddPerson_AlreadyInMeeting(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	svc.EXPECT().AddPerson(gomock.Any(), testID, bob.ID).
		Return(nil, &svcMeeting.ErrPersonAlreadyInMeeting{})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /meetings/{id}/people", h.AddPerson)
	w := doRequest(mux, "POST", "/meetings/"+testID+"/people", map[string]int{
		"person_id": bob.ID,
	})

	if w.Code != http.StatusConflict {
		t.Errorf("want 409, got %d", w.Code)
	}
}

func TestAddPerson_PersonNotExists(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	svc.EXPECT().AddPerson(gomock.Any(), testID, bob.ID).
		Return(nil, &svcMeeting.ErrInvalidIDs{IDs: []int{bob.ID}})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /meetings/{id}/people", h.AddPerson)
	w := doRequest(mux, "POST", "/meetings/"+testID+"/people", map[string]int{
		"person_id": bob.ID,
	})

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("want 422, got %d", w.Code)
	}
}

// --- RemovePerson ---

func TestRemovePerson_OK(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	updated := completeMeetingDomain()
	updated.People = []person.Person{alice}
	svc.EXPECT().RemovePerson(gomock.Any(), testID, bob.ID).Return(updated, nil)

	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /meetings/{id}/people/{pid}", h.RemovePerson)
	w := doRequest(mux, "DELETE", "/meetings/"+testID+"/people/2", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRemovePerson_IsChairperson(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	svc.EXPECT().RemovePerson(gomock.Any(), testID, alice.ID).
		Return(nil, &svcMeeting.ErrChairpersonRemoval{})

	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /meetings/{id}/people/{pid}", h.RemovePerson)
	w := doRequest(mux, "DELETE", "/meetings/"+testID+"/people/1", nil)

	if w.Code != http.StatusConflict {
		t.Errorf("want 409, got %d", w.Code)
	}
}

func TestRemovePerson_IsSpeaker(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	svc.EXPECT().RemovePerson(gomock.Any(), testID, bob.ID).
		Return(nil, &svcMeeting.ErrSpeakerRemoval{})

	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /meetings/{id}/people/{pid}", h.RemovePerson)
	w := doRequest(mux, "DELETE", "/meetings/"+testID+"/people/2", nil)

	if w.Code != http.StatusConflict {
		t.Errorf("want 409, got %d", w.Code)
	}
}

// --- AddAgendaItem ---

func TestAddAgendaItem_OK(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	svc.EXPECT().AddAgendaItem(gomock.Any(), testID, "new item", []int{bob.ID}).
		Return(completeMeetingDomain(), nil)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /meetings/{id}/agenda-items", h.AddAgendaItem)
	w := doRequest(mux, "POST", "/meetings/"+testID+"/agenda-items", map[string]any{
		"text": "new item", "speaker_ids": []int{bob.ID},
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAddAgendaItem_MissingFields(t *testing.T) {
	_, export, h := newMeetingHandler(t)
	_ = export

	mux := http.NewServeMux()
	mux.HandleFunc("POST /meetings/{id}/agenda-items", h.AddAgendaItem)
	// text present but no speaker_ids
	w := doRequest(mux, "POST", "/meetings/"+testID+"/agenda-items", map[string]any{
		"text": "item",
	})

	if w.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", w.Code)
	}
}

func TestAddAgendaItem_SpeakerNotInMeeting(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	svc.EXPECT().AddAgendaItem(gomock.Any(), testID, "item", gomock.Any()).
		Return(nil, &svcMeeting.ErrSpeakerNotInMeeting{})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /meetings/{id}/agenda-items", h.AddAgendaItem)
	w := doRequest(mux, "POST", "/meetings/"+testID+"/agenda-items", map[string]any{
		"text": "item", "speaker_ids": []int{99},
	})

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("want 422, got %d", w.Code)
	}
}

// --- UpdateAgendaItem ---

func TestUpdateAgendaItem_OK(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	svc.EXPECT().UpdateAgendaItem(gomock.Any(), testID, 1, "updated", []int{bob.ID}).
		Return(completeMeetingDomain(), nil)

	mux := http.NewServeMux()
	mux.HandleFunc("PUT /meetings/{id}/agenda-items/{item_id}", h.UpdateAgendaItem)
	w := doRequest(mux, "PUT", "/meetings/"+testID+"/agenda-items/1", map[string]any{
		"text": "updated", "speaker_ids": []int{bob.ID},
	})

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUpdateAgendaItem_NotFound(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	svc.EXPECT().UpdateAgendaItem(gomock.Any(), testID, 999, gomock.Any(), gomock.Any()).
		Return(nil, errs.ErrNotFound)

	mux := http.NewServeMux()
	mux.HandleFunc("PUT /meetings/{id}/agenda-items/{item_id}", h.UpdateAgendaItem)
	w := doRequest(mux, "PUT", "/meetings/"+testID+"/agenda-items/999", map[string]any{
		"text": "x", "speaker_ids": []int{1},
	})

	if w.Code != http.StatusNotFound {
		t.Errorf("want 404, got %d", w.Code)
	}
}

// --- DeleteAgendaItem ---

func TestDeleteAgendaItem_OK(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	svc.EXPECT().DeleteAgendaItem(gomock.Any(), testID, 1).Return(completeMeetingDomain(), nil)

	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /meetings/{id}/agenda-items/{item_id}", h.DeleteAgendaItem)
	w := doRequest(mux, "DELETE", "/meetings/"+testID+"/agenda-items/1", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}

// --- AddAgendaItemSpeaker ---

func TestAddAgendaItemSpeaker_OK(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	svc.EXPECT().AddAgendaItemSpeaker(gomock.Any(), testID, 1, alice.ID).
		Return(completeMeetingDomain(), nil)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /meetings/{id}/agenda-items/{item_id}/speakers", h.AddAgendaItemSpeaker)
	w := doRequest(mux, "POST", "/meetings/"+testID+"/agenda-items/1/speakers", map[string]int{
		"person_id": alice.ID,
	})

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAddAgendaItemSpeaker_AlreadySpeaker(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	svc.EXPECT().AddAgendaItemSpeaker(gomock.Any(), testID, 1, alice.ID).
		Return(nil, &svcMeeting.ErrSpeakerAlreadyOnItem{})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /meetings/{id}/agenda-items/{item_id}/speakers", h.AddAgendaItemSpeaker)
	w := doRequest(mux, "POST", "/meetings/"+testID+"/agenda-items/1/speakers", map[string]int{
		"person_id": alice.ID,
	})

	if w.Code != http.StatusConflict {
		t.Errorf("want 409, got %d", w.Code)
	}
}

func TestAddAgendaItemSpeaker_NotInMeeting(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	svc.EXPECT().AddAgendaItemSpeaker(gomock.Any(), testID, 1, 99).
		Return(nil, &svcMeeting.ErrSpeakerNotInMeeting{})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /meetings/{id}/agenda-items/{item_id}/speakers", h.AddAgendaItemSpeaker)
	w := doRequest(mux, "POST", "/meetings/"+testID+"/agenda-items/1/speakers", map[string]int{
		"person_id": 99,
	})

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("want 422, got %d", w.Code)
	}
}

// --- RemoveAgendaItemSpeaker ---

func TestRemoveAgendaItemSpeaker_OK(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	svc.EXPECT().RemoveAgendaItemSpeaker(gomock.Any(), testID, 1, bob.ID).
		Return(completeMeetingDomain(), nil)

	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /meetings/{id}/agenda-items/{item_id}/speakers/{pid}", h.RemoveAgendaItemSpeaker)
	w := doRequest(mux, "DELETE", "/meetings/"+testID+"/agenda-items/1/speakers/2", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}

func TestRemoveAgendaItemSpeaker_LastSpeaker(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	svc.EXPECT().RemoveAgendaItemSpeaker(gomock.Any(), testID, 1, bob.ID).
		Return(nil, &svcMeeting.ErrLastSpeaker{})

	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /meetings/{id}/agenda-items/{item_id}/speakers/{pid}", h.RemoveAgendaItemSpeaker)
	w := doRequest(mux, "DELETE", "/meetings/"+testID+"/agenda-items/1/speakers/2", nil)

	if w.Code != http.StatusConflict {
		t.Errorf("want 409, got %d", w.Code)
	}
}

// --- ReorderPeople ---

func TestReorderPeople_OK(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	svc.EXPECT().ReorderPeople(gomock.Any(), testID, []int{2, 1}).Return(nil)

	mux := http.NewServeMux()
	mux.HandleFunc("PUT /meetings/{id}/people/order", h.ReorderPeople)
	w := doRequest(mux, "PUT", "/meetings/"+testID+"/people/order", map[string]any{
		"person_ids": []int{2, 1},
	})

	if w.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestReorderPeople_SetMismatch(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	svc.EXPECT().ReorderPeople(gomock.Any(), testID, gomock.Any()).
		Return(&svcMeeting.ErrPersonSetMismatch{})

	mux := http.NewServeMux()
	mux.HandleFunc("PUT /meetings/{id}/people/order", h.ReorderPeople)
	w := doRequest(mux, "PUT", "/meetings/"+testID+"/people/order", map[string]any{
		"person_ids": []int{99},
	})

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("want 422, got %d", w.Code)
	}
}

// --- ExportAgenda ---

func TestExportAgenda_OK(t *testing.T) {
	svc, export, h := newMeetingHandler(t)

	m := completeMeetingDomain()
	svc.EXPECT().GetByID(gomock.Any(), testID).Return(m, nil)
	export.EXPECT().Agenda(m).Return([]byte("fakezip"), nil)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /meetings/{id}/export/agenda", h.ExportAgenda)
	w := doRequest(mux, "GET", "/meetings/"+testID+"/export/agenda", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
	ct := w.Header().Get("Content-Type")
	if ct != "application/vnd.openxmlformats-officedocument.wordprocessingml.document" {
		t.Errorf("unexpected Content-Type: %s", ct)
	}
}

func TestExportAgenda_IncompleteBlocked(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	// Incomplete: no chairperson
	m := completeMeetingDomain()
	m.Chairperson = nil
	svc.EXPECT().GetByID(gomock.Any(), testID).Return(m, nil)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /meetings/{id}/export/agenda", h.ExportAgenda)
	w := doRequest(mux, "GET", "/meetings/"+testID+"/export/agenda", nil)

	if w.Code != http.StatusConflict {
		t.Errorf("want 409, got %d", w.Code)
	}
}

// --- ExportParticipants ---

func TestExportParticipants_IncompleteBlocked(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	m := completeMeetingDomain()
	m.AgendaItems = nil // makes it incomplete
	svc.EXPECT().GetByID(gomock.Any(), testID).Return(m, nil)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /meetings/{id}/export/participants", h.ExportParticipants)
	w := doRequest(mux, "GET", "/meetings/"+testID+"/export/participants", nil)

	if w.Code != http.StatusConflict {
		t.Errorf("want 409, got %d", w.Code)
	}
}

// --- GetMeta / GetPeople / GetAgendaItems ---

func TestGetMeta_OK(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	svc.EXPECT().GetByID(gomock.Any(), testID).Return(completeMeetingDomain(), nil)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /meetings/{id}/meta", h.GetMeta)
	w := doRequest(mux, "GET", "/meetings/"+testID+"/meta", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	var resp model.MeetingSummaryResponse
	decodeJSON(t, w, &resp)
	if resp.ID != testID {
		t.Errorf("want id=%s, got %s", testID, resp.ID)
	}
}

func TestGetPeople_OK(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	svc.EXPECT().GetByID(gomock.Any(), testID).Return(completeMeetingDomain(), nil)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /meetings/{id}/people", h.GetPeople)
	w := doRequest(mux, "GET", "/meetings/"+testID+"/people", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	var resp []model.PersonResponse
	decodeJSON(t, w, &resp)
	if len(resp) != 2 {
		t.Errorf("want 2 people, got %d", len(resp))
	}
}

func TestGetAgendaItems_OK(t *testing.T) {
	svc, export, h := newMeetingHandler(t)
	_ = export

	svc.EXPECT().GetByID(gomock.Any(), testID).Return(completeMeetingDomain(), nil)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /meetings/{id}/agenda-items", h.GetAgendaItems)
	w := doRequest(mux, "GET", "/meetings/"+testID+"/agenda-items", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	var resp []model.AgendaItemResponse
	decodeJSON(t, w, &resp)
	if len(resp) != 1 || resp[0].ID != 1 {
		t.Errorf("unexpected agenda items: %v", resp)
	}
}
