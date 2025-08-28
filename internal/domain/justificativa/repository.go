package justificativa

import (
	"github.com/Loviiin/ponto-api-go/internal/model"
	"gorm.io/gorm"
)

type Repository interface {
	Create(justificativa *model.Justificativa) error
	WithTransaction(tx *gorm.DB) Repository
}

type repository struct {
	Db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{Db: db}
}

func (r *repository) Create(justificativa *model.Justificativa) error {
	return r.Db.Create(justificativa).Error
}

func (r *repository) WithTransaction(tx *gorm.DB) Repository {
	return &repository{Db: tx}
}
