package model

import (
	"encoding/json"
	"time"
)

type RegistroPonto struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"column:data_criacao" json:"data_criacao"`

	Timestamp time.Time `gorm:"not null" json:"timestamp"`

	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`

	Localizacao string `json:"localizacao"`
	Metodo      string `json:"metodo"`
	Status      string `gorm:"default:'CONFIRMADO'" json:"status"` // CONFIRMADO | PENDENTE_APROVACAO | APROVADO | REPROVADO

	JustificativaID *uint `json:"justificativa_id,omitempty"`
	// Justificativa é carregada via Preload quando necessário, não precisa estar aqui para evitar referência circular
	UsuarioID uint    `gorm:"not null" json:"usuario_id"`
	Usuario   Usuario `json:"-"`
	EmpresaID uint    `gorm:"not null" json:"empresa_id"`
	Empresa   Empresa `json:"-"`
}

// MarshalJSON customiza a serialização JSON para converter timestamps para o fuso horário do Brasil
func (r RegistroPonto) MarshalJSON() ([]byte, error) {
	// Carrega o fuso horário do Brasil
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.Local // fallback para timezone local do servidor
	}

	// Cria um tipo auxiliar para evitar recursão infinita
	type Alias RegistroPonto
	return json.Marshal(&struct {
		*Alias
		Timestamp time.Time `json:"timestamp"`
		CreatedAt time.Time `json:"data_criacao"`
	}{
		Alias:     (*Alias)(&r),
		Timestamp: r.Timestamp.In(loc),
		CreatedAt: r.CreatedAt.In(loc),
	})
}
