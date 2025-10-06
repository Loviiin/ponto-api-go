package model

import "time"

type Usuario struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Nome      string    `gorm:"not null"   json:"nome"`
    CPF       string    `gorm:"unique;not null" json:"cpf"`
    Email     string    `gorm:"unique;not null" json:"email"`
    Senha     string    `gorm:"not null" json:"-"`
    CreatedAt time.Time `gorm:"column:data_criacao" json:"data_criacao"`
    UpdatedAt time.Time `gorm:"column:data_atualizacao" json:"data_atualizacao"` 
    Contrato  Contrato  `gorm:"foreignKey:UsuarioID" json:"contrato,omitempty"`
}
