package repository

import (
	"github.com/Loviiin/ponto-api-go/internal/model"
	"gorm.io/gorm"
)

type UsuarioRepository interface {
	Save(usuario *model.Usuario) error
	FindByEmail(email string) (*model.Usuario, error)
	FindByID(id uint) (*model.Usuario, error)
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

func (r *usuarioRepository) FindByID(id uint) (*model.Usuario, error) {
	var usuario model.Usuario
	err := r.Db.First(&usuario, id).Error
	return &usuario, err
}
