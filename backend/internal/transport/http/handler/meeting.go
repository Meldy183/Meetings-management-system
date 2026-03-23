package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	domMeeting "meetings-editor/internal/domain/meeting"
	"meetings-editor/internal/domain/person"
	svcMeeting "meetings-editor/internal/service/meeting"
	"meetings-editor/internal/transport/http/model"
	"meetings-editor/pkg/errs"
)

type MeetingHandler struct {
	svc    svcMeeting.Service
	export ExportService
}

// ExportService is implemented by the docx package.
type ExportService interface {
	Agenda(m *domMeeting.Meeting) ([]byte, error)
	Participants(m *domMeeting.Meeting) ([]byte, error)
}

func NewMeetingHandler(svc svcMeeting.Service, export ExportService) *MeetingHandler {
	return &MeetingHandler{svc: svc, export: export}
}

// GET /meetings?limit=20&offset=0
func (h *MeetingHandler) List(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 20)
	offset := queryInt(r, "offset", 0)

	if limit < 1 {
		limit = 1
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	meetings, total, err := h.svc.GetAll(r.Context(), limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal error", nil)
		return
	}

	items := make([]model.MeetingSummaryResponse, 0, len(meetings))
	for i := range meetings {
		items = append(items, toMeetingSummaryResponse(&meetings[i]))
	}

	respond(w, http.StatusOK, model.MeetingListResponse{
		Total:  total,
		Limit:  limit,
		Offset: offset,
		Items:  items,
	})
}

// POST /meetings
func (h *MeetingHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.MeetingCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	if req.Title == "" || req.Date.IsZero() {
		respondError(w, http.StatusBadRequest, "title and date are required", nil)
		return
	}

	svcReq := &svcMeeting.CreateRequest{
		Title: req.Title,
		Date:  req.Date,
		Place: req.Place,
	}

	m, err := h.svc.Create(r.Context(), svcReq)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal error", nil)
		return
	}

	respond(w, http.StatusCreated, toMeetingResponse(m))
}

// GET /meetings/{id}
func (h *MeetingHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing meeting id", nil)
		return
	}

	m, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			respondError(w, http.StatusNotFound, "meeting not found", nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error", nil)
		return
	}

	respond(w, http.StatusOK, toMeetingResponse(m))
}

// PATCH /meetings/{id}
func (h *MeetingHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing meeting id", nil)
		return
	}

	var req model.MeetingUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	if req.Title == "" || req.Date.IsZero() {
		respondError(w, http.StatusBadRequest, "title and date are required", nil)
		return
	}

	m, err := h.svc.Update(r.Context(), id, &svcMeeting.UpdateRequest{
		Title: req.Title,
		Date:  req.Date,
		Place: req.Place,
	})
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			respondError(w, http.StatusNotFound, "meeting not found", nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error", nil)
		return
	}
	respond(w, http.StatusOK, toMeetingResponse(m))
}

// DELETE /meetings/{id}
func (h *MeetingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing meeting id", nil)
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			respondError(w, http.StatusNotFound, "meeting not found", nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error", nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PUT /meetings/{id}/chairperson
func (h *MeetingHandler) SetChairperson(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing meeting id", nil)
		return
	}

	var req model.SetChairpersonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PersonID == 0 {
		respondError(w, http.StatusBadRequest, "person_id is required", nil)
		return
	}

	m, err := h.svc.SetChairperson(r.Context(), id, req.PersonID)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			respondError(w, http.StatusNotFound, "meeting not found", nil)
			return
		}
		var e *svcMeeting.ErrChairpersonNotInMeeting
		if errors.As(err, &e) {
			respondError(w, http.StatusUnprocessableEntity, e.Error(), nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error", nil)
		return
	}
	respond(w, http.StatusOK, toMeetingResponse(m))
}

// POST /meetings/{id}/people
func (h *MeetingHandler) AddPerson(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing meeting id", nil)
		return
	}

	var req model.AddMeetingPersonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PersonID == 0 {
		respondError(w, http.StatusBadRequest, "person_id is required", nil)
		return
	}

	m, err := h.svc.AddPerson(r.Context(), id, req.PersonID)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			respondError(w, http.StatusNotFound, "meeting not found", nil)
			return
		}
		var e1 *svcMeeting.ErrInvalidIDs
		if errors.As(err, &e1) {
			respondError(w, http.StatusUnprocessableEntity, "person not found", nil)
			return
		}
		var e2 *svcMeeting.ErrPersonAlreadyInMeeting
		if errors.As(err, &e2) {
			respondError(w, http.StatusConflict, e2.Error(), nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error", nil)
		return
	}
	respond(w, http.StatusOK, toMeetingResponse(m))
}

