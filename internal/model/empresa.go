package model

type Empresa struct {
	ID                 uint    `gorm:"primaryKey" json:"id"`
	Nome               string  `gorm:"not null" json:"nome"`
	SedeLatitude       float64 `json:"sedeLatitude"`
	SedeLongitude      float64 `json:"sedeLongitude"`
	RaioGeofenceMetros float64 `json:"raioGeofenceMetros"`
}
