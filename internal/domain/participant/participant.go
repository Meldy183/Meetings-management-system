package participant

// implement me
type Participant struct {
	ID         int
	LastName   string
	FirstName  string
	MiddleName string
	Info       string
}
type Repository interface {
	// implement me
	FindByName(lastName, firstName, middleName string) (*Participant, error)
	Create(p *Participant) (*Participant, error)
	GetByIDs(ids []int) ([]Participant, error)
}
