package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/mock/gomock"

	"meetings-editor/internal/domain/person"
	"meetings-editor/internal/mocks"
	"meetings-editor/internal/testutil"
	"meetings-editor/internal/transport/http/model"
	"meetings-editor/pkg/errs"
)

func newPersonHandler(t *testing.T) (*mocks.MockPersonService, *PersonHandler) {
	t.Helper()
	ctrl := gomock.NewController(t)
	svc := mocks.NewMockPersonService(ctrl)
	return svc, NewPersonHandler(svc)
}

func doRequest(h http.Handler, method, target string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	r := httptest.NewRequest(method, target, &buf)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func decodeJSON(t *testing.T, w *httptest.ResponseRecorder, dst any) {
	t.Helper()
	if err := json.NewDecoder(w.Body).Decode(dst); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

// GET /people — no query

func TestPersonList_NoQuery_ReturnsAll(t *testing.T) {
	svc, h := newPersonHandler(t)
	ctx := testutil.Ctx()

	people := []person.Person{{ID: 1, LastName: "Иванов", FirstName: "Иван"}}
	svc.EXPECT().GetAll(gomock.Any()).DoAndReturn(func(c any) ([]person.Person, error) {
		_ = c
		return people, nil
	})

	mux := http.NewServeMux()
	mux.HandleFunc("GET /people", h.List)
	w := doRequest(mux, "GET", "/people", nil)
	_ = ctx

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	var resp []model.PersonResponse
	decodeJSON(t, w, &resp)
	if len(resp) != 1 || resp[0].ID != 1 {
		t.Errorf("unexpected response: %v", resp)
	}
}

func TestPersonList_WithQuery_CallsSearch(t *testing.T) {
	svc, h := newPersonHandler(t)

	svc.EXPECT().Search(gomock.Any(), "иванов").Return(nil, nil)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /people", h.List)
	w := doRequest(mux, "GET", "/people?q=иванов", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}

// GET /people/{id}

func TestPersonGetByID_Found(t *testing.T) {
	svc, h := newPersonHandler(t)

	svc.EXPECT().GetByID(gomock.Any(), 42).Return(&person.Person{ID: 42, LastName: "Тест", FirstName: "X"}, nil)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /people/{id}", h.GetByID)
	w := doRequest(mux, "GET", "/people/42", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	var resp model.PersonResponse
	decodeJSON(t, w, &resp)
	if resp.ID != 42 {
		t.Errorf("want id=42, got %d", resp.ID)
	}
}

func TestPersonGetByID_NotFound(t *testing.T) {
	svc, h := newPersonHandler(t)

	svc.EXPECT().GetByID(gomock.Any(), 99).Return(nil, errs.ErrNotFound)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /people/{id}", h.GetByID)
	w := doRequest(mux, "GET", "/people/99", nil)

	if w.Code != http.StatusNotFound {
		t.Errorf("want 404, got %d", w.Code)
	}
}

func TestPersonGetByID_InvalidID(t *testing.T) {
	_, h := newPersonHandler(t)

	r := httptest.NewRequest("GET", "/people/notanumber", nil)
	r.SetPathValue("id", "notanumber")
	w := httptest.NewRecorder()
	h.GetByID(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", w.Code)
	}
}

// POST /people

func TestPersonCreate_OK(t *testing.T) {
	svc, h := newPersonHandler(t)

	created := &person.Person{ID: 1, LastName: "Новый", FirstName: "Участник"}
	svc.EXPECT().Create(gomock.Any(), gomock.Any()).Return(created, nil)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /people", h.Create)
	w := doRequest(mux, "POST", "/people", map[string]string{
		"last_name": "Новый", "first_name": "Участник",
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d: %s", w.Code, w.Body.String())
	}
	var resp model.PersonResponse
	decodeJSON(t, w, &resp)
	if resp.ID != 1 {
		t.Errorf("want id=1, got %d", resp.ID)
	}
}

func TestPersonCreate_MissingFields(t *testing.T) {
	_, h := newPersonHandler(t)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /people", h.Create)
	w := doRequest(mux, "POST", "/people", map[string]string{"last_name": "Только фамилия"})

	if w.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", w.Code)
	}
}

func TestPersonCreate_Conflict(t *testing.T) {
	svc, h := newPersonHandler(t)

	svc.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errs.ErrConflict)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /people", h.Create)
	w := doRequest(mux, "POST", "/people", map[string]string{
		"last_name": "Дубль", "first_name": "Один",
	})

	if w.Code != http.StatusConflict {
		t.Errorf("want 409, got %d", w.Code)
	}
}

// PATCH /people/{id}

func TestPersonUpdate_InvalidID(t *testing.T) {
	_, h := newPersonHandler(t)

	r := httptest.NewRequest("PATCH", "/people/notanumber", nil)
	r.SetPathValue("id", "notanumber")
	w := httptest.NewRecorder()
	h.Update(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", w.Code)
	}
}

func TestPersonUpdate_MissingFields(t *testing.T) {
	_, h := newPersonHandler(t)

	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /people/{id}", h.Update)
	w := doRequest(mux, "PATCH", "/people/5", map[string]string{"last_name": "ТолькоФамилия"})

	if w.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", w.Code)
	}
}

func TestPersonUpdate_OK(t *testing.T) {
	svc, h := newPersonHandler(t)

	updated := &person.Person{ID: 5, LastName: "Изменён", FirstName: "Да"}
	svc.EXPECT().Update(gomock.Any(), gomock.Any()).Return(updated, nil)

	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /people/{id}", h.Update)
	w := doRequest(mux, "PATCH", "/people/5", map[string]string{
		"last_name": "Изменён", "first_name": "Да",
	})

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestPersonUpdate_NotFound(t *testing.T) {
	svc, h := newPersonHandler(t)

	svc.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, errs.ErrNotFound)

	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /people/{id}", h.Update)
	w := doRequest(mux, "PATCH", "/people/5", map[string]string{
		"last_name": "X", "first_name": "Y",
	})

	if w.Code != http.StatusNotFound {
		t.Errorf("want 404, got %d", w.Code)
	}
}

func TestPersonUpdate_Conflict(t *testing.T) {
	svc, h := newPersonHandler(t)

	svc.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, errs.ErrConflict)

	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /people/{id}", h.Update)
	w := doRequest(mux, "PATCH", "/people/5", map[string]string{
		"last_name": "X", "first_name": "Y",
	})

	if w.Code != http.StatusConflict {
		t.Errorf("want 409, got %d", w.Code)
	}
}

// DELETE /people/{id}

func TestPersonDelete_InvalidID(t *testing.T) {
	_, h := newPersonHandler(t)

	r := httptest.NewRequest("DELETE", "/people/notanumber", nil)
	r.SetPathValue("id", "notanumber")
	w := httptest.NewRecorder()
	h.Delete(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", w.Code)
	}
}

func TestPersonDelete_OK(t *testing.T) {
	svc, h := newPersonHandler(t)

	svc.EXPECT().Delete(gomock.Any(), 3).Return(nil)

	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /people/{id}", h.Delete)
	w := doRequest(mux, "DELETE", "/people/3", nil)

	if w.Code != http.StatusNoContent {
		t.Errorf("want 204, got %d", w.Code)
	}
}

func TestPersonDelete_NotFound(t *testing.T) {
	svc, h := newPersonHandler(t)

	svc.EXPECT().Delete(gomock.Any(), 99).Return(errs.ErrNotFound)

	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /people/{id}", h.Delete)
	w := doRequest(mux, "DELETE", "/people/99", nil)

	if w.Code != http.StatusNotFound {
		t.Errorf("want 404, got %d", w.Code)
	}
}

func TestPersonDelete_Conflict(t *testing.T) {
	svc, h := newPersonHandler(t)

	svc.EXPECT().Delete(gomock.Any(), 3).Return(errs.ErrConflict)

	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /people/{id}", h.Delete)
	w := doRequest(mux, "DELETE", "/people/3", nil)

	if w.Code != http.StatusConflict {
		t.Errorf("want 409, got %d", w.Code)
	}
}
