package empresa

import (
	"github.com/Loviiin/ponto-api-go/internal/model"
	"gorm.io/gorm"
)

type EmpresaRepository interface {
	FindByID(id uint) (*model.Empresa, error)
}

type empresaRepository struct {
	Db *gorm.DB
}

func NewEmpresaRepository(db *gorm.DB) EmpresaRepository {
	return &empresaRepository{Db: db}
}

func (r *empresaRepository) FindByID(id uint) (*model.Empresa, error) {
	var empresa model.Empresa
	err := r.Db.Where("id = ?", id).First(&empresa).Error
	return &empresa, err
}
