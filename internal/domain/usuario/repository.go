package usuario

import (
	"github.com/Loviiin/ponto-api-go/internal/model"
	"gorm.io/gorm"
)

type UsuarioRepository interface {
	Save(usuario *model.Usuario) error
	FindByEmail(email string) (*model.Usuario, error)
	FindByID(id uint, empresaID uint) (*model.Usuario, error)
	GetAll(empresaID uint) ([]model.Usuario, error)
	Update(id uint, dados map[string]interface{}) error
	Delete(id uint) error
	FindAll() ([]model.Usuario, error)
	WithTransaction(tx *gorm.DB) UsuarioRepository
}

type usuarioRepository struct {
	Db *gorm.DB
}

func NewUsuarioRepository(db *gorm.DB) UsuarioRepository {
	return &usuarioRepository{Db: db}
}

func (r *usuarioRepository) Save(usuario *model.Usuario) error {
	return r.Db.Create(usuario).Error
}

func (r *usuarioRepository) FindByEmail(email string) (*model.Usuario, error) {
	var usuario model.Usuario
	err := r.Db.Where("email = ?", email).
		Preload("Contrato.Cargo.Permissoes").
		First(&usuario).Error
	return &usuario, err
}

func (r *usuarioRepository) FindByID(id uint, empresaID uint) (*model.Usuario, error) {
	var usuario model.Usuario
	err := r.Db.Joins("JOIN contratos on contratos.usuario_id = usuarios.id").
		Where("usuarios.id = ? AND contratos.empresa_id = ?", id, empresaID).
		Preload("Contrato.Cargo.Permissoes").
		Preload("Contrato.Empresa").
        Preload("Contrato.Localidade").
        Preload("Contrato.Cargo.Permissoes").
		First(&usuario).Error
	return &usuario, err
}

func (r *usuarioRepository) GetAll(empresaID uint) ([]model.Usuario, error) {
	var usuarios []model.Usuario
	err := r.Db.Joins("JOIN contratos ON contratos.usuario_id = usuarios.id").
		Where("contratos.empresa_id = ?", empresaID).
		Order("usuarios.id asc").
		Preload("Contrato.Cargo").
		Find(&usuarios).Error
	return usuarios, err
}

func (r *usuarioRepository) Update(id uint, dados map[string]interface{}) error {
	err := r.Db.Model(&model.Usuario{}).Where("id = ?", id).Updates(dados).Error
	return err
}

func (r *usuarioRepository) Delete(id uint) error {
	if err := r.Db.Where("usuario_id = ?", id).Delete(&model.Contrato{}).Error; err != nil {
		return err
	}
	return r.Db.Unscoped().Delete(&model.Usuario{}, id).Error
}

func (r *usuarioRepository) FindAll() ([]model.Usuario, error) {
	var usuarios []model.Usuario
	err := r.Db.Preload("Contrato").Find(&usuarios).Error
	return usuarios, err
}

func (r *usuarioRepository) WithTransaction(tx *gorm.DB) UsuarioRepository {
    return &usuarioRepository{Db: tx}
}
