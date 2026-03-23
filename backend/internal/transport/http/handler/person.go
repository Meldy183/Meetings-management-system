package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"meetings-editor/internal/domain/person"
	svcPerson "meetings-editor/internal/service/person"
	"meetings-editor/internal/transport/http/model"
	"meetings-editor/pkg/errs"
)

type PersonHandler struct {
	svc svcPerson.Service
}

func NewPersonHandler(svc svcPerson.Service) *PersonHandler {
	return &PersonHandler{svc: svc}
}

// GET /people?q=...
func (h *PersonHandler) List(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))

	var people []person.Person
	var err error
	if q == "" {
		people, err = h.svc.GetAll(r.Context())
	} else {
		people, err = h.svc.Search(r.Context(), q)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal error", nil)
		return
	}

	resp := make([]model.PersonResponse, 0, len(people))
	for i := range people {
		resp = append(resp, toPersonResponse(&people[i]))
	}
	respond(w, http.StatusOK, resp)
}

// GET /people/{id}
func (h *PersonHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid person id", nil)
		return
	}

	p, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			respondError(w, http.StatusNotFound, "person not found", nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error", nil)
		return
	}

	respond(w, http.StatusOK, toPersonResponse(p))
}

// POST /people
func (h *PersonHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.PersonCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	if req.LastName == "" || req.FirstName == "" {
		respondError(w, http.StatusBadRequest, "last_name and first_name are required", nil)
		return
	}

	p := &person.Person{
		LastName:   req.LastName,
		FirstName:  req.FirstName,
		MiddleName: req.MiddleName,
		Info:       req.Info,
	}

	created, err := h.svc.Create(r.Context(), p)
	if err != nil {
		if errors.Is(err, errs.ErrConflict) {
			respondError(w, http.StatusConflict, "person already exists", nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error", nil)
		return
	}

	respond(w, http.StatusCreated, toPersonResponse(created))
}

// PATCH /people/{id}
func (h *PersonHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid person id", nil)
		return
	}

	var req model.PersonCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	if req.LastName == "" || req.FirstName == "" {
		respondError(w, http.StatusBadRequest, "last_name and first_name are required", nil)
		return
	}

	p := &person.Person{
		ID:         id,
		LastName:   req.LastName,
		FirstName:  req.FirstName,
		MiddleName: req.MiddleName,
		Info:       req.Info,
	}

	updated, err := h.svc.Update(r.Context(), p)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			respondError(w, http.StatusNotFound, "person not found", nil)
			return
		}
		if errors.Is(err, errs.ErrConflict) {
			respondError(w, http.StatusConflict, "person with this name already exists", nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error", nil)
		return
	}

	respond(w, http.StatusOK, toPersonResponse(updated))
}

// POST /people/sort
func (h *PersonHandler) Sort(w http.ResponseWriter, r *http.Request) {
	var req model.SortPeopleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	sorted, err := h.svc.SortByIDs(r.Context(), req.IDs)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal error", nil)
		return
	}

	respond(w, http.StatusOK, model.SortPeopleResponse{IDs: sorted})
}

// DELETE /people/{id}
func (h *PersonHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid person id", nil)
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			respondError(w, http.StatusNotFound, "person not found", nil)
			return
		}
		if errors.Is(err, errs.ErrConflict) {
			respondError(w, http.StatusConflict, "person is referenced in existing meetings", nil)
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func toPersonResponse(p *person.Person) model.PersonResponse {
	return model.PersonResponse{
		ID:         p.ID,
		LastName:   p.LastName,
		FirstName:  p.FirstName,
		MiddleName: p.MiddleName,
		Info:       p.Info,
	}
}
