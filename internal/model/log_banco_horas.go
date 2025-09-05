package model

import "time"

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
