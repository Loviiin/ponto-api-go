package profile

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/cache"
	"gorm.io/gorm"
)

// Repository define a interface para operações de perfil
type Repository interface {
	// Dados do Usuário
	GetUserByID(userID uint) (*model.Usuario, error)
	UpdateUser(user *model.Usuario) error

	// Estatísticas de Ponto
	GetPontoStatsByUserAndMonth(userID uint, month, year int) (*PontoMonthStats, error)
	GetRecentPontos(userID uint, limit int) ([]model.RegistroPonto, error)
	GetTotalPontosCount(userID uint) (int64, error)

	// Banco de Horas
	GetLatestBancoHoras(userID, empresaID uint) (*model.LogBancoHoras, error)

	// Justificativas
	GetJustificativasCount(userID, empresaID uint) (int64, error)
	GetJustificativasPendentesCount(userID, empresaID uint) (int64, error)

	// Localidades
	GetLocalidadesByEmpresa(empresaID uint) ([]model.Localidade, error)

	// Permissões
	GetPermissoesByUserID(userID uint) ([]model.Permissao, error)

	// Calendário
	GetPontosByMonth(userID uint, month, year int) ([]model.RegistroPonto, error)
}

type repository struct {
	db    *gorm.DB
	cache cache.Service
}

// NewRepository cria uma nova instância do repository
func NewRepository(db *gorm.DB, cache cache.Service) Repository {
	return &repository{
		db:    db,
		cache: cache,
	}
}

// PontoMonthStats representa estatísticas agregadas de ponto para um mês
type PontoMonthStats struct {
	TotalRegistros        int64
	DiasTrabalados        int
	TotalAtrasos          int
	TotalSaidasAdiantadas int
	MinutosTrabalhados    int
}

// GetUserByID busca um usuário por ID com seus relacionamentos
func (r *repository) GetUserByID(userID uint) (*model.Usuario, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("profile:user:%d", userID)

	// Tentar buscar do cache
	if cached, err := r.cache.Get(ctx, cacheKey); err == nil && cached != "" {
		var user model.Usuario
		if err := json.Unmarshal([]byte(cached), &user); err == nil {
			return &user, nil
		}
	}

	var user model.Usuario
	err := r.db.
		Preload("Contrato").
		Preload("Contrato.Cargo").
		Preload("Contrato.Empresa").
		First(&user, userID).Error

	if err != nil {
		return nil, err
	}

	// Cachear por 5 minutos
	if data, err := json.Marshal(user); err == nil {
		r.cache.Set(ctx, cacheKey, string(data), 5*time.Minute)
	}

	return &user, nil
}

// UpdateUser atualiza os dados do usuário e invalida o cache
func (r *repository) UpdateUser(user *model.Usuario) error {
	err := r.db.Save(user).Error
	if err == nil {
		// Invalidar cache
		ctx := context.Background()
		cacheKey := fmt.Sprintf("profile:user:%d", user.ID)
		r.cache.Delete(ctx, cacheKey)
	}
	return err
}

