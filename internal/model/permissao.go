package model

type Permissao struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Nome      string `gorm:"unique;not null" json:"nome"`
	Descricao string `json:"descricao"`
}
