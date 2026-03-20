package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"meetings-editor/internal/domain/participant"
	svcParticipant "meetings-editor/internal/service/participant"
	"meetings-editor/internal/transport/http/model"
	"meetings-editor/pkg/errs"
)


type ParticipantHandler struct {
	svc svcParticipant.Service
}

func NewParticipantHandler(svc svcParticipant.Service) *ParticipantHandler {
	return &ParticipantHandler{svc: svc}
}

// GET /participants?q=...
func (h *ParticipantHandler) List(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))

	var participants []participant.Participant
	var err error
	if q == "" {
		participants, err = h.svc.GetAll(r.Context())
	} else {
		participants, err = h.svc.Search(r.Context(), q)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal error", nil)
		return
	}

	resp := make([]model.ParticipantResponse, 0, len(participants))
	for i := range participants {
		resp = append(resp, toParticipantResponse(&participants[i]))
	}
	respond(w, http.StatusOK, resp)
}

// POST /participants
func (h *ParticipantHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.ParticipantCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	if req.LastName == "" || req.FirstName == "" {
		respondError(w, http.StatusBadRequest, "last_name and first_name are required", nil)
		return
	}

	p := &participant.Participant{
		LastName:   req.LastName,
		FirstName:  req.FirstName,
		MiddleName: req.MiddleName,
		Info:       req.Info,
	}

	created, err := h.svc.Create(r.Context(), p)
	if err != nil {
		if errors.Is(err, errs.ErrConflict) {
			respondError(w, http.StatusConflict, "participant already exists", nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error", nil)
		return
	}

	respond(w, http.StatusCreated, toParticipantResponse(created))
}

// PUT /participants/{id}
func (h *ParticipantHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid participant id", nil)
		return
	}

	var req model.ParticipantCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	if req.LastName == "" || req.FirstName == "" {
		respondError(w, http.StatusBadRequest, "last_name and first_name are required", nil)
		return
	}

	p := &participant.Participant{
		ID:         id,
		LastName:   req.LastName,
		FirstName:  req.FirstName,
		MiddleName: req.MiddleName,
		Info:       req.Info,
	}

	updated, err := h.svc.Update(r.Context(), p)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			respondError(w, http.StatusNotFound, "participant not found", nil)
			return
		}
		if errors.Is(err, errs.ErrConflict) {
			respondError(w, http.StatusConflict, "participant with this name already exists", nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error", nil)
		return
	}

	respond(w, http.StatusOK, toParticipantResponse(updated))
}

// DELETE /participants/{id}
func (h *ParticipantHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid participant id", nil)
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			respondError(w, http.StatusNotFound, "participant not found", nil)
			return
		}
		if errors.Is(err, errs.ErrConflict) {
			respondError(w, http.StatusConflict, "participant is referenced in existing meetings", nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func toParticipantResponse(p *participant.Participant) model.ParticipantResponse {
	return model.ParticipantResponse{
		ID:         p.ID,
		LastName:   p.LastName,
		FirstName:  p.FirstName,
		MiddleName: p.MiddleName,
		Info:       p.Info,
	}
}
