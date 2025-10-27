// Em: internal/domain/localidade/service.go
package localidade

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/cache"
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
	cache          cache.Service
}

func NewService(repo Repository, geoSvc geolocation.Service, cacheService cache.Service) Service {
	return &service{
		repo:           repo,
		geolocationSvc: geoSvc,
		cache:          cacheService,
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

	if err := s.repo.Save(dadosCompletos); err != nil {
		return err
	}

	// Invalida o cache da empresa após criar nova localidade
	s.invalidateCache(localidadeParcial.EmpresaID)
	return nil
}

func (s *service) FindByID(id uint) (*model.Localidade, error) {
	return s.repo.FindByID(id)
}

func (s *service) FindAllByEmpresaID(empresaID uint) ([]model.Localidade, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("localidades:empresa:%d", empresaID)

	// Tenta buscar do cache
	if s.cache != nil {
		cachedData, err := s.cache.Get(ctx, cacheKey)
		if err == nil && cachedData != "" {
			var localidades []model.Localidade
			if err := json.Unmarshal([]byte(cachedData), &localidades); err == nil {
				return localidades, nil
			}
		}
	}

	// Se não encontrou no cache, busca do banco
	localidades, err := s.repo.FindAllByEmpresaID(empresaID)
	if err != nil {
		return nil, err
	}

	// Salva no cache por 10 minutos
	if s.cache != nil {
		if data, err := json.Marshal(localidades); err == nil {
			_ = s.cache.Set(ctx, cacheKey, string(data), 10*time.Minute)
		}
	}

	return localidades, nil
}

// invalidateCache remove o cache de localidades de uma empresa
func (s *service) invalidateCache(empresaID uint) {
	if s.cache != nil {
		ctx := context.Background()
		cacheKey := fmt.Sprintf("localidades:empresa:%d", empresaID)
		_ = s.cache.Delete(ctx, cacheKey)
	}
}
