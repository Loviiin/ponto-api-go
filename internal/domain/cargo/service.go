package cargo

import (
	"github.com/Loviiin/ponto-api-go/internal/model"
)

// CargoService define a interface para os serviços de Cargo.
type CargoService interface {
	Create(cargo *model.Cargo, empresaID uint) error
	FindByID(id uint, empresaID uint) (*model.Cargo, error)
	GetAllByEmpresaID(empresaID uint) ([]model.Cargo, error)
	Update(id uint, empresaID uint, dados map[string]interface{}) error
	Delete(id uint, empresaID uint) error
}

type cargoService struct {
	repo CargoRepository
}

func NewCargoService(repo CargoRepository) CargoService {
	return &cargoService{repo: repo}
}

func (s *cargoService) Create(cargo *model.Cargo, empresaID uint) error {
	cargo.EmpresaID = empresaID
	return s.repo.Create(cargo)
}

func (s *cargoService) FindByID(id uint, empresaID uint) (*model.Cargo, error) {
	return s.repo.FindByID(id, empresaID)
}

func (s *cargoService) GetAllByEmpresaID(empresaID uint) ([]model.Cargo, error) {
	return s.repo.GetAllByEmpresaID(empresaID)
}

func (s *cargoService) Update(id uint, empresaID uint, dados map[string]interface{}) error {
	_, err := s.repo.FindByID(id, empresaID)
	if err != nil {
		return err // Retorna o erro (ex: not found)
	}
	return s.repo.Update(id, empresaID, dados)
}

func (s *cargoService) Delete(id uint, empresaID uint) error {
	_, err := s.repo.FindByID(id, empresaID)
	if err != nil {
		return err
	}
	return s.repo.Delete(id, empresaID)
}
