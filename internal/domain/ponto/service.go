package ponto

import (
	"github.com/Loviiin/ponto-api-go/internal/domain/empresa"
	"github.com/Loviiin/ponto-api-go/internal/domain/usuario"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/umahmood/haversine"
	"time"
)

type PontoService interface {
	BaterPonto(usuarioID uint, empresaID uint, latitude, longitude float64) (*model.RegistroPonto, error)
	GetPontosDoDia(usuarioID uint, dia time.Time) ([]model.RegistroPonto, error)
}

type pontoService struct {
	pontoRepo   RegistroPontoRepository
	empresaRepo empresa.EmpresaRepository
	userRepo    usuario.UsuarioRepository
}

func NewPontoService(
	pontoRepo RegistroPontoRepository,
	userRepo usuario.UsuarioRepository,
	empresaRepo empresa.EmpresaRepository,
) PontoService {
	return &pontoService{
		pontoRepo:   pontoRepo,
		userRepo:    userRepo,
		empresaRepo: empresaRepo,
	}
}

func (s *pontoService) BaterPonto(usuarioID uint, empresaID uint, latitude, longitude float64) (*model.RegistroPonto, error) {
	_, err := s.userRepo.FindByID(usuarioID, empresaID)
	if err != nil {
		return nil, err
	}

	dadoEmpresa, err := s.empresaRepo.FindByID(empresaID)
	if err != nil {
		return nil, err
	}

	pontoSede := haversine.Coord{Lat: dadoEmpresa.SedeLatitude, Lon: dadoEmpresa.SedeLongitude}
	pontoBatida := haversine.Coord{Lat: latitude, Lon: longitude}

	km, _ := haversine.Distance(pontoSede, pontoBatida)
	distanciaEmMetros := km * 1000

	var tipoBatida string

	if distanciaEmMetros > dadoEmpresa.RaioGeofenceMetros {
		tipoBatida = "Remoto"
	} else {
		tipoBatida = "Presencial"
	}

	registroPonto := &model.RegistroPonto{
		UsuarioID: usuarioID,
		Latitude:  latitude,
		Longitude: longitude,
		Timestamp: time.Now(),
		EmpresaID: empresaID,
		Tipo:      tipoBatida,
	}

	err = s.pontoRepo.SavePonto(registroPonto)
	if err != nil {
		return nil, err
	}

	return registroPonto, nil
}

func (s *pontoService) GetPontosDoDia(usuarioID uint, dia time.Time) ([]model.RegistroPonto, error) {
	return s.pontoRepo.FindPontosByUserIDAndDate(usuarioID, dia)
}
