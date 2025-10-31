package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type Usuario struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Nome      string         `gorm:"not null" json:"nome"`
	CPF       string         `gorm:"unique;not null" json:"cpf"`
	Email     string         `gorm:"unique;not null" json:"email"`
	Senha     string         `gorm:"not null" json:"-"`
	Telefone  string         `json:"telefone,omitempty"`
	Avatar    string         `json:"avatar,omitempty"`
	CreatedAt time.Time      `gorm:"column:data_criacao" json:"data_criacao"`
	UpdatedAt time.Time      `gorm:"column:data_atualizacao" json:"data_atualizacao"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`
	Contrato  Contrato       `gorm:"foreignKey:UsuarioID" json:"contrato,omitempty"`
}

// MarshalJSON customiza a serialização JSON para converter timestamps para o fuso horário do Brasil
func (u Usuario) MarshalJSON() ([]byte, error) {
	// Carrega o fuso horário do Brasil
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.Local // fallback para timezone local do servidor
	}

	// Cria um tipo auxiliar para evitar recursão infinita
	type Alias Usuario
	return json.Marshal(&struct {
		*Alias
		CreatedAt time.Time `json:"data_criacao"`
		UpdatedAt time.Time `json:"data_atualizacao"`
	}{
		Alias:     (*Alias)(&u),
		CreatedAt: u.CreatedAt.In(loc),
		UpdatedAt: u.UpdatedAt.In(loc),
	})
}
