package service

import (
	"errors"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/internal/repository"
	"github.com/Loviiin/ponto-api-go/pkg/password"
	"gorm.io/gorm"
)

type UsuarioService struct {
	UsuarioRepository *repository.UsuarioRepository
}

func (s *UsuarioService) CriarUsuario(usuario *model.Usuario) error {
	_, err := s.UsuarioRepository.FindByEmail(usuario.Email)
	if err == nil {
		return errors.New("Tem email cadastrado ja chefe")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	SenhaHash, err := password.CriptografaSenha(usuario.Senha)
	if err != nil {
		return err
	}
	usuario.Senha = SenhaHash
	return s.UsuarioRepository.Save(usuario)
}
