package repository

import (
	"github.com/Loviiin/ponto-api-go/internal/model"
	"gorm.io/gorm"
)

type UsuarioRepository struct {
	Db *gorm.DB
}

func (r *UsuarioRepository) Save(usuario *model.Usuario) error {
	return r.Db.Create(usuario).Error
}

func (r *UsuarioRepository) FindByEmail(email string) (*model.Usuario, error) {
	var usuario model.Usuario
	err := r.Db.Where("email = ?", email).First(&usuario).Error
	return &usuario, err
}

func (r *UsuarioRepository) FindByID(id uint) (*model.Usuario, error) {
	var usuario model.Usuario
	err := r.Db.First(&usuario, id).Error
	return &usuario, err
}
