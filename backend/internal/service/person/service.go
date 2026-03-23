package person

import (
	"context"
	"strings"
	"unicode"

	"go.uber.org/zap"

	"meetings-editor/internal/domain/person"
	"meetings-editor/pkg/logger"
)

// titleCase trims whitespace and capitalises the first letter of every word,
// lowercasing the rest. Works correctly for multi-byte Unicode (Cyrillic, etc.).
func titleCase(s string) string {
	words := strings.Fields(strings.TrimSpace(s))
	for i, w := range words {
		runes := []rune(w)
		runes[0] = unicode.ToUpper(runes[0])
		for j := 1; j < len(runes); j++ {
			runes[j] = unicode.ToLower(runes[j])
		}
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}

func normaliseName(p *person.Person) {
	p.LastName = titleCase(p.LastName)
	p.FirstName = titleCase(p.FirstName)
	if p.MiddleName != "" {
		p.MiddleName = titleCase(p.MiddleName)
	}
}

type Service interface {
	GetAll(ctx context.Context) ([]person.Person, error)
	Search(ctx context.Context, q string) ([]person.Person, error)
	GetByID(ctx context.Context, id int) (*person.Person, error)
	SortByIDs(ctx context.Context, ids []int) ([]int, error)
	Create(ctx context.Context, p *person.Person) (*person.Person, error)
	Update(ctx context.Context, p *person.Person) (*person.Person, error)
	Delete(ctx context.Context, id int) error
}

type service struct {
	repo person.Repository
}

func New(repo person.Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetAll(ctx context.Context) ([]person.Person, error) {
	log := logger.FromContext(ctx)
	log.Info(ctx, "service: get all people")
	return s.repo.GetAll(ctx)
}

func (s *service) Search(ctx context.Context, q string) ([]person.Person, error) {
	log := logger.FromContext(ctx)
	log.Info(ctx, "service: search people", zap.String("q", q))
	words := strings.Fields(strings.ToLower(strings.TrimSpace(q)))
	if len(words) == 0 {
		return s.repo.GetAll(ctx)
	}
	return s.repo.Search(ctx, words)
}

func (s *service) GetByID(ctx context.Context, id int) (*person.Person, error) {
	log := logger.FromContext(ctx)
	log.Info(ctx, "service: get person by id", zap.Int("id", id))
	return s.repo.GetByID(ctx, id)
}

func (s *service) SortByIDs(ctx context.Context, ids []int) ([]int, error) {
	log := logger.FromContext(ctx)
	log.Info(ctx, "service: sort people by IDs", zap.Int("count", len(ids)))
	return s.repo.SortByIDs(ctx, ids)
}

func (s *service) Create(ctx context.Context, p *person.Person) (*person.Person, error) {
	log := logger.FromContext(ctx)
	normaliseName(p)
	log.Info(ctx, "service: creating person",
		zap.String("last_name", p.LastName),
		zap.String("first_name", p.FirstName),
	)
	return s.repo.Create(ctx, p)
}

func (s *service) Update(ctx context.Context, p *person.Person) (*person.Person, error) {
	log := logger.FromContext(ctx)
	normaliseName(p)
	log.Info(ctx, "service: updating person", zap.Int("id", p.ID))
	return s.repo.Update(ctx, p)
}

func (s *service) Delete(ctx context.Context, id int) error {
	log := logger.FromContext(ctx)
	log.Info(ctx, "service: deleting person", zap.Int("id", id))
	return s.repo.Delete(ctx, id)
}
