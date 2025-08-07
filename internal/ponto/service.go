package ponto

import (
	"github.com/Loviiin/ponto-api-go/internal/model"
	"time"
)

type PontoService interface {
	BaterPonto(usuarioID uint, latitude, longitude float64) (*model.RegistroPonto, error)
}

type pontoService struct {
	pontoRepo RegistroPontoRepository
}

func NewPontoService(repo RegistroPontoRepository) PontoService {
	return &pontoService{
		pontoRepo: repo,
	}
}

func (s *pontoService) BaterPonto(usuarioID uint, latitude, longitude float64) (*model.RegistroPonto, error) {
	registroPonto := &model.RegistroPonto{
		UsuarioID: usuarioID,
		Latitude:  latitude,
		Longitude: longitude,
		Timestamp: time.Now(),
	}

	err := s.pontoRepo.SavePonto(registroPonto)
	if err != nil {
		return nil, err

	}

	return registroPonto, nil
}
