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
	Update(id uint, empresaID uint, dados map[string]interface{}) error
	Delete(id uint, empresaID uint) error
	FindAll() ([]model.Usuario, error)
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
	err := r.Db.Where("email = ?", email).First(&usuario).Error
	return &usuario, err
}

func (r *usuarioRepository) FindByID(id uint, empresaID uint) (*model.Usuario, error) {
	var usuario model.Usuario
	err := r.Db.Where("id = ? AND empresa_id = ?", id, empresaID).Preload("Cargo.Permissoes").First(&usuario).Error
	return &usuario, err
}

func (r *usuarioRepository) GetAll(empresaID uint) ([]model.Usuario, error) {
	var usuarios []model.Usuario
	err := r.Db.Where("empresa_id = ?", empresaID).Order("id asc").Find(&usuarios).Error
	return usuarios, err
}

func (r *usuarioRepository) Update(id uint, empresaID uint, dados map[string]interface{}) error {
	err := r.Db.Model(&model.Usuario{}).Where("id = ? AND empresa_id = ?", id, empresaID).Updates(dados).Error
	return err
}

func (r *usuarioRepository) Delete(id uint, empresaID uint) error {
	err := r.Db.Delete(&model.Usuario{}, "id = ? AND empresa_id = ?", id, empresaID).Error
	return err
}

func (r *usuarioRepository) FindAll() ([]model.Usuario, error) {
	var usuarios []model.Usuario
	err := r.Db.Find(&usuarios).Error
	return usuarios, err
}
