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

	// Relacionamentos - expõe dados básicos do usuário no JSON
	Usuario   *UsuarioBasico `gorm:"foreignKey:UsuarioID" json:"usuario,omitempty"`
	Aprovador *UsuarioBasico `gorm:"foreignKey:AprovadorID" json:"aprovador,omitempty"`
	Empresa   Empresa        `json:"-"`
}

// UsuarioBasico contém apenas campos essenciais para exibição em justificativas
type UsuarioBasico struct {
	ID    uint   `json:"id"`
	Nome  string `json:"nome"`
	Email string `json:"email"`
	CPF   string `json:"cpf"`
}
