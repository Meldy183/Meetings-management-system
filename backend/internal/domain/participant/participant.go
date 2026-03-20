package participant

import "context"

type Participant struct {
	ID         int
	LastName   string
	FirstName  string
	MiddleName string
	Info       string
}

type Repository interface {
	GetAll(ctx context.Context) ([]Participant, error)
	GetByIDs(ctx context.Context, ids []int) ([]Participant, error)
	Create(ctx context.Context, p *Participant) (*Participant, error)
	Update(ctx context.Context, p *Participant) (*Participant, error)
	Delete(ctx context.Context, id int) error
}
