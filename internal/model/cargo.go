package model

type Cargo struct {
	ID                        uint        `gorm:"primaryKey" json:"id"`
	Nome                      string      `gorm:"not null" json:"nome"`
	EmpresaID                 uint        `gorm:"not null" json:"empresa_id"`
	Permissoes                []Permissao `gorm:"many2many:cargo_permissoes;" json:"permissoes,omitempty"`
	CargaHorariaDiariaMinutos uint        `json:"carga_horaria_diaria_minutos"`
	EntradaEsperadaMinutos    uint        `json:"entrada_esperada_minutos"`
	SaidaEsperadaMinutos      uint        `json:"saida_esperada_minutos"`
	MinutosAlmocoEsperado     uint        `json:"minutos_almoco_esperado"`
}
