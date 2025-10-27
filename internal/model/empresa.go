package model

type Empresa struct {
    ID                        uint         `gorm:"primaryKey" json:"id"`
    NomeFantasia              string       `gorm:"not null" json:"nome_fantasia"`
    RazaoSocial               string       `gorm:"not null" json:"razao_social"`
    CNPJ                      string       `gorm:"unique;not null" json:"cnpj"`
    PresencialRestritoAoRaio  bool         `gorm:"default:false" json:"presencial_restrito_ao_raio"`
    Localidades               []Localidade `gorm:"foreignKey:EmpresaID" json:"localidades,omitempty"`
}