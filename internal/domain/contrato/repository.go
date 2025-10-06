// Em: internal/domain/contrato/repository.go
package contrato

import (
	"github.com/Loviiin/ponto-api-go/internal/model"
	"gorm.io/gorm"
)

type ContratoRepository interface {
	Save(contrato *model.Contrato) error
	WithTransaction(tx *gorm.DB) ContratoRepository
}

type contratoRepository struct {
	Db *gorm.DB
}

func NewContratoRepository(db *gorm.DB) ContratoRepository {
	return &contratoRepository{Db: db}
}

func (r *contratoRepository) Save(contrato *model.Contrato) error {
	return r.Db.Create(contrato).Error
}

func (r *contratoRepository) WithTransaction(tx *gorm.DB) ContratoRepository {
	return &contratoRepository{Db: tx}
}