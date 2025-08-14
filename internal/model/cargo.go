package model

type Cargo struct {
	ID         uint        `gorm:"primaryKey" json:"id"`
	Nome       string      `gorm:"not null" json:"nome"`
	EmpresaID  uint        `gorm:"not null" json:"empresa_id"`
	Permissoes []Permissao `gorm:"many2many:cargo_permissoes;" json:"permissoes,omitempty"`
}
