package bancohoras

import (
	"context"
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
	FecharDiaParaUsuario(usuarioID uint, empresaID uint, dia time.Time) (*model.Contrato, error)
	GetDashboardForUsuario(usuarioID uint, empresaID uint) (*DashboardResponse, error)
}

type bancoHorasService struct {
	pontoRepo   ponto.RegistroPontoRepository
	usuarioRepo usuario.UsuarioRepository
	logRepo     logbancohoras.Repository
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
	user, err := s.usuarioRepo.FindByID(context.Background(), usuarioID, empresaID)
	if err != nil {
		return 0, err
	}

	if user.Contrato.ID == 0 || user.Contrato.Cargo.ID == 0 {
		return 0, errors.New("utilizador não possui um contrato ou cargo ativo para calcular o saldo")
	}

	pontos, err := s.pontoRepo.FindPontosByUserIDAndDate(user.ID, dia)
	if err != nil {
		return 0, err
	}

	doDia, err := CalcularSaldoDoDia(pontos, user.Contrato.Cargo)
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

func (s *bancoHorasService) FecharDiaParaUsuario(usuarioID uint, empresaID uint, dia time.Time) (*model.Contrato, error) {
	usuarioAtual, err := s.usuarioRepo.FindByID(context.Background(), usuarioID, empresaID)
	if err != nil {
		return nil, err
	}
	if usuarioAtual.Contrato.ID == 0 {
		return nil, errors.New("utilizador não possui um contrato para fechar o dia")
	}

	saldoDoDia, err := s.CalcularSaldoParaUsuario(usuarioID, empresaID, dia)
	if err != nil {
		return nil, err
	}

	saldoAnterior := usuarioAtual.Contrato.SaldoBancoHorasMinutos
	novoSaldoTotal := saldoAnterior + saldoDoDia

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Contrato{}).Where("id = ?", usuarioAtual.Contrato.ID).
			Update("saldo_banco_horas_minutos", novoSaldoTotal).Error; err != nil {
			return err
		}

		// Cria o log do banco de horas
		log := &model.LogBancoHoras{
			UsuarioID:            usuarioID,
			AutorID:              nil, // Fechamento automático não tem autor
			EmpresaID:            empresaID,
			Data:                 time.Now(),
			ValorAlteradoMinutos: saldoDoDia,
			SaldoAnteriorMinutos: saldoAnterior,
			SaldoNovoMinutos:     novoSaldoTotal,
			Motivo:               "Fechamento automático do dia " + dia.Format("2006-01-02"),
		}
		// Usa o logRepo dentro da transação
		if err := s.logRepo.WithTransaction(tx).Create(log); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Retorna o contrato atualizado
	usuarioAtual.Contrato.SaldoBancoHorasMinutos = novoSaldoTotal
	return &usuarioAtual.Contrato, nil
}

// GetDashboardForUsuario retorna o saldo total atual e o histórico de alterações do banco de horas
// do usuário informado já ordenado por data desc.
func (s *bancoHorasService) GetDashboardForUsuario(usuarioID uint, empresaID uint) (*DashboardResponse, error) {
	// a) Buscar o usuário (com contrato)
	user, err := s.usuarioRepo.FindByID(context.Background(), usuarioID, empresaID)
	if err != nil {
		return nil, err
	}

	// b) Buscar todos os logs do usuário ordenados por data desc
	logs, err := s.logRepo.GetAllByUsuarioAndEmpresa(usuarioID, empresaID)
	if err != nil {
		return nil, err
	}

	// Mesmo com a query ordenada, garantimos ordenação desc por segurança
	sort.SliceStable(logs, func(i, j int) bool { return logs[i].Data.After(logs[j].Data) })

	// c) Mapear para DTO
	historico := make([]HistoricoDia, 0, len(logs))
	for _, l := range logs {
		historico = append(historico, HistoricoDia{
			Data:                   l.Data.Format("2006-01-02"),
			ValorAlteradoMinutos:   l.ValorAlteradoMinutos,
			SaldoResultanteMinutos: l.SaldoNovoMinutos,
			Motivo:                 l.Motivo,
		})
	}

	// d) Saldo total vem do contrato do usuário
	saldoTotal := 0
	if user != nil {
		saldoTotal = user.Contrato.SaldoBancoHorasMinutos
	}

	// e) Montar resposta
	resp := &DashboardResponse{
		SaldoTotalMinutos: saldoTotal,
		Historico:         historico,
	}
	return resp, nil
}
