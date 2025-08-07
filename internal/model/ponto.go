package model

import "time"

type RegistroPonto struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"column:data_criacao" json:"data_criacao"`

	Timestamp time.Time `gorm:"not null" json:"timestamp"`

	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`

	UsuarioID uint    `gorm:"not null" json:"usuario_id"`
	Usuario   Usuario `json:"-"`

	EmpresaID uint `gorm:"not null" json:"empresa_id"`
}
