package permissao

import (
	"github.com/Loviiin/ponto-api-go/internal/model"
	"gorm.io/gorm"
)

type Repository interface {
	Create(permissao *model.Permissao) error
	FindAll() ([]model.Permissao, error)
}

type repository struct {
	Db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{Db: db}
}

func (r *repository) Create(permissao *model.Permissao) error {
	return r.Db.Create(permissao).Error
}

func (r *repository) FindAll() ([]model.Permissao, error) {
	var permissoes []model.Permissao
	err := r.Db.Find(&permissoes).Error
	return permissoes, err
}
