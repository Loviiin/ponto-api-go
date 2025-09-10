// internal/model/justificativa.go
package model

import "time"

type Justificativa struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	DataOcorrencia time.Time `gorm:"not null" json:"data_ocorrencia"`
	Tipo           string    `gorm:"not null" json:"tipo"`
	Descricao      string    `gorm:"not null" json:"descricao"`
	Status         string    `gorm:"not null;default:'PENDENTE'" json:"status"`

	UsuarioID   uint  `gorm:"not null" json:"usuario_id"`
	AprovadorID *uint `json:"aprovador_id"`
	EmpresaID   uint  `gorm:"not null" json:"empresa_id"`

	ObservacaoAprovador string `json:"observacao_aprovador,omitempty"` // Para feedback ou motivo da reprovação

	Usuario   Usuario `json:"-"`
	Aprovador Usuario `gorm:"foreignKey:AprovadorID" json:"-"`
	Empresa   Empresa `json:"-"`
}
