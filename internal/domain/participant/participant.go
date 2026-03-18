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
	FindByName(lastName, firstName, middleName string) (*Participant, error)
	GetByIDs(ids []int) ([]Participant, error)
}
