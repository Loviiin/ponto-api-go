package empresa

import (
	"github.com/Loviiin/ponto-api-go/internal/model"
	"gorm.io/gorm"
)

type EmpresaRepository interface {
	FindByID(id uint) (*model.Empresa, error)
	CreateEmpresa(empresa *model.Empresa) error
	GetAllEmpresas() ([]model.Empresa, error)
	GetEmpresaByID(idempresa uint) (*model.Empresa, error)
	UpdateEmpresa(idempresa uint, dados map[string]interface{}) error
	DeleteEmpresa(idempresa uint) error
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

func (r *empresaRepository) CreateEmpresa(empresa *model.Empresa) error {
	return r.Db.Create(empresa).Error
}

func (r *empresaRepository) GetAllEmpresas() ([]model.Empresa, error) {
	var empresas []model.Empresa
	err := r.Db.Order("id asc").Find(&empresas).Error
	return empresas, err
}

func (r *empresaRepository) GetEmpresaByID(idempresa uint) (*model.Empresa, error) {
	var empresa model.Empresa
	err := r.Db.Where("id = ?", idempresa).First(&empresa).Error
	return &empresa, err
}

func (r *empresaRepository) UpdateEmpresa(idempresa uint, dados map[string]interface{}) error {
	err := r.Db.Model(&model.Empresa{}).Where("id = ?", idempresa).Updates(dados).Error
	return err
}

// DeleteEmpresa remove uma empresa do banco de dados.
func (r *empresaRepository) DeleteEmpresa(idempresa uint) error {
	// A função Unscoped() garante uma exclusão permanente (hard delete).
	// Sem ela, o GORM faria um soft delete se o modelo tivesse um campo gorm.DeletedAt.
	return r.Db.Unscoped().Delete(&model.Empresa{}, idempresa).Error
}
