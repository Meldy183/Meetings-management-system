package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	domMeeting "meetings-editor/internal/domain/meeting"
	"meetings-editor/internal/domain/participant"
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

	if req.Title == "" || req.Date.IsZero() || req.ChairpersonID == 0 ||
		len(req.AgendaItems) == 0 || len(req.ParticipantIDs) == 0 {
		respondError(w, http.StatusBadRequest, "missing required fields", nil)
		return
	}

	svcReq := &svcMeeting.CreateRequest{
		Title:          req.Title,
		Date:           req.Date,
		ChairpersonID:  req.ChairpersonID,
		ParticipantIDs: req.ParticipantIDs,
	}
	for _, item := range req.AgendaItems {
		svcReq.AgendaItems = append(svcReq.AgendaItems, svcMeeting.AgendaItemRequest{
			Text:      item.Text,
			SpeakerID: item.SpeakerID,
		})
	}

	m, err := h.svc.Create(r.Context(), svcReq)
	if err != nil {
		var invalidIDs *svcMeeting.ErrInvalidIDs
		if errors.As(err, &invalidIDs) {
			respondError(w, http.StatusUnprocessableEntity, "one or more participants not found",
				map[string]any{"invalid_participant_ids": invalidIDs.IDs})
			return
		}
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

// PUT /meetings/{id}/participants/order
func (h *MeetingHandler) ReorderParticipants(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing meeting id", nil)
		return
	}

	var req model.ReorderParticipantsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}
	if len(req.ParticipantIDs) == 0 {
		respondError(w, http.StatusBadRequest, "participant_ids must not be empty", nil)
		return
	}

	err := h.svc.ReorderParticipants(r.Context(), id, req.ParticipantIDs)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			respondError(w, http.StatusNotFound, "meeting not found", nil)
			return
		}
		var mismatch *svcMeeting.ErrParticipantSetMismatch
		if errors.As(err, &mismatch) {
			respondError(w, http.StatusUnprocessableEntity, mismatch.Error(), nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GET /meetings/{id}/export/agenda
func (h *MeetingHandler) ExportAgenda(w http.ResponseWriter, r *http.Request) {
	m, ok := h.fetchMeeting(w, r)
	if !ok {
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

func toParticipantResp(p participant.Participant) model.ParticipantResponse {
	return model.ParticipantResponse{
		ID:         p.ID,
		LastName:   p.LastName,
		FirstName:  p.FirstName,
		MiddleName: p.MiddleName,
		Info:       p.Info,
	}
}

func toMeetingSummaryResponse(m *domMeeting.Meeting) model.MeetingSummaryResponse {
	return model.MeetingSummaryResponse{
		ID:          m.ID,
		Title:       m.Title,
		Date:        m.Date,
		Chairperson: toParticipantResp(m.Chairperson),
		CreatedAt:   m.CreatedAt,
	}
}

func toMeetingResponse(m *domMeeting.Meeting) model.MeetingResponse {
	items := make([]model.AgendaItemResponse, 0, len(m.AgendaItems))
	for _, item := range m.AgendaItems {
		items = append(items, model.AgendaItemResponse{
			Text:    item.Text,
			Speaker: toParticipantResp(item.Speaker),
		})
	}

	participants := make([]model.ParticipantResponse, 0, len(m.Participants))
	for _, p := range m.Participants {
		participants = append(participants, toParticipantResp(p))
	}

	return model.MeetingResponse{
		ID:           m.ID,
		Title:        m.Title,
		Date:         m.Date,
		Chairperson:  toParticipantResp(m.Chairperson),
		AgendaItems:  items,
		Participants: participants,
		CreatedAt:    m.CreatedAt,
	}
}
