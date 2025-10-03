package logbancohoras

import (
	"github.com/Loviiin/ponto-api-go/internal/model"
	"gorm.io/gorm"
)

type Repository interface {
	Create(log *model.LogBancoHoras) error
	WithTransaction(tx *gorm.DB) Repository // ADICIONAR ESTA LINHA
}

type repository struct {
	Db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{Db: db}
}

func (r *repository) Create(log *model.LogBancoHoras) error {
	return r.Db.Create(log).Error
}

func (r *repository) WithTransaction(tx *gorm.DB) Repository {
	return &repository{Db: tx}
}