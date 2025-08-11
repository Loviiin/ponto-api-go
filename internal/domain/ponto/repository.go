package ponto

import (
	"github.com/Loviiin/ponto-api-go/internal/model"
	"gorm.io/gorm"
)

type RegistroPontoRepository interface {
	SavePonto(ponto *model.RegistroPonto) error
}

type pontoRepository struct {
	Db *gorm.DB
}

func NewPontoRepository(db *gorm.DB) RegistroPontoRepository {
	return &pontoRepository{Db: db}
}

func (r *pontoRepository) SavePonto(ponto *model.RegistroPonto) error {
	return r.Db.Create(ponto).Error
}
