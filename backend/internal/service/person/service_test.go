package person

import (
	"testing"

	"go.uber.org/mock/gomock"

	"meetings-editor/internal/domain/person"
	"meetings-editor/internal/mocks"
	"meetings-editor/internal/testutil"
	"meetings-editor/pkg/errs"
)

func setup(t *testing.T) (*mocks.MockPersonRepository, Service) {
	t.Helper()
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockPersonRepository(ctrl)
	svc := New(repo)
	return repo, svc
}

func TestSearch_EmptyQuery_CallsGetAll(t *testing.T) {
	repo, svc := setup(t)
	ctx := testutil.Ctx()

	want := []person.Person{{ID: 1, LastName: "Иванов", FirstName: "Иван"}}
	repo.EXPECT().GetAll(ctx).Return(want, nil)

	got, err := svc.Search(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != 1 {
		t.Errorf("unexpected result: %v", got)
	}
}

func TestSearch_WhitespaceOnly_CallsGetAll(t *testing.T) {
	repo, svc := setup(t)
	ctx := testutil.Ctx()

	repo.EXPECT().GetAll(ctx).Return(nil, nil)
	_, err := svc.Search(ctx, "   ")
	if err != nil {
		t.Fatal(err)
	}
}

func TestSearch_WithQuery_CallsSearchWithWords(t *testing.T) {
	repo, svc := setup(t)
	ctx := testutil.Ctx()

	// "Иван Петров" → ["иван", "петров"]
	repo.EXPECT().Search(ctx, []string{"иван", "петров"}).Return(nil, nil)
	_, err := svc.Search(ctx, "  Иван  Петров  ")
	if err != nil {
		t.Fatal(err)
	}
}

func TestSearch_SingleWord_CallsSearchWithOneWord(t *testing.T) {
	repo, svc := setup(t)
	ctx := testutil.Ctx()

	repo.EXPECT().Search(ctx, []string{"иванов"}).Return(nil, nil)
	_, err := svc.Search(ctx, "Иванов")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetAll_Delegates(t *testing.T) {
	repo, svc := setup(t)
	ctx := testutil.Ctx()

	want := []person.Person{{ID: 5}}
	repo.EXPECT().GetAll(ctx).Return(want, nil)

	got, err := svc.GetAll(ctx)
	if err != nil || len(got) != 1 || got[0].ID != 5 {
		t.Errorf("unexpected: %v, %v", got, err)
	}
}

func TestGetByID_Found(t *testing.T) {
	repo, svc := setup(t)
	ctx := testutil.Ctx()

	want := &person.Person{ID: 7, LastName: "Тест"}
	repo.EXPECT().GetByID(ctx, 7).Return(want, nil)

	got, err := svc.GetByID(ctx, 7)
	if err != nil || got.ID != 7 {
		t.Errorf("unexpected: %v, %v", got, err)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	repo, svc := setup(t)
	ctx := testutil.Ctx()

	repo.EXPECT().GetByID(ctx, 99).Return(nil, errs.ErrNotFound)

	_, err := svc.GetByID(ctx, 99)
	if err != errs.ErrNotFound {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestCreate_Delegates(t *testing.T) {
	repo, svc := setup(t)
	ctx := testutil.Ctx()

	p := &person.Person{LastName: "Новый", FirstName: "Участник"}
	created := &person.Person{ID: 10, LastName: "Новый", FirstName: "Участник"}
	repo.EXPECT().Create(ctx, p).Return(created, nil)

	got, err := svc.Create(ctx, p)
	if err != nil || got.ID != 10 {
		t.Errorf("unexpected: %v, %v", got, err)
	}
}

func TestCreate_Conflict(t *testing.T) {
	repo, svc := setup(t)
	ctx := testutil.Ctx()

	p := &person.Person{LastName: "Дубль", FirstName: "Один"}
	repo.EXPECT().Create(ctx, p).Return(nil, errs.ErrConflict)

	_, err := svc.Create(ctx, p)
	if err != errs.ErrConflict {
		t.Errorf("want ErrConflict, got %v", err)
	}
}

func TestUpdate_Delegates(t *testing.T) {
	repo, svc := setup(t)
	ctx := testutil.Ctx()

	p := &person.Person{ID: 3, LastName: "Обновлён", FirstName: "Да"}
	repo.EXPECT().Update(ctx, p).Return(p, nil)

	got, err := svc.Update(ctx, p)
	if err != nil || got.LastName != "Обновлён" {
		t.Errorf("unexpected: %v, %v", got, err)
	}
}

func TestDelete_Delegates(t *testing.T) {
	repo, svc := setup(t)
	ctx := testutil.Ctx()

	repo.EXPECT().Delete(ctx, 4).Return(nil)

	if err := svc.Delete(ctx, 4); err != nil {
		t.Fatal(err)
	}
}

func TestDelete_Conflict(t *testing.T) {
	repo, svc := setup(t)
	ctx := testutil.Ctx()

	repo.EXPECT().Delete(ctx, 4).Return(errs.ErrConflict)

	if err := svc.Delete(ctx, 4); err != errs.ErrConflict {
		t.Errorf("want ErrConflict, got %v", err)
	}
}
