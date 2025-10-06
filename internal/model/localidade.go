package model

type Localidade struct {
    ID                 uint    `gorm:"primaryKey" json:"id"`
    Nome               string  `gorm:"not null" json:"nome"` // Ex: "Matriz São Paulo", "Filial Rio"
    EmpresaID          uint    `gorm:"not null" json:"empresa_id"`
    CEP                string  `json:"cep"`
    Logradouro         string  `json:"logradouro"`
    Numero             string  `json:"numero"`
    Bairro             string  `json:"bairro"`
    Cidade             string  `json:"cidade"`
    Estado             string  `json:"estado"`
    Latitude           float64 `json:"latitude"`
    Longitude          float64 `json:"longitude"`
    RaioGeofenceMetros float64 `json:"raio_geofence_metros"`
}