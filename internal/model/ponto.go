package model

import "time"

type RegistroPonto struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"column:data_criacao" json:"data_criacao"`

	Timestamp time.Time `gorm:"not null" json:"timestamp"`

	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`

	Localizacao string `json:"localizacao"`
	Metodo      string `gorm:"not null" json:"metodo"`

	JustificativaID *uint          `json:"justificativa_id,omitempty"`
	Justificativa   *Justificativa `json:"justificativa,omitempty"` // Opcional, para carregar os dados se necessário
	UsuarioID       uint           `gorm:"not null" json:"usuario_id"`
	Usuario         Usuario        `json:"-"`
	EmpresaID       uint           `gorm:"not null" json:"empresa_id"`
	Empresa         Empresa        `json:"-"`
}
