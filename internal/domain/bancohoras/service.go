package bancohoras

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/domain/logbancohoras"
	"github.com/Loviiin/ponto-api-go/pkg/cache"
	"gorm.io/gorm"

	"github.com/Loviiin/ponto-api-go/internal/domain/ponto"
	"github.com/Loviiin/ponto-api-go/internal/domain/usuario"
	"github.com/Loviiin/ponto-api-go/internal/model"
)

type BancoHorasService interface {
	CalcularSaldoParaUsuario(usuarioID uint, empresaID uint, dia time.Time) (int, error)
	FecharDiaParaUsuario(usuarioID uint, empresaID uint, dia time.Time) (*model.Contrato, error)
	GetDashboardForUsuario(usuarioID uint, empresaID uint) (*DashboardResponse, error)
	GetSaldoAtualUsuario(usuarioID uint, empresaID uint) (int, error)
	InvalidarCacheDia(usuarioID uint, empresaID uint, dia time.Time)
	InvalidarCacheUsuario(usuarioID uint, empresaID uint)
}

type bancoHorasService struct {
	pontoRepo   ponto.RegistroPontoRepository
	usuarioRepo usuario.UsuarioRepository
	logRepo     logbancohoras.Repository
	db          *gorm.DB
	cache       cache.Service
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
		cache:       nil, // Mantém compatibilidade com código existente
	}
}

func NewBancoHorasServiceWithCache(
	pontoRepo ponto.RegistroPontoRepository,
	userRepo usuario.UsuarioRepository,
	logRepo logbancohoras.Repository,
	db *gorm.DB,
	cacheService cache.Service,
) BancoHorasService {
	return &bancoHorasService{
		pontoRepo:   pontoRepo,
		usuarioRepo: userRepo,
		logRepo:     logRepo,
		db:          db,
		cache:       cacheService,
	}
}

