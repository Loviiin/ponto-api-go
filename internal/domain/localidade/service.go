// Em: internal/domain/localidade/service.go
package localidade

import (

	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/geolocation"
)

type Service interface {
	Create(localidadeParcial *model.Localidade) error
	FindByID(id uint) (*model.Localidade, error)
	FindAllByEmpresaID(empresaID uint) ([]model.Localidade, error)
	// Adicione Update e Delete se necessário
}

type service struct {
	repo           Repository
	geolocationSvc geolocation.Service
}

func NewService(repo Repository, geoSvc geolocation.Service) Service {
	return &service{
		repo:           repo,
		geolocationSvc: geoSvc,
	}
}

func (s *service) Create(localidadeParcial *model.Localidade) error {
	dadosCompletos, err := s.geolocationSvc.GetLocationFromCEP(localidadeParcial.CEP)
	if err != nil {
		return err
	}

	dadosCompletos.Nome = localidadeParcial.Nome
	dadosCompletos.EmpresaID = localidadeParcial.EmpresaID
	dadosCompletos.RaioGeofenceMetros = localidadeParcial.RaioGeofenceMetros

	return s.repo.Save(dadosCompletos)
}

func (s *service) FindByID(id uint) (*model.Localidade, error) {
	return s.repo.FindByID(id)
}

func (s *service) FindAllByEmpresaID(empresaID uint) ([]model.Localidade, error) {
	return s.repo.FindAllByEmpresaID(empresaID)
}