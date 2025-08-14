package permissao

import "github.com/Loviiin/ponto-api-go/internal/model"

type Service interface {
	Create(permissao *model.Permissao) error
	FindAll() ([]model.Permissao, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(permissao *model.Permissao) error {
	return s.repo.Create(permissao)
}

func (s *service) FindAll() ([]model.Permissao, error) {
	return s.repo.FindAll()
}