// CalcularSaldoParaUsuario calcula o saldo de horas de um dia específico para um usuário
// Cache de 30 minutos: após o dia terminar, os pontos não mudam mais (exceto ajustes manuais)
func (s *bancoHorasService) CalcularSaldoParaUsuario(usuarioID uint, empresaID uint, dia time.Time) (int, error) {
	ctx := context.Background()
	diaFormatado := dia.Format("2006-01-02")
	cacheKey := fmt.Sprintf("bancohoras:saldodia:empresa:%d:usuario:%d:dia:%s", empresaID, usuarioID, diaFormatado)

	// Tenta buscar do cache se disponível
	if s.cache != nil {
		cachedValue, err := s.cache.Get(ctx, cacheKey)
		if err == nil && cachedValue != "" {
			var saldo int
			if err := json.Unmarshal([]byte(cachedValue), &saldo); err == nil {
				return saldo, nil
			}
		}
	}

	user, err := s.usuarioRepo.FindByID(ctx, usuarioID, empresaID)
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

	// Armazena no cache por 30 minutos (dias passados não mudam, mas pode haver ajustes)
	if s.cache != nil {
		saldoJSON, _ := json.Marshal(doDia)
		_ = s.cache.Set(ctx, cacheKey, string(saldoJSON), 30*time.Minute)
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

	// Usa o helper para obter a carga horária (respeita hierarquia Contrato > Cargo > Padrão)
	cargaHorariaDiaria := float64(cargoDoUsuario.CargaHorariaDiariaMinutos)
	saldo := totalTrabalhadoEmMinutos - cargaHorariaDiaria

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

	// Invalida todos os caches relacionados ao banco de horas do usuário após atualização
	if s.cache != nil {
		ctx := context.Background()
		// Invalida cache do saldo simples
		cacheKeySaldo := fmt.Sprintf("bancohoras:saldo:empresa:%d:usuario:%d", empresaID, usuarioID)
		_ = s.cache.Delete(ctx, cacheKeySaldo)
		// Invalida cache do dashboard completo (com histórico)
		cacheKeyDashboard := fmt.Sprintf("bancohoras:dashboard:empresa:%d:usuario:%d", empresaID, usuarioID)
		_ = s.cache.Delete(ctx, cacheKeyDashboard)
	}

	// Retorna o contrato atualizado
	usuarioAtual.Contrato.SaldoBancoHorasMinutos = novoSaldoTotal
	return &usuarioAtual.Contrato, nil
}

// GetDashboardForUsuario retorna o saldo total atual e o histórico de alterações do banco de horas
// do usuário informado já ordenado por data desc.
// Cache de 12 horas pois o dashboard é pesado (busca todos os logs) e só muda 1x/dia (scheduler 1h AM)
func (s *bancoHorasService) GetDashboardForUsuario(usuarioID uint, empresaID uint) (*DashboardResponse, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("bancohoras:dashboard:empresa:%d:usuario:%d", empresaID, usuarioID)

	// Tenta buscar do cache se disponível
	if s.cache != nil {
		cachedValue, err := s.cache.Get(ctx, cacheKey)
		if err == nil && cachedValue != "" {
			var dashboard DashboardResponse
			if err := json.Unmarshal([]byte(cachedValue), &dashboard); err == nil {
				log.Printf("[cache] HIT dashboard banco horas - usuário %d empresa %d", usuarioID, empresaID)
				return &dashboard, nil
			}
		}
	}

	log.Printf("[cache] MISS dashboard banco horas - usuário %d empresa %d", usuarioID, empresaID)

	// a) Buscar o usuário (com contrato)
	user, err := s.usuarioRepo.FindByID(ctx, usuarioID, empresaID)
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

	// Armazena no cache por 12 horas (dados mudam 1x/dia no scheduler às 1h AM)
	// Cache invalidado automaticamente em ajustes manuais via InvalidarCacheUsuario()
	if s.cache != nil {
		dashboardJSON, _ := json.Marshal(resp)
		_ = s.cache.Set(ctx, cacheKey, string(dashboardJSON), 12*time.Hour)
	}

	return resp, nil
}

// GetSaldoAtualUsuario retorna o saldo atual do banco de horas de um usuário
// com cache de 5 minutos para melhorar performance
func (s *bancoHorasService) GetSaldoAtualUsuario(usuarioID uint, empresaID uint) (int, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("bancohoras:saldo:empresa:%d:usuario:%d", empresaID, usuarioID)

	// Tenta buscar do cache se disponível
	if s.cache != nil {
		cachedValue, err := s.cache.Get(ctx, cacheKey)
		if err == nil && cachedValue != "" {
			var saldo int
			if err := json.Unmarshal([]byte(cachedValue), &saldo); err == nil {
				return saldo, nil
			}
		}
	}

	// Se não encontrou no cache ou cache não disponível, busca do banco
	user, err := s.usuarioRepo.FindByID(ctx, usuarioID, empresaID)
	if err != nil {
		return 0, err
	}

	if user.Contrato.ID == 0 {
		return 0, errors.New("usuário não possui contrato ativo")
	}

	saldo := user.Contrato.SaldoBancoHorasMinutos

	// Armazena no cache por 5 minutos (saldo muda apenas no fechamento diário)
	if s.cache != nil {
		saldoJSON, _ := json.Marshal(saldo)
		_ = s.cache.Set(ctx, cacheKey, string(saldoJSON), 5*time.Minute)
	}

	return saldo, nil
}

// InvalidarCacheDia invalida o cache de saldo de um dia específico
// Usado quando há ajustes manuais de ponto
func (s *bancoHorasService) InvalidarCacheDia(usuarioID uint, empresaID uint, dia time.Time) {
	if s.cache != nil {
		ctx := context.Background()
		diaFormatado := dia.Format("2006-01-02")
		cacheKey := fmt.Sprintf("bancohoras:saldodia:empresa:%d:usuario:%d:dia:%s", empresaID, usuarioID, diaFormatado)
		_ = s.cache.Delete(ctx, cacheKey)
	}
}

// InvalidarCacheUsuario invalida todos os caches relacionados a um usuário
// Útil para operações que afetam múltiplos aspectos do banco de horas
func (s *bancoHorasService) InvalidarCacheUsuario(usuarioID uint, empresaID uint) {
	if s.cache != nil {
		ctx := context.Background()
		// Invalida cache do saldo atual
		cacheKeySaldo := fmt.Sprintf("bancohoras:saldo:empresa:%d:usuario:%d", empresaID, usuarioID)
		_ = s.cache.Delete(ctx, cacheKeySaldo)
		// Invalida cache do dashboard completo
		cacheKeyDashboard := fmt.Sprintf("bancohoras:dashboard:empresa:%d:usuario:%d", empresaID, usuarioID)
		_ = s.cache.Delete(ctx, cacheKeyDashboard)
	}
}
