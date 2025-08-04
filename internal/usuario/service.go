package usuario

import (
	"errors"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/password"
	"gorm.io/gorm"
)

type UsuarioService interface {
	CriarUsuario(usuario *model.Usuario) error
	GetAll() ([]model.Usuario, error)
	FindByID(id uint) (*model.Usuario, error)
	Update(id uint, dados map[string]interface{}) error
}

type usuarioService struct {
	usuarioRepo UsuarioRepository
}

func NewUsuarioService(repo UsuarioRepository) UsuarioService {
	return &usuarioService{
		usuarioRepo: repo,
	}
}

func (s *usuarioService) GetAll() ([]model.Usuario, error) {
	return s.usuarioRepo.GetAll()
}

func (s *usuarioService) FindByID(id uint) (*model.Usuario, error) {
	return s.usuarioRepo.FindByID(id)
}

func (s *usuarioService) CriarUsuario(usuario *model.Usuario) error {
	_, err := s.usuarioRepo.FindByEmail(usuario.Email)
	if err == nil {
		// Se err é nulo, o usuário foi encontrado, então o e-mail já existe.
		return errors.New("e-mail já cadastrado")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err // Retorna o erro original do banco.
	}
	SenhaHash, err := password.CriptografaSenha(usuario.Senha)
	if err != nil {
		return err
	}
	usuario.Senha = SenhaHash
	return s.usuarioRepo.Save(usuario)
}

func (s *usuarioService) Update(id uint, dados map[string]interface{}) error {
	_, err := s.usuarioRepo.FindByID(id)
	if err != nil {
		return err
	}
	return s.usuarioRepo.Update(id, dados)
}