// GetPontoStatsByUserAndMonth retorna estatísticas de ponto para um mês específico
func (r *repository) GetPontoStatsByUserAndMonth(userID uint, month, year int) (*PontoMonthStats, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("profile:stats:%d:%d:%d", userID, month, year)

	// Tentar buscar do cache
	if cached, err := r.cache.Get(ctx, cacheKey); err == nil && cached != "" {
		var stats PontoMonthStats
		if err := json.Unmarshal([]byte(cached), &stats); err == nil {
			return &stats, nil
		}
	}

	var stats PontoMonthStats

	// Período do mês
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	// Total de registros
	err := r.db.Model(&model.RegistroPonto{}).
		Where("usuario_id = ? AND timestamp >= ? AND timestamp <= ?", userID, startDate, endDate).
		Count(&stats.TotalRegistros).Error

	if err != nil {
		return nil, err
	}

	// Buscar todos os pontos do mês para calcular estatísticas
	var pontos []model.RegistroPonto
	err = r.db.Where("usuario_id = ? AND timestamp >= ? AND timestamp <= ?", userID, startDate, endDate).
		Order("timestamp ASC").
		Find(&pontos).Error

	if err != nil {
		return nil, err
	}

	// Calcular estatísticas
	diasMap := make(map[string]bool)

	for _, ponto := range pontos {
		// Contar dias únicos trabalhados (agrupa por data)
		dataKey := ponto.Timestamp.Format("2006-01-02")
		diasMap[dataKey] = true

		// Nota: Como cada RegistroPonto é uma batida individual (entrada OU saída),
		// o cálculo de horas trabalhadas precisa agrupar pares de pontos por dia
		// Isso será implementado no service layer com lógica mais complexa
	}

	stats.DiasTrabalados = len(diasMap)

	// Cachear por 10 minutos (stats mudam com frequência)
	if data, err := json.Marshal(stats); err == nil {
		r.cache.Set(ctx, cacheKey, string(data), 10*time.Minute)
	}

	return &stats, nil
}

// GetRecentPontos retorna os últimos registros de ponto
func (r *repository) GetRecentPontos(userID uint, limit int) ([]model.RegistroPonto, error) {
	var pontos []model.RegistroPonto
	err := r.db.
		Where("usuario_id = ?", userID).
		Order("timestamp DESC").
		Limit(limit).
		Find(&pontos).Error

	return pontos, err
}

// GetTotalPontosCount retorna o total de registros de ponto
func (r *repository) GetTotalPontosCount(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.RegistroPonto{}).
		Where("usuario_id = ?", userID).
		Count(&count).Error

	return count, err
}

// GetLatestBancoHoras retorna o último registro de banco de horas
func (r *repository) GetLatestBancoHoras(userID, empresaID uint) (*model.LogBancoHoras, error) {
	var log model.LogBancoHoras
	err := r.db.
		Where("usuario_id = ? AND empresa_id = ?", userID, empresaID).
		Order("data_atualizacao DESC").
		First(&log).Error

	if err != nil {
		return nil, err
	}

	return &log, nil
}

// GetJustificativasCount retorna o total de justificativas
func (r *repository) GetJustificativasCount(userID, empresaID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Justificativa{}).
		Where("usuario_id = ? AND empresa_id = ?", userID, empresaID).
		Count(&count).Error

	return count, err
}

// GetJustificativasPendentesCount retorna o total de justificativas pendentes
func (r *repository) GetJustificativasPendentesCount(userID, empresaID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Justificativa{}).
		Where("usuario_id = ? AND empresa_id = ? AND status = ?", userID, empresaID, "pendente").
		Count(&count).Error

	return count, err
}

// GetLocalidadesByEmpresa retorna todas as localidades de uma empresa
func (r *repository) GetLocalidadesByEmpresa(empresaID uint) ([]model.Localidade, error) {
	var localidades []model.Localidade
	err := r.db.
		Where("empresa_id = ?", empresaID).
		Find(&localidades).Error

	return localidades, err
}

// GetPontosByMonth retorna todos os pontos de um mês específico
func (r *repository) GetPontosByMonth(userID uint, month, year int) ([]model.RegistroPonto, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	var pontos []model.RegistroPonto
	err := r.db.
		Where("usuario_id = ? AND timestamp >= ? AND timestamp <= ?", userID, startDate, endDate).
		Order("timestamp ASC").
		Find(&pontos).Error

	return pontos, err
}

// GetPermissoesByUserID retorna todas as permissões de um usuário
func (r *repository) GetPermissoesByUserID(userID uint) ([]model.Permissao, error) {
	var user model.Usuario
	err := r.db.
		Preload("Contrato.Cargo.Permissoes").
		First(&user, userID).Error

	if err != nil {
		return nil, err
	}

	if user.Contrato.ID == 0 || user.Contrato.Cargo.ID == 0 {
		return []model.Permissao{}, nil
	}

	return user.Contrato.Cargo.Permissoes, nil
}
