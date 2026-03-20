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
	Create(ctx context.Context, p *Participant) (*Participant, error)
	FindByName(ctx context.Context, lastName, firstName, middleName string) (*Participant, error)
	GetByIDs(ctx context.Context, ids []int) ([]Participant, error)
	Update(ctx context.Context, p *Participant) (*Participant, error)
	Delete(ctx context.Context, id int) error
}
