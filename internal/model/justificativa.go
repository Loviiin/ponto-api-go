// internal/model/justificativa.go
package model

import "time"

type Justificativa struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	DataOcorrencia time.Time `gorm:"not null" json:"data_ocorrencia"`
	Tipo           string    `gorm:"not null" json:"tipo"` // Ex: 'AJUSTE_PONTO', 'ATESTADO_MEDICO'
	Descricao      string    `gorm:"not null" json:"descricao"`
	Status         string    `gorm:"not null;default:'PENDENTE'" json:"status"` // PENDENTE, APROVADO, REPROVADO

	UsuarioID   uint  `gorm:"not null" json:"usuario_id"` // Quem está a justificar
	AprovadorID *uint `json:"aprovador_id"`               // Quem aprovou/reprovou (pode ser nulo)
	EmpresaID   uint  `gorm:"not null" json:"empresa_id"`

	Usuario   Usuario `json:"-"`
	Aprovador Usuario `gorm:"foreignKey:AprovadorID" json:"-"`
	Empresa   Empresa `json:"-"`
}
