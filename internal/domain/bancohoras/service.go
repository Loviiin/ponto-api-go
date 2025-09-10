package bancohoras

import (
	"errors"
	"sort"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/domain/logbancohoras"
	"gorm.io/gorm"

	"github.com/Loviiin/ponto-api-go/internal/domain/ponto"
	"github.com/Loviiin/ponto-api-go/internal/domain/usuario"
	"github.com/Loviiin/ponto-api-go/internal/model"
)

type BancoHorasService interface {
	CalcularSaldoParaUsuario(usuarioID uint, empresaID uint, dia time.Time) (int, error)
	FecharDiaParaUsuario(usuarioID uint, empresaID uint, dia time.Time) (*model.Usuario, error)
}

type bancoHorasService struct {
	pontoRepo   ponto.RegistroPontoRepository
	usuarioRepo usuario.UsuarioRepository
	logRepo     logbancohoras.Repository // <-- ADICIONE O NOVO REPOSITÓRIO
	db          *gorm.DB
}

func NewBancoHorasService(
	pontoRepo ponto.RegistroPontoRepository,
	userRepo usuario.UsuarioRepository,
	logRepo logbancohoras.Repository,
	db *gorm.DB,
) BancoHorasService {
	return &bancoHorasService{
		pontoRepo:   pontoRepo,
		usuarioRepo: userRepo,
		logRepo:     logRepo,
		db:          db,
	}
}

func (s *bancoHorasService) CalcularSaldoParaUsuario(usuarioID uint, empresaID uint, dia time.Time) (int, error) {
	user, err := s.usuarioRepo.FindByID(usuarioID, empresaID)
	if err != nil {
		return 0, err
	}
	pontos, err := s.pontoRepo.FindPontosByUserIDAndDate(user.ID, dia)
	if err != nil {
		return 0, err
	}
	doDia, err := CalcularSaldoDoDia(pontos, user.Cargo)
	if err != nil {
		return 0, err
	}

	return doDia, err
}

func CalcularSaldoDoDia(pontosDoDia []model.RegistroPonto, cargoDoUsuario model.Cargo) (saldoEmMinutos int, err error) {
	sort.Slice(pontosDoDia, func(i, j int) bool {
		return pontosDoDia[i].Timestamp.Before(pontosDoDia[j].Timestamp)
	})

	if len(pontosDoDia)%2 == 1 {
		return 0, errors.New("o dia tem um número ímpar de marcações de ponto e não pode ser fechado automaticamente")
	}

	var totalTrabalhadoEmMinutos float64 = 0
	for i := 0; i < len(pontosDoDia)-1; i += 2 {
		entrada := pontosDoDia[i].Timestamp
		saida := pontosDoDia[i+1].Timestamp

		duracao := saida.Sub(entrada).Minutes()
		totalTrabalhadoEmMinutos += duracao
	}

	saldo := totalTrabalhadoEmMinutos - float64(cargoDoUsuario.CargaHorariaDiariaMinutos)

	return int(saldo), nil
}

func (s *bancoHorasService) FecharDiaParaUsuario(usuarioID uint, empresaID uint, dia time.Time) (*model.Usuario, error) {
	saldoDoDia, err := s.CalcularSaldoParaUsuario(usuarioID, empresaID, dia)
	if err != nil {
		return nil, err
	}

	usuarioAtual, err := s.usuarioRepo.FindByID(usuarioID, empresaID)
	if err != nil {
		return nil, err
	}

	novoSaldoTotal := usuarioAtual.SaldoBancoHorasMinutos + saldoDoDia

	err = s.db.Transaction(func(tx *gorm.DB) error {
		dadosParaAtualizar := map[string]interface{}{
			"saldo_banco_horas_minutos": novoSaldoTotal,
		}
		if err := s.usuarioRepo.Update(usuarioID, empresaID, dadosParaAtualizar); err != nil {
			return err
		}

		log := &model.LogBancoHoras{
			UsuarioID:            usuarioID,
			AutorID:              nil,
			EmpresaID:            empresaID,
			Data:                 time.Now(),
			ValorAlteradoMinutos: saldoDoDia,
			SaldoAnteriorMinutos: usuarioAtual.SaldoBancoHorasMinutos,
			SaldoNovoMinutos:     novoSaldoTotal,
			Motivo:               "Fechamento automático do dia " + dia.Format("2006-01-02"),
		}
		if err := s.logRepo.Create(log); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	usuarioAtual.SaldoBancoHorasMinutos = novoSaldoTotal
	return usuarioAtual, nil
}
