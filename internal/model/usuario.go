package model

import "time"

type Usuario struct {
	ID                     uint      `gorm:"primaryKey" json:"id"`
	Nome                   string    `gorm:"not null"   json:"nome"`
	Email                  string    `gorm:"unique;not null" json:"email"`
	Senha                  string    `gorm:"not null" json:"-"`
	EmpresaID              uint      `gorm:"not null" json:"empresa_id"`
	Empresa                Empresa   `json:"-"`
	CargoID                uint      `gorm:"not null" json:"cargo_id"`
	Cargo                  Cargo     `json:"-"`
	CreatedAt              time.Time `gorm:"column:data_criacao" json:"data_criacao"`
	UpdatedAt              time.Time `gorm:"column:data_atualizacao" json:"data_atualizacao"`
	SaldoBancoHorasMinutos int       `json:"saldo_banco_horas_minutos"`
}
