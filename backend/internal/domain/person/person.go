package person

import "context"

type Person struct {
	ID         int
	LastName   string
	FirstName  string
	MiddleName string
	Info       string
}

type Repository interface {
	GetAll(ctx context.Context) ([]Person, error)
	Search(ctx context.Context, words []string) ([]Person, error)
	GetByID(ctx context.Context, id int) (*Person, error)
	GetByIDs(ctx context.Context, ids []int) ([]Person, error)
	SortByIDs(ctx context.Context, ids []int) ([]int, error)
	Create(ctx context.Context, p *Person) (*Person, error)
	Update(ctx context.Context, p *Person) (*Person, error)
	Delete(ctx context.Context, id int) error
}
