package justificativa

import (
	"github.com/Loviiin/ponto-api-go/internal/model"
	"gorm.io/gorm"
)

type Repository interface {
	Create(justificativa *model.Justificativa) error
	FindByID(id uint, empresaID uint) (*model.Justificativa, error)
	FindByStatus(empresaID uint, status string) ([]model.Justificativa, error)
	Update(justificativa *model.Justificativa) error
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

func (r *repository) FindByID(id uint, empresaID uint) (*model.Justificativa, error) {
	var justificativa model.Justificativa
	err := r.Db.Where("id = ? AND empresa_id = ?", id, empresaID).First(&justificativa).Error
	if err != nil {
		return nil, err
	}
	return &justificativa, nil
}

func (r *repository) FindByStatus(empresaID uint, status string) ([]model.Justificativa, error) {
	var justificativas []model.Justificativa
	err := r.Db.Where("empresa_id = ? AND status = ?", empresaID, status).Find(&justificativas).Error
	return justificativas, err
}

func (r *repository) Update(justificativa *model.Justificativa) error {
	return r.Db.Save(justificativa).Error
}
