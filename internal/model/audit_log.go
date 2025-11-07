package model

import (
	"time"

	"gorm.io/gorm"
)

// AuditLog registra todas as alterações importantes no sistema
type AuditLog struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UsuarioID   uint           `gorm:"not null;index" json:"usuario_id"`
	EmpresaID   uint           `gorm:"not null;index" json:"empresa_id"`
	Acao        string         `gorm:"not null" json:"acao"` // ex: "UPDATE_PROFILE", "CHANGE_PASSWORD", "UPDATE_CPF"
	Entidade    string         `gorm:"not null" json:"entidade"` // ex: "usuario", "ponto", "banco_horas"
	EntidadeID  uint           `json:"entidade_id,omitempty"`
	DadosAntigos string        `gorm:"type:jsonb" json:"dados_antigos,omitempty"` // JSON com dados antes da alteração
	DadosNovos   string        `gorm:"type:jsonb" json:"dados_novos,omitempty"`   // JSON com dados após alteração
	IP          string         `json:"ip,omitempty"`
	UserAgent   string         `json:"user_agent,omitempty"`
	CreatedAt   time.Time      `gorm:"column:data_criacao" json:"data_criacao"`
	DeletedAt   gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`
	
	// Relações
	Usuario Usuario `gorm:"foreignKey:UsuarioID" json:"usuario,omitempty"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
