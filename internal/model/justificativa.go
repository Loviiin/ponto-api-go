// internal/model/justificativa.go
package model

import "time"

type Justificativa struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	DataOcorrencia time.Time  `gorm:"not null" json:"data_ocorrencia"`
	Tipo           string     `gorm:"not null" json:"tipo"` // PONTO_FALTANTE | CORRECAO_PONTO
	Descricao      string     `gorm:"not null" json:"descricao"`
	Status         string     `gorm:"not null;default:'PENDENTE'" json:"status"`
	NovoHorario    *time.Time `json:"novo_horario,omitempty"` // Horário corrigido solicitado (para CORRECAO_PONTO)

	UsuarioID   uint  `gorm:"not null" json:"usuario_id"`
	AprovadorID *uint `json:"aprovador_id"`
	EmpresaID   uint  `gorm:"not null" json:"empresa_id"`
	PontoID     *uint `json:"ponto_id,omitempty"` // ID do ponto a ser corrigido (apenas para CORRECAO_PONTO)

	ObservacaoAprovador string `json:"observacao_aprovador,omitempty"` // Para feedback ou motivo da reprovação

	// Relacionamentos - expõe dados básicos do usuário no JSON
	Usuario   *UsuarioBasico `gorm:"foreignKey:UsuarioID" json:"usuario,omitempty"`
	Aprovador *UsuarioBasico `gorm:"foreignKey:AprovadorID" json:"aprovador,omitempty"`
	Empresa   Empresa        `json:"-"`
	Ponto     *RegistroPonto `gorm:"foreignKey:PontoID" json:"ponto,omitempty"` // Ponto referenciado (se tipo CORRECAO_PONTO)
}

// UsuarioBasico contém apenas campos essenciais para exibição em justificativas
type UsuarioBasico struct {
	ID    uint   `json:"id"`
	Nome  string `json:"nome"`
	Email string `json:"email"`
	CPF   string `json:"cpf"`
}
