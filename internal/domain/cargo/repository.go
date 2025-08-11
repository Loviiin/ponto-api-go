package cargo

import (
	"github.com/Loviiin/ponto-api-go/internal/model"
	"gorm.io/gorm"
)

// CargoRepository define a interface para as operações de dados de Cargo.
type CargoRepository interface {
	Create(cargo *model.Cargo) error
	FindByID(id uint, empresaID uint) (*model.Cargo, error)
	GetAllByEmpresaID(empresaID uint) ([]model.Cargo, error)
	Update(id uint, empresaID uint, dados map[string]interface{}) error
	Delete(id uint, empresaID uint) error
}

type cargoRepository struct {
	Db *gorm.DB
}

// NewCargoRepository cria uma nova instância do repositório de cargos.
func NewCargoRepository(db *gorm.DB) CargoRepository {
	return &cargoRepository{Db: db}
}

func (r *cargoRepository) Create(cargo *model.Cargo) error {
	return r.Db.Create(cargo).Error
}

func (r *cargoRepository) FindByID(id uint, empresaID uint) (*model.Cargo, error) {
	var cargo model.Cargo
	err := r.Db.Where("id = ? AND empresa_id = ?", id, empresaID).First(&cargo).Error
	return &cargo, err
}

func (r *cargoRepository) GetAllByEmpresaID(empresaID uint) ([]model.Cargo, error) {
	var cargos []model.Cargo
	err := r.Db.Where("empresa_id = ?", empresaID).Order("id asc").Find(&cargos).Error
	return cargos, err
}

func (r *cargoRepository) Update(id uint, empresaID uint, dados map[string]interface{}) error {
	return r.Db.Model(&model.Cargo{}).Where("id = ? AND empresa_id = ?", id, empresaID).Updates(dados).Error
}

func (r *cargoRepository) Delete(id uint, empresaID uint) error {
	return r.Db.Unscoped().Delete(&model.Cargo{}, "id = ? AND empresa_id = ?", id, empresaID).Error
}
