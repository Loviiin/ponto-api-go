package logbancohoras

import (
	"github.com/Loviiin/ponto-api-go/internal/model"
	"gorm.io/gorm"
)

type Repository interface {
	Create(log *model.LogBancoHoras) error
	GetAllByUsuarioAndEmpresa(usuarioID uint, empresaID uint) ([]model.LogBancoHoras, error)
	WithTransaction(tx *gorm.DB) Repository
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

func (r *repository) GetAllByUsuarioAndEmpresa(usuarioID uint, empresaID uint) ([]model.LogBancoHoras, error) {
	var logs []model.LogBancoHoras
	err := r.Db.Where("usuario_id = ? AND empresa_id = ?", usuarioID, empresaID).
		Order("data desc").
		Find(&logs).Error
	return logs, err
}

func (r *repository) WithTransaction(tx *gorm.DB) Repository {
	return &repository{Db: tx}
}