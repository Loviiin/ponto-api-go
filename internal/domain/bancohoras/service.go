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
	GetDashboardForUsuario(usuarioID uint, empresaID uint, dataInicio *time.Time, dataFim *time.Time) (*DashboardResponse, error)
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

	pontos, err := s.pontoRepo.FindPontosByUserIDAndDate(ctx, user.ID, dia)
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
		// Invalida cache do histórico
		cacheKeyHistorico := fmt.Sprintf("bancohoras:historico:empresa:%d:usuario:%d", empresaID, usuarioID)
		_ = s.cache.Delete(ctx, cacheKeyHistorico)
	}

	// Retorna o contrato atualizado
	usuarioAtual.Contrato.SaldoBancoHorasMinutos = novoSaldoTotal
	return &usuarioAtual.Contrato, nil
}

// GetDashboardForUsuario retorna o saldo total atual e o histórico de alterações do banco de horas
// do usuário informado já ordenado por data desc.
// Usa cache de 2 horas para o histórico (logs) que muda pouco, e recalcula o saldo atual sempre
// Permite filtrar o histórico por dataInicio e dataFim (opcional)
func (s *bancoHorasService) GetDashboardForUsuario(usuarioID uint, empresaID uint, dataInicio *time.Time, dataFim *time.Time) (*DashboardResponse, error) {
	ctx := context.Background()
	cacheKeyHistorico := fmt.Sprintf("bancohoras:historico:empresa:%d:usuario:%d", empresaID, usuarioID)

	var logs []model.LogBancoHoras
	var err error

	// Tenta buscar o histórico do cache
	cacheHit := false
	if s.cache != nil {
		cachedValue, errCache := s.cache.Get(ctx, cacheKeyHistorico)
		if errCache == nil && cachedValue != "" {
			if errUnmarshal := json.Unmarshal([]byte(cachedValue), &logs); errUnmarshal == nil {
				cacheHit = true
				log.Printf("[cache] HIT histórico banco horas - usuário %d empresa %d", usuarioID, empresaID)
			}
		}
	}

	// Se não encontrou no cache, busca do banco
	if !cacheHit {
		log.Printf("[cache] MISS histórico banco horas - usuário %d empresa %d", usuarioID, empresaID)
		logs, err = s.logRepo.GetAllByUsuarioAndEmpresa(usuarioID, empresaID)
		if err != nil {
			return nil, err
		}

		// Armazena no cache por 2 horas
		if s.cache != nil {
			logsJSON, _ := json.Marshal(logs)
			_ = s.cache.Set(ctx, cacheKeyHistorico, string(logsJSON), 2*time.Hour)
		}
	}

	// a) Buscar o usuário (com contrato)
	user, err := s.usuarioRepo.FindByID(ctx, usuarioID, empresaID)
	if err != nil {
		return nil, err
	}

	// Mesmo com a query ordenada, garantimos ordenação desc por segurança
	sort.SliceStable(logs, func(i, j int) bool { return logs[i].Data.After(logs[j].Data) })

	// c) Mapear para DTO e aplicar filtros de data se fornecidos
	historico := make([]HistoricoDia, 0, len(logs))
	for _, l := range logs {
		// Aplica filtro de data_inicio se fornecido
		if dataInicio != nil && l.Data.Before(*dataInicio) {
			continue
		}
		// Aplica filtro de data_fim se fornecido
		if dataFim != nil && l.Data.After(*dataFim) {
			continue
		}

		historico = append(historico, HistoricoDia{
			Data:                   l.Data.Format("2006-01-02"),
			ValorAlteradoMinutos:   l.ValorAlteradoMinutos,
			SaldoResultanteMinutos: l.SaldoNovoMinutos,
			Motivo:                 l.Motivo,
		})
	}

	// d) Saldo total vem do contrato do usuário (saldo até ontem)
	saldoTotal := 0
	if user != nil {
		saldoTotal = user.Contrato.SaldoBancoHorasMinutos
	}

	// e) Adiciona o saldo do dia atual (se houver pontos hoje)
	loc, _ := time.LoadLocation("America/Sao_Paulo")
	if loc == nil {
		loc = time.Local
	}
	hoje := time.Now().In(loc)
	saldoHoje, err := s.CalcularSaldoParaUsuario(usuarioID, empresaID, hoje)
	if err == nil {
		// Se conseguiu calcular o saldo de hoje, adiciona ao total
		saldoTotal += saldoHoje

		// Adiciona uma entrada no histórico para o dia atual (saldo provisório)
		// Verifica se já não existe um log para hoje
		hojeFormatado := hoje.Format("2006-01-02")
		jaExisteHoje := false
		for _, h := range historico {
			if h.Data == hojeFormatado {
				jaExisteHoje = true
				break
			}
		}

		if !jaExisteHoje && saldoHoje != 0 {
			// Insere no início (mais recente)
			historicoComHoje := make([]HistoricoDia, 0, len(historico)+1)
			historicoComHoje = append(historicoComHoje, HistoricoDia{
				Data:                   hojeFormatado,
				ValorAlteradoMinutos:   saldoHoje,
				SaldoResultanteMinutos: saldoTotal,
				Motivo:                 "Saldo provisório do dia atual (em tempo real)",
			})
			historicoComHoje = append(historicoComHoje, historico...)
			historico = historicoComHoje
		}
	} else {
		// Se não conseguiu (ex: número ímpar de marcações), ignora e só mostra até ontem
		log.Printf("Não foi possível calcular saldo de hoje para usuário %d: %v", usuarioID, err)
	}

	// f) Montar resposta
	resp := &DashboardResponse{
		SaldoTotalMinutos: saldoTotal,
		Historico:         historico,
	}

	// NÃO armazena no cache pois o saldo inclui o dia atual que muda constantemente
	// O cache continua sendo útil para os logs (histórico) mas não para o saldo total

	return resp, nil
}

// GetSaldoAtualUsuario retorna o saldo atual do banco de horas de um usuário
// Inclui o saldo acumulado até ontem + saldo do dia atual em tempo real
func (s *bancoHorasService) GetSaldoAtualUsuario(usuarioID uint, empresaID uint) (int, error) {
	ctx := context.Background()

	// Se não encontrou no cache ou cache não disponível, busca do banco
	user, err := s.usuarioRepo.FindByID(ctx, usuarioID, empresaID)
	if err != nil {
		return 0, err
	}

	if user.Contrato.ID == 0 {
		return 0, errors.New("usuário não possui contrato ativo")
	}

	// Saldo até ontem (armazenado no contrato)
	saldoTotal := user.Contrato.SaldoBancoHorasMinutos

	// Adiciona o saldo do dia atual (se houver pontos hoje)
	loc, _ := time.LoadLocation("America/Sao_Paulo")
	if loc == nil {
		loc = time.Local
	}
	hoje := time.Now().In(loc)
	saldoHoje, err := s.CalcularSaldoParaUsuario(usuarioID, empresaID, hoje)
	if err == nil {
		// Se conseguiu calcular o saldo de hoje, adiciona ao total
		saldoTotal += saldoHoje
	}
	// Se não conseguiu (ex: número ímpar de marcações), ignora e só mostra até ontem

	return saldoTotal, nil
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
		// Invalida cache do histórico
		cacheKeyHistorico := fmt.Sprintf("bancohoras:historico:empresa:%d:usuario:%d", empresaID, usuarioID)
		_ = s.cache.Delete(ctx, cacheKeyHistorico)
	}
}
