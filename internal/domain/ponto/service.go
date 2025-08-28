package ponto

import (
	"errors"
	"github.com/Loviiin/ponto-api-go/internal/domain/empresa"
	"github.com/Loviiin/ponto-api-go/internal/domain/justificativa"
	"github.com/Loviiin/ponto-api-go/internal/domain/usuario"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/umahmood/haversine"
	"gorm.io/gorm"
	"time"
)

type PontoService interface {
	BaterPonto(usuarioID uint, empresaID uint, latitude, longitude float64) (*model.RegistroPonto, error)
	GetPontosDoDia(usuarioID uint, dia time.Time) ([]model.RegistroPonto, error)
	AjustarPonto(usuarioID, empresaID, adminID uint, timestamp time.Time, justificativaDescricao string) (*model.RegistroPonto, error)
	EditarPonto(pontoID, empresaID, adminID uint, novoTimestamp time.Time, justificativaDescricao string) (*model.RegistroPonto, error)
}

type pontoService struct {
	pontoRepo         RegistroPontoRepository
	empresaRepo       empresa.EmpresaRepository
	userRepo          usuario.UsuarioRepository
	justificativaRepo justificativa.Repository
	db                *gorm.DB
}

func NewPontoService(
	pontoRepo RegistroPontoRepository,
	userRepo usuario.UsuarioRepository,
	empresaRepo empresa.EmpresaRepository,
	justificativaRepo justificativa.Repository,
	db *gorm.DB,
) PontoService {
	return &pontoService{
		pontoRepo:         pontoRepo,
		userRepo:          userRepo,
		empresaRepo:       empresaRepo,
		justificativaRepo: justificativaRepo,
		db:                db,
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

func (s *pontoService) AjustarPonto(usuarioID, empresaID, adminID uint, timestamp time.Time, justificativaDescricao string) (*model.RegistroPonto, error) {
	_, err := s.userRepo.FindByID(usuarioID, empresaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("usuário alvo não encontrado ou não pertence a esta empresa")
		}
		return nil, err
	}

	var pontoRegistrado *model.RegistroPonto

	err = s.db.Transaction(func(tx *gorm.DB) error {
		justificativaRepoTx := s.justificativaRepo.WithTransaction(tx)
		pontoRepoTx := s.pontoRepo.WithTransaction(tx)

		// 3. Criar a Justificativa
		novaJustificativa := &model.Justificativa{
			DataOcorrencia: timestamp,
			Tipo:           "AJUSTE_PONTO",
			Descricao:      justificativaDescricao,
			Status:         "APROVADO",
			UsuarioID:      usuarioID,
			AprovadorID:    &adminID,
			EmpresaID:      empresaID,
		}
		if err := justificativaRepoTx.Create(novaJustificativa); err != nil {
			return err
		}

		pontoRegistrado = &model.RegistroPonto{
			UsuarioID:       usuarioID,
			EmpresaID:       empresaID,
			Timestamp:       timestamp,
			Metodo:          "AJUSTE_MANUAL_ADMIN",
			Localizacao:     "N/A",
			JustificativaID: &novaJustificativa.ID,
		}
		if err := pontoRepoTx.SavePonto(pontoRegistrado); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return pontoRegistrado, nil
}

func (s *pontoService) EditarPonto(pontoID, empresaID, adminID uint, novoTimestamp time.Time, justificativaDescricao string) (*model.RegistroPonto, error) {
	var pontoAtualizado *model.RegistroPonto
	
	err := s.db.Transaction(func(tx *gorm.DB) error {
		pontoRepoTx := s.pontoRepo.WithTransaction(tx)
		justificativaRepoTx := s.justificativaRepo.WithTransaction(tx)

		pontoParaEditar, err := pontoRepoTx.FindPontoByID(pontoID, empresaID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("registro de ponto não encontrado ou não pertence a esta empresa")
			}
			return err
		}

		novaJustificativa := &model.Justificativa{
			DataOcorrencia: novoTimestamp,
			Tipo:           "EDICAO_PONTO",
			Descricao:      justificativaDescricao,
			Status:         "APROVADO",
			UsuarioID:      pontoParaEditar.UsuarioID,
			AprovadorID:    &adminID,
			EmpresaID:      empresaID,
		}
		if err := justificativaRepoTx.Create(novaJustificativa); err != nil {
			return err
		}

		pontoParaEditar.Timestamp = novoTimestamp
		pontoParaEditar.Metodo = "AJUSTE_MANUAL_ADMIN"
		pontoParaEditar.JustificativaID = &novaJustificativa.ID

		if err := pontoRepoTx.UpdatePonto(pontoParaEditar); err != nil {
			return err
		}

		pontoAtualizado = pontoParaEditar
		return nil
	})

	if err != nil {
		return nil, err
	}

	return pontoAtualizado, nil
}
