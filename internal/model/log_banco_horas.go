package model

import (
	"encoding/json"
	"time"
)

type LogBancoHoras struct {
	ID                   uint      `gorm:"primaryKey" json:"id"`
	UsuarioID            uint      `gorm:"not null;index" json:"usuario_id"`
	AutorID              *uint     `json:"autor_id"`
	EmpresaID            uint      `gorm:"not null;index" json:"empresa_id"`
	Data                 time.Time `gorm:"not null" json:"data"`
	ValorAlteradoMinutos int       `gorm:"not null" json:"valor_alterado_minutos"`
	SaldoAnteriorMinutos int       `gorm:"not null" json:"saldo_anterior_minutos"`
	SaldoNovoMinutos     int       `gorm:"not null" json:"saldo_novo_minutos"`
	Motivo               string    `gorm:"not null" json:"motivo"`

	Usuario Usuario `json:"-"`
	Autor   Usuario `gorm:"foreignKey:AutorID" json:"-"`
	Empresa Empresa `json:"-"`
}

// MarshalJSON customiza a serialização JSON para converter timestamps para o fuso horário do Brasil
func (l LogBancoHoras) MarshalJSON() ([]byte, error) {
	// Carrega o fuso horário do Brasil
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.Local // fallback para timezone local do servidor
	}

	// Cria um tipo auxiliar para evitar recursão infinita
	type Alias LogBancoHoras
	return json.Marshal(&struct {
		*Alias
		Data time.Time `json:"data"`
	}{
		Alias: (*Alias)(&l),
		Data:  l.Data.In(loc),
	})
}
