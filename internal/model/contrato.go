package model

import "time"

type Contrato struct {
    ID                     uint       `gorm:"primaryKey" json:"id"`
    UsuarioID              uint       `gorm:"unique;not null" json:"usuario_id"`
    EmpresaID              uint       `gorm:"not null" json:"empresa_id"`
    LocalidadeID           uint       `gorm:"not null" json:"localidade_id"`
    CargoID                uint       `gorm:"not null" json:"cargo_id"`
    Salario                float64    `json:"salario"`
    DataAdmissao           time.Time  `json:"data_admissao"`
    DataDemissao           *time.Time `json:"data_demissao,omitempty"`
    SaldoBancoHorasMinutos int        `json:"saldo_banco_horas_minutos"`
}