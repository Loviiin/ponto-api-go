package localidade

import (
	"github.com/Loviiin/ponto-api-go/internal/model"
	"gorm.io/gorm"
)

type Repository interface {
	Save(localidade *model.Localidade) error
	FindByID(id uint) (*model.Localidade, error)
	FindAllByEmpresaID(empresaID uint) ([]model.Localidade, error)
	Update(localidade *model.Localidade) error
	Delete(id uint, empresaID uint) error
}

type repository struct {
	Db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{Db: db}
}

func (r *repository) Save(localidade *model.Localidade) error {
	return r.Db.Create(localidade).Error
}

func (r *repository) FindByID(id uint) (*model.Localidade, error) {
	var localidade model.Localidade
	err := r.Db.First(&localidade, id).Error
	return &localidade, err
}

func (r *repository) FindAllByEmpresaID(empresaID uint) ([]model.Localidade, error) {
	var localidades []model.Localidade
	err := r.Db.Where("empresa_id = ?", empresaID).Find(&localidades).Error
	return localidades, err
}

func (r *repository) Update(localidade *model.Localidade) error {
	return r.Db.Save(localidade).Error
}

func (r *repository) Delete(id uint, empresaID uint) error {
	return r.Db.Unscoped().Delete(&model.Localidade{}, "id = ? AND empresa_id = ?", id, empresaID).Error
}