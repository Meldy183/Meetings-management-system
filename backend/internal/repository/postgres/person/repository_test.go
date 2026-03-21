package person

import (
	"testing"

	"meetings-editor/internal/domain/person"
	"meetings-editor/internal/testutil"
	"meetings-editor/pkg/errs"
)

func TestPersonRepo_CreateAndGetByID(t *testing.T) {
	pool := testutil.NewDB(t)
	testutil.TruncateTables(t, pool)
	repo := New(pool)
	ctx := testutil.Ctx()

	p := &person.Person{LastName: "Иванов", FirstName: "Иван", MiddleName: "Иванович", Info: "Директор"}
	created, err := repo.Create(ctx, p)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero ID after create")
	}

	got, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.LastName != "Иванов" || got.FirstName != "Иван" || got.MiddleName != "Иванович" {
		t.Errorf("unexpected person: %+v", got)
	}
}

func TestPersonRepo_Create_Conflict(t *testing.T) {
	pool := testutil.NewDB(t)
	testutil.TruncateTables(t, pool)
	repo := New(pool)
	ctx := testutil.Ctx()

	p := &person.Person{LastName: "Дубль", FirstName: "Иван", MiddleName: ""}
	if _, err := repo.Create(ctx, p); err != nil {
		t.Fatalf("first Create: %v", err)
	}
	_, err := repo.Create(ctx, &person.Person{LastName: "Дубль", FirstName: "Иван", MiddleName: ""})
	if err != errs.ErrConflict {
		t.Errorf("want ErrConflict on duplicate, got %v", err)
	}
}

func TestPersonRepo_GetByID_NotFound(t *testing.T) {
	pool := testutil.NewDB(t)
	testutil.TruncateTables(t, pool)
	repo := New(pool)
	ctx := testutil.Ctx()

	_, err := repo.GetByID(ctx, 9999999)
	if err != errs.ErrNotFound {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestPersonRepo_GetAll_OrderedByName(t *testing.T) {
	pool := testutil.NewDB(t)
	testutil.TruncateTables(t, pool)
	repo := New(pool)
	ctx := testutil.Ctx()

	names := []person.Person{
		{LastName: "Яблоков", FirstName: "Я"},
		{LastName: "Абрамов", FirstName: "А"},
		{LastName: "Иванов", FirstName: "И"},
	}
	for i := range names {
		if _, err := repo.Create(ctx, &names[i]); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	all, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("want 3, got %d", len(all))
	}
	if all[0].LastName != "Абрамов" || all[1].LastName != "Иванов" || all[2].LastName != "Яблоков" {
		t.Errorf("unexpected order: %v %v %v", all[0].LastName, all[1].LastName, all[2].LastName)
	}
}

func TestPersonRepo_Search_MatchesByLastName(t *testing.T) {
	pool := testutil.NewDB(t)
	testutil.TruncateTables(t, pool)
	repo := New(pool)
	ctx := testutil.Ctx()

	for _, p := range []person.Person{
		{LastName: "Иванов", FirstName: "Иван"},
		{LastName: "Петров", FirstName: "Пётр"},
	} {
		pp := p
		if _, err := repo.Create(ctx, &pp); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	results, err := repo.Search(ctx, []string{"иванов"})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 || results[0].LastName != "Иванов" {
		t.Errorf("unexpected search result: %v", results)
	}
}

func TestPersonRepo_Search_MultipleWords(t *testing.T) {
	pool := testutil.NewDB(t)
	testutil.TruncateTables(t, pool)
	repo := New(pool)
	ctx := testutil.Ctx()

	people := []person.Person{
		{LastName: "Иванов", FirstName: "Иван", MiddleName: "Иванович"},
		{LastName: "Иванов", FirstName: "Пётр", MiddleName: ""},
	}
	for i := range people {
		if _, err := repo.Create(ctx, &people[i]); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	results, err := repo.Search(ctx, []string{"иванов", "иван"})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	// Should only match the one with both "иванов" and "иван" in their name
	if len(results) != 1 || results[0].FirstName != "Иван" {
		t.Errorf("unexpected multi-word search result: %v", results)
	}
}

func TestPersonRepo_Update(t *testing.T) {
	pool := testutil.NewDB(t)
	testutil.TruncateTables(t, pool)
	repo := New(pool)
	ctx := testutil.Ctx()

	created, _ := repo.Create(ctx, &person.Person{LastName: "Старый", FirstName: "Имя"})
	created.LastName = "Новый"
	created.Info = "обновлён"

	updated, err := repo.Update(ctx, created)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.LastName != "Новый" || updated.Info != "обновлён" {
		t.Errorf("unexpected update result: %+v", updated)
	}
}

func TestPersonRepo_Update_NotFound(t *testing.T) {
	pool := testutil.NewDB(t)
	testutil.TruncateTables(t, pool)
	repo := New(pool)
	ctx := testutil.Ctx()

	_, err := repo.Update(ctx, &person.Person{ID: 9999999, LastName: "X", FirstName: "Y"})
	if err != errs.ErrNotFound {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestPersonRepo_Delete(t *testing.T) {
	pool := testutil.NewDB(t)
	testutil.TruncateTables(t, pool)
	repo := New(pool)
	ctx := testutil.Ctx()

	created, _ := repo.Create(ctx, &person.Person{LastName: "Удалить", FirstName: "Меня"})

	if err := repo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.GetByID(ctx, created.ID); err != errs.ErrNotFound {
		t.Errorf("after delete: want ErrNotFound, got %v", err)
	}
}

func TestPersonRepo_Delete_NotFound(t *testing.T) {
	pool := testutil.NewDB(t)
	testutil.TruncateTables(t, pool)
	repo := New(pool)
	ctx := testutil.Ctx()

	if err := repo.Delete(ctx, 9999999); err != errs.ErrNotFound {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestPersonRepo_GetByIDs(t *testing.T) {
	pool := testutil.NewDB(t)
	testutil.TruncateTables(t, pool)
	repo := New(pool)
	ctx := testutil.Ctx()

	p1, _ := repo.Create(ctx, &person.Person{LastName: "А", FirstName: "А"})
	p2, _ := repo.Create(ctx, &person.Person{LastName: "Б", FirstName: "Б"})
	_, _ = repo.Create(ctx, &person.Person{LastName: "В", FirstName: "В"})

	got, err := repo.GetByIDs(ctx, []int{p1.ID, p2.ID})
	if err != nil {
		t.Fatalf("GetByIDs: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("want 2 results, got %d", len(got))
	}
}