// DELETE /meetings/{id}/people/{pid}
func (h *MeetingHandler) RemovePerson(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	pidStr := r.PathValue("pid")
	pid, err := strconv.Atoi(pidStr)
	if id == "" || err != nil {
		respondError(w, http.StatusBadRequest, "invalid path parameters", nil)
		return
	}

	m, err := h.svc.RemovePerson(r.Context(), id, pid)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			respondError(w, http.StatusNotFound, "meeting or person not found", nil)
			return
		}
		var e1 *svcMeeting.ErrChairpersonRemoval
		if errors.As(err, &e1) {
			respondError(w, http.StatusConflict, e1.Error(), nil)
			return
		}
		var e2 *svcMeeting.ErrSpeakerRemoval
		if errors.As(err, &e2) {
			respondError(w, http.StatusConflict, e2.Error(), nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error", nil)
		return
	}
	respond(w, http.StatusOK, toMeetingResponse(m))
}

// POST /meetings/{id}/agenda-items
func (h *MeetingHandler) AddAgendaItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing meeting id", nil)
		return
	}

	var req model.AgendaItemUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	if req.Text == "" || len(req.SpeakerIDs) == 0 {
		respondError(w, http.StatusBadRequest, "text and speaker_ids are required", nil)
		return
	}

	m, err := h.svc.AddAgendaItem(r.Context(), id, req.Text, req.SpeakerIDs)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			respondError(w, http.StatusNotFound, "meeting not found", nil)
			return
		}
		var e *svcMeeting.ErrSpeakerNotInMeeting
		if errors.As(err, &e) {
			respondError(w, http.StatusUnprocessableEntity, e.Error(), nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error", nil)
		return
	}
	respond(w, http.StatusCreated, toMeetingResponse(m))
}

// PUT /meetings/{id}/agenda-items/{item_id}
func (h *MeetingHandler) UpdateAgendaItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	itemIDStr := r.PathValue("item_id")
	itemID, err := strconv.Atoi(itemIDStr)
	if id == "" || err != nil {
		respondError(w, http.StatusBadRequest, "invalid path parameters", nil)
		return
	}

	var req model.AgendaItemUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	if req.Text == "" || len(req.SpeakerIDs) == 0 {
		respondError(w, http.StatusBadRequest, "text and speaker_ids are required", nil)
		return
	}

	m, err := h.svc.UpdateAgendaItem(r.Context(), id, itemID, req.Text, req.SpeakerIDs)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			respondError(w, http.StatusNotFound, "meeting or agenda item not found", nil)
			return
		}
		var e *svcMeeting.ErrSpeakerNotInMeeting
		if errors.As(err, &e) {
			respondError(w, http.StatusUnprocessableEntity, e.Error(), nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error", nil)
		return
	}
	respond(w, http.StatusOK, toMeetingResponse(m))
}

// DELETE /meetings/{id}/agenda-items/{item_id}
func (h *MeetingHandler) DeleteAgendaItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	itemIDStr := r.PathValue("item_id")
	itemID, err := strconv.Atoi(itemIDStr)
	if id == "" || err != nil {
		respondError(w, http.StatusBadRequest, "invalid path parameters", nil)
		return
	}

	m, err := h.svc.DeleteAgendaItem(r.Context(), id, itemID)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			respondError(w, http.StatusNotFound, "meeting or agenda item not found", nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error", nil)
		return
	}
	respond(w, http.StatusOK, toMeetingResponse(m))
}

// POST /meetings/{id}/agenda-items/{item_id}/speakers
func (h *MeetingHandler) AddAgendaItemSpeaker(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	itemIDStr := r.PathValue("item_id")
	itemID, err := strconv.Atoi(itemIDStr)
	if id == "" || err != nil {
		respondError(w, http.StatusBadRequest, "invalid path parameters", nil)
		return
	}

	var req model.AddAgendaItemSpeakerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PersonID == 0 {
		respondError(w, http.StatusBadRequest, "person_id is required", nil)
		return
	}

	m, err := h.svc.AddAgendaItemSpeaker(r.Context(), id, itemID, req.PersonID)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			respondError(w, http.StatusNotFound, "meeting or agenda item not found", nil)
			return
		}
		var e1 *svcMeeting.ErrSpeakerNotInMeeting
		if errors.As(err, &e1) {
			respondError(w, http.StatusUnprocessableEntity, e1.Error(), nil)
			return
		}
		var e2 *svcMeeting.ErrSpeakerAlreadyOnItem
		if errors.As(err, &e2) {
			respondError(w, http.StatusConflict, e2.Error(), nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error", nil)
		return
	}
	respond(w, http.StatusOK, toMeetingResponse(m))
}

