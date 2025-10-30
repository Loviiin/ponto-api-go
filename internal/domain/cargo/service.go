package cargo

import (
	"context"

	"github.com/Loviiin/ponto-api-go/internal/model"
)

// CargoService define a interface para os serviços de Cargo.
type CargoService interface {
	Create(cargo *model.Cargo) error
	FindByID(id uint, empresaID uint) (*model.Cargo, error)
	FindByName(nome string, empresaID uint) (*model.Cargo, error)
	GetAllByEmpresaID(empresaID uint) ([]model.Cargo, error)
	Update(id uint, empresaID uint, dados map[string]interface{}) error
	Delete(id uint, empresaID uint) error
	AddPermissionToCargo(cargoID uint, permissaoID uint, empresaID uint) error
	RemovePermissionFromCargo(cargoID uint, permissaoID uint, empresaID uint) error
	GetPermissionsByCargo(cargoID uint, empresaID uint) ([]model.Permissao, error)
	HasUsuarios(cargoID uint, empresaID uint) (bool, error)
}

type cargoService struct {
	repo CargoRepository
}

func NewCargoService(repo CargoRepository) CargoService {
	return &cargoService{repo: repo}
}

func (s *cargoService) Create(cargo *model.Cargo) error {
	return s.repo.Create(cargo)
}

func (s *cargoService) FindByID(id uint, empresaID uint) (*model.Cargo, error) {
	return s.repo.FindByID(context.Background(), id, empresaID)
}

func (s *cargoService) FindByName(nome string, empresaID uint) (*model.Cargo, error) {
	return s.repo.FindByName(nome, empresaID)
}

func (s *cargoService) GetAllByEmpresaID(empresaID uint) ([]model.Cargo, error) {
	return s.repo.GetAllByEmpresaID(empresaID)
}

func (s *cargoService) Update(id uint, empresaID uint, dados map[string]interface{}) error {
	_, err := s.repo.FindByID(context.Background(), id, empresaID)
	if err != nil {
		return err // Retorna o erro (ex: not found)
	}
	return s.repo.Update(id, empresaID, dados)
}

func (s *cargoService) Delete(id uint, empresaID uint) error {
	_, err := s.repo.FindByID(context.Background(), id, empresaID)
	if err != nil {
		return err
	}
	return s.repo.Delete(id, empresaID)
}

// HasUsuarios verifica se existem usuários associados ao cargo.
func (s *cargoService) HasUsuarios(cargoID uint, empresaID uint) (bool, error) {
	return s.repo.HasUsuarios(cargoID, empresaID)
}

func (s *cargoService) AddPermissionToCargo(cargoID uint, permissaoID uint, empresaID uint) error {
	_, err := s.repo.FindByID(context.Background(), cargoID, empresaID)
	if err != nil {
		return err
	}
	return s.repo.AddPermissionToCargo(cargoID, permissaoID)
}

func (s *cargoService) RemovePermissionFromCargo(cargoID uint, permissaoID uint, empresaID uint) error {
	_, err := s.repo.FindByID(context.Background(), cargoID, empresaID)
	if err != nil {
		return err
	}
	return s.repo.RemovePermissionFromCargo(cargoID, permissaoID)
}

func (s *cargoService) GetPermissionsByCargo(cargoID uint, empresaID uint) ([]model.Permissao, error) {
	return s.repo.GetPermissionsByCargo(cargoID, empresaID)
}
