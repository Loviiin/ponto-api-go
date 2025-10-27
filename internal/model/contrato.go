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

	// Carga horária personalizada (nullable - se NULL, usa padrão do Cargo)
	CargaHorariaDiariaMinutos  *uint `json:"carga_horaria_diaria_minutos,omitempty"`             // Ex: 480 (8h), 360 (6h), 240 (4h part-time)
	CargaHorariaSemanalMinutos *uint `json:"carga_horaria_semanal_minutos,omitempty"`            // Ex: 2640 (44h CLT), calculado se NULL
	DiasTrabalhadosSemana      *uint `json:"dias_trabalhados_semana,omitempty" gorm:"default:5"` // Segunda a sexta = 5

	// Tipo de contrato
	TipoContrato string `json:"tipo_contrato" gorm:"default:'CLT'"` // CLT, PJ, Estagiário, Part-time

	Empresa    Empresa    `json:"empresa,omitempty"`
	Localidade Localidade `json:"localidade,omitempty"`
	Cargo      Cargo      `json:"cargo,omitempty"`
}