// DELETE /meetings/{id}/agenda-items/{item_id}/speakers/{pid}
func (h *MeetingHandler) RemoveAgendaItemSpeaker(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	itemIDStr := r.PathValue("item_id")
	pidStr := r.PathValue("pid")
	itemID, err1 := strconv.Atoi(itemIDStr)
	pid, err2 := strconv.Atoi(pidStr)
	if id == "" || err1 != nil || err2 != nil {
		respondError(w, http.StatusBadRequest, "invalid path parameters", nil)
		return
	}

	m, err := h.svc.RemoveAgendaItemSpeaker(r.Context(), id, itemID, pid)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			respondError(w, http.StatusNotFound, "meeting, agenda item, or speaker not found", nil)
			return
		}
		var e *svcMeeting.ErrLastSpeaker
		if errors.As(err, &e) {
			respondError(w, http.StatusConflict, e.Error(), nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error", nil)
		return
	}
	respond(w, http.StatusOK, toMeetingResponse(m))
}

// PUT /meetings/{id}/agenda-items/{item_id}/speakers/order
func (h *MeetingHandler) ReorderAgendaItemSpeakers(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	itemIDStr := r.PathValue("item_id")
	itemID, err := strconv.Atoi(itemIDStr)
	if id == "" || err != nil {
		respondError(w, http.StatusBadRequest, "invalid path parameters", nil)
		return
	}

	var req model.ReorderAgendaItemSpeakersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	if len(req.PersonIDs) == 0 {
		respondError(w, http.StatusBadRequest, "person_ids must not be empty", nil)
		return
	}

	err = h.svc.ReorderAgendaItemSpeakers(r.Context(), id, itemID, req.PersonIDs)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			respondError(w, http.StatusNotFound, "meeting or agenda item not found", nil)
			return
		}
		var mismatch *svcMeeting.ErrAgendaItemSetMismatch
		if errors.As(err, &mismatch) {
			respondError(w, http.StatusUnprocessableEntity, mismatch.Error(), nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error", nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PUT /meetings/{id}/people/order
func (h *MeetingHandler) ReorderPeople(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing meeting id", nil)
		return
	}

	var req model.ReorderPeopleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	if len(req.PersonIDs) == 0 {
		respondError(w, http.StatusBadRequest, "person_ids must not be empty", nil)
		return
	}

	err := h.svc.ReorderPeople(r.Context(), id, req.PersonIDs)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			respondError(w, http.StatusNotFound, "meeting not found", nil)
			return
		}
		var mismatch *svcMeeting.ErrPersonSetMismatch
		if errors.As(err, &mismatch) {
			respondError(w, http.StatusUnprocessableEntity, mismatch.Error(), nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PUT /meetings/{id}/agenda-items/order
func (h *MeetingHandler) ReorderAgendaItems(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing meeting id", nil)
		return
	}

	var req model.ReorderAgendaItemsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	if len(req.AgendaItemIDs) == 0 {
		respondError(w, http.StatusBadRequest, "agenda_item_ids must not be empty", nil)
		return
	}

	err := h.svc.ReorderAgendaItems(r.Context(), id, req.AgendaItemIDs)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			respondError(w, http.StatusNotFound, "meeting not found", nil)
			return
		}
		var mismatch *svcMeeting.ErrAgendaItemSetMismatch
		if errors.As(err, &mismatch) {
			respondError(w, http.StatusUnprocessableEntity, mismatch.Error(), nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GET /meetings/{id}/meta
func (h *MeetingHandler) GetMeta(w http.ResponseWriter, r *http.Request) {
	m, ok := h.fetchMeeting(w, r)
	if !ok {
		return
	}
	respond(w, http.StatusOK, toMeetingSummaryResponse(m))
}

// GET /meetings/{id}/people
func (h *MeetingHandler) GetPeople(w http.ResponseWriter, r *http.Request) {
	m, ok := h.fetchMeeting(w, r)
	if !ok {
		return
	}
	people := make([]model.PersonResponse, 0, len(m.People))
	for _, p := range m.People {
		people = append(people, toPersonResp(p))
	}
	respond(w, http.StatusOK, people)
}

// GET /meetings/{id}/agenda-items
func (h *MeetingHandler) GetAgendaItems(w http.ResponseWriter, r *http.Request) {
	m, ok := h.fetchMeeting(w, r)
	if !ok {
		return
	}
	items := make([]model.AgendaItemResponse, 0, len(m.AgendaItems))
	for _, item := range m.AgendaItems {
		speakers := make([]model.PersonResponse, 0, len(item.Speakers))
		for _, spk := range item.Speakers {
			speakers = append(speakers, toPersonResp(spk))
		}
		items = append(items, model.AgendaItemResponse{
			ID:       item.ID,
			Text:     item.Text,
			Speakers: speakers,
		})
	}
	respond(w, http.StatusOK, items)
}

// GET /meetings/{id}/export/agenda
func (h *MeetingHandler) ExportAgenda(w http.ResponseWriter, r *http.Request) {
	m, ok := h.fetchMeeting(w, r)
	if !ok {
		return
	}

	if m.Status() == "incomplete" {
		respondError(w, http.StatusConflict, (&svcMeeting.ErrMeetingIncomplete{}).Error(), nil)
		return
	}

	data, err := h.export.Agenda(m)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate document", nil)
		return
	}

	sendDocx(w, data, "agenda-"+m.ID[:8]+".docx")
}

// GET /meetings/{id}/export/participants
func (h *MeetingHandler) ExportParticipants(w http.ResponseWriter, r *http.Request) {
	m, ok := h.fetchMeeting(w, r)
	if !ok {
		return
	}

	if m.Status() == "incomplete" {
		respondError(w, http.StatusConflict, (&svcMeeting.ErrMeetingIncomplete{}).Error(), nil)
		return
	}

	data, err := h.export.Participants(m)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate document", nil)
		return
	}

	sendDocx(w, data, "participants-"+m.ID[:8]+".docx")
}

// --- helpers ---

func (h *MeetingHandler) fetchMeeting(w http.ResponseWriter, r *http.Request) (*domMeeting.Meeting, bool) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing meeting id", nil)
		return nil, false
	}
	m, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			respondError(w, http.StatusNotFound, "meeting not found", nil)
		} else {
			respondError(w, http.StatusInternalServerError, "internal error", nil)
		}
		return nil, false
	}
	return m, true
}

