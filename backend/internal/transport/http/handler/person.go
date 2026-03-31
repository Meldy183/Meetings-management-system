package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/extrame/xls"
	"github.com/xuri/excelize/v2"

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

// GET /people?q=...&order=alpha|id
func (h *PersonHandler) List(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	order := r.URL.Query().Get("order")
	if order != "id" {
		order = "alpha"
	}

	var people []person.Person
	var err error
	if q == "" {
		people, err = h.svc.GetAll(r.Context(), order)
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

// POST /people/import
func (h *PersonHandler) Import(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "failed to parse multipart form", nil)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "file field required", nil)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to read file", nil)
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))

	var rows [][]string
	switch ext {
	case ".xlsx":
		rows, err = parseXLSX(data)
	case ".xls":
		rows, err = parseXLS(data)
	default:
		respondError(w, http.StatusBadRequest, "unsupported file format, use .xlsx or .xls", nil)
		return
	}
	if err != nil {
		respondError(w, http.StatusBadRequest, "failed to parse file", nil)
		return
	}

	imported := 0
	for _, row := range rows {
		if len(row) < 2 {
			continue
		}
		lastName := strings.TrimSpace(row[0])
		firstName := strings.TrimSpace(row[1])
		if lastName == "" || firstName == "" {
			continue
		}
		middleName := ""
		if len(row) > 2 {
			middleName = strings.TrimSpace(row[2])
		}
		info := ""
		if len(row) > 3 {
			info = strings.TrimSpace(row[3])
		}
		p := &person.Person{
			LastName:   lastName,
			FirstName:  firstName,
			MiddleName: middleName,
			Info:       info,
		}
		if _, err := h.svc.Create(r.Context(), p); err != nil {
			continue
		}
		imported++
	}

	respond(w, http.StatusOK, model.ImportPeopleResponse{Imported: imported})
}

func parseXLSX(data []byte) ([][]string, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, nil
	}
	allRows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, err
	}
	if len(allRows) <= 1 {
		return nil, nil
	}
	return allRows[1:], nil // skip header row
}

func parseXLS(data []byte) ([][]string, error) {
	tmp, err := os.CreateTemp("", "import-*.xls")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		return nil, err
	}
	tmp.Close()

	wb, err := xls.Open(tmp.Name(), "utf-8")
	if err != nil {
		return nil, err
	}

	sheet := wb.GetSheet(0)
	if sheet == nil {
		return nil, nil
	}

	var rows [][]string
	for i := 1; i <= int(sheet.MaxRow); i++ { // skip row 0 (header)
		row := sheet.Row(i)
		cols := make([]string, 4)
		for j := 0; j < 4 && j < int(row.LastCol()); j++ {
			cols[j] = row.Col(j)
		}
		rows = append(rows, cols)
	}
	return rows, nil
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
