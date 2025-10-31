package model

import (
	"encoding/json"
	"time"
)

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

// MarshalJSON customiza a serialização JSON para converter timestamps para o fuso horário do Brasil
func (c Contrato) MarshalJSON() ([]byte, error) {
	// Carrega o fuso horário do Brasil
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.Local // fallback para timezone local do servidor
	}

	// Cria um tipo auxiliar para evitar recursão infinita
	type Alias Contrato
	aux := &struct {
		*Alias
		DataAdmissao time.Time  `json:"data_admissao"`
		DataDemissao *time.Time `json:"data_demissao,omitempty"`
	}{
		Alias:        (*Alias)(&c),
		DataAdmissao: c.DataAdmissao.In(loc),
	}

	// Converte DataDemissao se não for nil
	if c.DataDemissao != nil {
		converted := c.DataDemissao.In(loc)
		aux.DataDemissao = &converted
	}

	return json.Marshal(aux)
}
