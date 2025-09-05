package ponto

import (
	"errors"
	"github.com/Loviiin/ponto-api-go/internal/domain/empresa"
	"github.com/Loviiin/ponto-api-go/internal/domain/usuario"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/umahmood/haversine"
	"gorm.io/gorm"
	"time"
)

type PontoService interface {
	BaterPonto(usuarioID uint, empresaID uint, latitude, longitude float64) (*model.RegistroPonto, error)
	GetPontosDoDia(usuarioID uint, dia time.Time) ([]model.RegistroPonto, error)
	AjustarPonto(usuarioID, empresaID, adminID uint, timestamp time.Time, justificativaID *uint) (*model.RegistroPonto, error)
	EditarPonto(pontoID, empresaID uint, novoTimestamp time.Time, justificativaID *uint) (*model.RegistroPonto, error)
}

type pontoService struct {
	pontoRepo   RegistroPontoRepository
	empresaRepo empresa.EmpresaRepository
	userRepo    usuario.UsuarioRepository
	db          *gorm.DB
}

func NewPontoService(
	pontoRepo RegistroPontoRepository,
	userRepo usuario.UsuarioRepository,
	empresaRepo empresa.EmpresaRepository,
	db *gorm.DB,
) PontoService {
	return &pontoService{
		pontoRepo:   pontoRepo,
		userRepo:    userRepo,
		empresaRepo: empresaRepo,
		db:          db,
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
		Metodo:    tipoBatida,
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

func (s *pontoService) AjustarPonto(usuarioID, empresaID, adminID uint, timestamp time.Time, justificativaID *uint) (*model.RegistroPonto, error) {
	_, err := s.userRepo.FindByID(usuarioID, empresaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("usuário alvo não encontrado ou não pertence a esta empresa")
		}
		return nil, err
	}

	pontoRegistrado := &model.RegistroPonto{
		UsuarioID:       usuarioID,
		EmpresaID:       empresaID,
		Timestamp:       timestamp,
		Metodo:          "AJUSTE_MANUAL_ADMIN",
		Localizacao:     "N/A",
		JustificativaID: justificativaID, // Agora apenas associamos um ID existente
	}

	if err := s.pontoRepo.SavePonto(pontoRegistrado); err != nil {
		return nil, err
	}

	return pontoRegistrado, nil
}
func (s *pontoService) EditarPonto(pontoID, empresaID uint, novoTimestamp time.Time, justificativaID *uint) (*model.RegistroPonto, error) {
	pontoParaEditar, err := s.pontoRepo.FindPontoByID(pontoID, empresaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("registro de ponto não encontrado ou não pertence a esta empresa")
		}
		return nil, err
	}

	pontoParaEditar.Timestamp = novoTimestamp
	pontoParaEditar.Metodo = "EDICAO_MANUAL_ADMIN"
	pontoParaEditar.JustificativaID = justificativaID // Agora associamos a justificativa

	if err := s.pontoRepo.UpdatePonto(pontoParaEditar); err != nil {
		return nil, err
	}

	return pontoParaEditar, nil
}