func sendDocx(w http.ResponseWriter, data []byte, filename string) {
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func queryInt(r *http.Request, key string, def int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

func toPersonResp(p person.Person) model.PersonResponse {
	return model.PersonResponse{
		ID:         p.ID,
		LastName:   p.LastName,
		FirstName:  p.FirstName,
		MiddleName: p.MiddleName,
		Info:       p.Info,
	}
}

func toMeetingSummaryResponse(m *domMeeting.Meeting) model.MeetingSummaryResponse {
	var chairperson *model.PersonResponse
	if m.Chairperson != nil {
		r := toPersonResp(*m.Chairperson)
		chairperson = &r
	}
	return model.MeetingSummaryResponse{
		ID:          m.ID,
		Title:       m.Title,
		Date:        m.Date,
		Place:       m.Place,
		Chairperson: chairperson,
		Status:      m.Status(),
		CreatedAt:   m.CreatedAt,
	}
}

func toMeetingResponse(m *domMeeting.Meeting) model.MeetingResponse {
	var chairperson *model.PersonResponse
	if m.Chairperson != nil {
		r := toPersonResp(*m.Chairperson)
		chairperson = &r
	}

	items := make([]model.AgendaItemResponse, 0, len(m.AgendaItems))
	for _, item := range m.AgendaItems {
		speakers := make([]model.PersonResponse, 0, len(item.Speakers))
		for _, spk := range item.Speakers {
			speakers = append(speakers, toPersonResp(spk))
		}
		items = append(items, model.AgendaItemResponse{
			ID:       item.ID,
			Text:     item.Text,
			Speakers: speakers,
		})
	}

	people := make([]model.PersonResponse, 0, len(m.People))
	for _, p := range m.People {
		people = append(people, toPersonResp(p))
	}

	return model.MeetingResponse{
		ID:          m.ID,
		Title:       m.Title,
		Date:        m.Date,
		Place:       m.Place,
		Chairperson: chairperson,
		AgendaItems: items,
		People:      people,
		Status:      m.Status(),
		CreatedAt:   m.CreatedAt,
	}
}
