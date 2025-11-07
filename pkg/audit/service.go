package audit

import (
	"encoding/json"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/model"
	"gorm.io/gorm"
)

// Service define a interface para serviço de auditoria
type Service interface {
	LogAction(usuarioID, empresaID uint, acao, entidade string, entidadeID uint, dadosAntigos, dadosNovos interface{}, ip, userAgent string) error
	GetUserAuditLogs(usuarioID uint, limit int) ([]model.AuditLog, error)
	GetEntityAuditLogs(entidade string, entidadeID uint, limit int) ([]model.AuditLog, error)
}

type service struct {
	db *gorm.DB
}

// NewService cria uma nova instância do serviço de auditoria
func NewService(db *gorm.DB) Service {
	return &service{db: db}
}

// LogAction registra uma ação no log de auditoria
func (s *service) LogAction(usuarioID, empresaID uint, acao, entidade string, entidadeID uint, dadosAntigos, dadosNovos interface{}, ip, userAgent string) error {
	// Converter dados para JSON
	dadosAntigosJSON := ""
	if dadosAntigos != nil {
		bytes, err := json.Marshal(dadosAntigos)
		if err == nil {
			dadosAntigosJSON = string(bytes)
		}
	}

	dadosNovosJSON := ""
	if dadosNovos != nil {
		bytes, err := json.Marshal(dadosNovos)
		if err == nil {
			dadosNovosJSON = string(bytes)
		}
	}

	log := model.AuditLog{
		UsuarioID:    usuarioID,
		EmpresaID:    empresaID,
		Acao:         acao,
		Entidade:     entidade,
		EntidadeID:   entidadeID,
		DadosAntigos: dadosAntigosJSON,
		DadosNovos:   dadosNovosJSON,
		IP:           ip,
		UserAgent:    userAgent,
		CreatedAt:    time.Now(),
	}

	return s.db.Create(&log).Error
}

// GetUserAuditLogs obtém logs de auditoria de um usuário
func (s *service) GetUserAuditLogs(usuarioID uint, limit int) ([]model.AuditLog, error) {
	var logs []model.AuditLog
	err := s.db.Where("usuario_id = ?", usuarioID).
		Order("data_criacao DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// GetEntityAuditLogs obtém logs de auditoria de uma entidade específica
func (s *service) GetEntityAuditLogs(entidade string, entidadeID uint, limit int) ([]model.AuditLog, error) {
	var logs []model.AuditLog
	err := s.db.Where("entidade = ? AND entidade_id = ?", entidade, entidadeID).
		Order("data_criacao DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}
