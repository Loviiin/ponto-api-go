package model

type Empresa struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	NomeFantasia string `gorm:"not null" json:"nome_fantasia"`
	RazaoSocial  string `gorm:"not null" json:"razao_social"`
	CNPJ         string `gorm:"unique;not null" json:"cnpj"`

	// Configuração de Vinculação Google (strict ou flexible)
	GoogleLinkStrategy string `gorm:"default:'flexible'" json:"google_link_strategy,omitempty"`

	// Configurações globais de carga horária (padrão CLT)
	CargaHorariaPadraoDiariaMinutos  uint `json:"carga_horaria_padrao_diaria_minutos" gorm:"default:480"`   // 8h
	CargaHorariaPadraoSemanalMinutos uint `json:"carga_horaria_padrao_semanal_minutos" gorm:"default:2640"` // 44h
	ToleranciaAtrasoMinutos          uint `json:"tolerancia_atraso_minutos" gorm:"default:10"`
	ToleranciaExtraMinutos           uint `json:"tolerancia_extra_minutos" gorm:"default:10"`

	Localidades []Localidade `gorm:"foreignKey:EmpresaID" json:"localidades,omitempty"`
}
