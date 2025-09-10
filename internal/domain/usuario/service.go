package usuario

import (
	"errors"
	"github.com/Loviiin/ponto-api-go/internal/domain/cargo" // <-- 1. IMPORTAR O PACOTE DO CARGO
	"github.com/Loviiin/ponto-api-go/internal/domain/empresa"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/password"
	"gorm.io/gorm"
)

type UsuarioService interface {
	CriarUsuario(usuario *model.Usuario) error
	GetAll(empresaID uint) ([]model.Usuario, error)
	FindByID(id uint, empresaID uint) (*model.Usuario, error)
	Update(id uint, empresaID uint, dados map[string]interface{}) error
	Delete(id uint, empresaID uint) error
	FindAll() ([]model.Usuario, error)
}

var criptografaSenha = password.CriptografaSenha

type usuarioService struct {
	usuarioRepo UsuarioRepository
	cargoRepo   cargo.CargoRepository
	empresaRepo empresa.EmpresaRepository
}

func NewUsuarioService(repo UsuarioRepository, cargoRepo cargo.CargoRepository, empresaRepo empresa.EmpresaRepository) UsuarioService {
	return &usuarioService{
		usuarioRepo: repo,
		cargoRepo:   cargoRepo,
		empresaRepo: empresaRepo,
	}
}

func (s *usuarioService) GetAll(empresaID uint) ([]model.Usuario, error) {
	return s.usuarioRepo.GetAll(empresaID)
}

func (s *usuarioService) FindByID(id uint, empresaID uint) (*model.Usuario, error) {
	return s.usuarioRepo.FindByID(id, empresaID)
}

func (s *usuarioService) CriarUsuario(usuario *model.Usuario) error {
	_, err := s.usuarioRepo.FindByEmail(usuario.Email)
	if err == nil {
		return errors.New("e-mail já cadastrado")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	_, err = s.empresaRepo.FindByID(usuario.EmpresaID)
	if err != nil {
		return errors.New("a empresa especificada não existe")
	}
	_, err = s.cargoRepo.FindByID(usuario.CargoID, usuario.EmpresaID)
	if err != nil {
		return errors.New("o cargo especificado não existe ou não pertence a esta empresa")
	}

	senhaHash, err := criptografaSenha(usuario.Senha)
	if err != nil {
		return err
	}
	usuario.Senha = senhaHash
	return s.usuarioRepo.Save(usuario)
}

func (s *usuarioService) Update(id uint, empresaID uint, dados map[string]interface{}) error {
	_, err := s.usuarioRepo.FindByID(id, empresaID)
	if err != nil {
		return err
	}
	return s.usuarioRepo.Update(id, empresaID, dados)
}

func (s *usuarioService) Delete(id uint, empresaID uint) error {
	_, err := s.usuarioRepo.FindByID(id, empresaID)
	if err != nil {
		return err
	}
	return s.usuarioRepo.Delete(id, empresaID)
}

func (s *usuarioService) FindAll() ([]model.Usuario, error) {
	return s.usuarioRepo.FindAll()
}
