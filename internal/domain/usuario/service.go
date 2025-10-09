package usuario

import (
	"errors"
	"fmt"

	"github.com/Loviiin/ponto-api-go/internal/domain/cargo" // <-- 1. IMPORTAR O PACOTE DO CARGO
	"github.com/Loviiin/ponto-api-go/internal/domain/contrato"
	"github.com/Loviiin/ponto-api-go/internal/domain/empresa"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/password"
	"gorm.io/gorm"
)

type UsuarioService interface {
	// Agora exige o id do requisitante para validação de hierarquia
	CriarUsuarioEContrato(usuario *model.Usuario, contrato *model.Contrato, idRequisitante uint) error
	GetAll(empresaID uint) ([]model.Usuario, error)
	FindByID(id uint, empresaID uint) (*model.Usuario, error)
	Update(id uint, empresaID uint, dados map[string]interface{}) error
	Delete(id uint, empresaID uint) error
	FindAll() ([]model.Usuario, error)
}

var criptografaSenha = password.CriptografaSenha

type usuarioService struct {
	Db           *gorm.DB
	usuarioRepo  UsuarioRepository
	cargoRepo    cargo.CargoRepository
	empresaRepo  empresa.EmpresaRepository
	contratoRepo contrato.ContratoRepository
}

func NewUsuarioService(db *gorm.DB, repo UsuarioRepository, cargoRepo cargo.CargoRepository, empresaRepo empresa.EmpresaRepository, contratoRepo contrato.ContratoRepository) UsuarioService {
	return &usuarioService{
		Db:           db,
		usuarioRepo:  repo,
		cargoRepo:    cargoRepo,
		empresaRepo:  empresaRepo,
		contratoRepo: contratoRepo,
	}
}

func (s *usuarioService) GetAll(empresaID uint) ([]model.Usuario, error) {
	return s.usuarioRepo.GetAll(empresaID)
}

func (s *usuarioService) FindByID(id uint, empresaID uint) (*model.Usuario, error) {
	return s.usuarioRepo.FindByID(id, empresaID)
}

func (s *usuarioService) CriarUsuarioEContrato(usuario *model.Usuario, contrato *model.Contrato, idRequisitante uint) error {
	tx := s.Db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	_, err := s.usuarioRepo.FindByEmail(usuario.Email)
	if err == nil {
		tx.Rollback()
		return errors.New("e-mail já cadastrado")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return err
	}

	//TODO Adicionar outras validações igual email

	cargoAlvo, err := s.cargoRepo.FindByID(contrato.CargoID, contrato.EmpresaID)
	if err != nil {
		tx.Rollback()
		return errors.New("o cargo especificado não existe ou não pertence a esta empresa")
	}

	// Buscar o cargo do requisitante para regras de hierarquia
	requisitante, err := s.usuarioRepo.FindByID(idRequisitante, contrato.EmpresaID)
	if err != nil {
		tx.Rollback()
		return errors.New("usuário requisitante não encontrado para validar permissão")
	}
	cargoRequisitante := requisitante.Contrato.Cargo

	// Regra genérica de hierarquia: requisitante não pode criar cargo com nível superior ao seu
	if cargoRequisitante.NivelHierarquia < cargoAlvo.NivelHierarquia {
		tx.Rollback()
		return errors.New("acesso negado: você não pode atribuir um cargo com nível hierárquico superior ao seu")
	}

	if cargoAlvo.SalarioMaximo > 0 && (contrato.Salario < cargoAlvo.SalarioMinimo || contrato.Salario > cargoAlvo.SalarioMaximo) {
		tx.Rollback()
		return fmt.Errorf("o salário R$%.2f está fora da faixa permitida (R$%.2f - R$%.2f) para o cargo %s",
			contrato.Salario, cargoAlvo.SalarioMinimo, cargoAlvo.SalarioMaximo, cargoAlvo.Nome)
	}

	senhaHash, err := criptografaSenha(usuario.Senha)
	if err != nil {
		tx.Rollback()
		return err
	}
	usuario.Senha = senhaHash

	if err := s.usuarioRepo.WithTransaction(tx).Save(usuario); err != nil {
		tx.Rollback()
		return err
	}

	contrato.UsuarioID = usuario.ID

	if err := s.contratoRepo.WithTransaction(tx).Save(contrato); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *usuarioService) Update(id uint, empresaID uint, dados map[string]interface{}) error {
	_, err := s.usuarioRepo.FindByID(id, empresaID)
	if err != nil {
		return err
	}
	return s.usuarioRepo.Update(id, dados)
}

func (s *usuarioService) Delete(id uint, empresaID uint) error {
	_, err := s.usuarioRepo.FindByID(id, empresaID)
	if err != nil {
		return err
	}
	return s.usuarioRepo.Delete(id)
}

func (s *usuarioService) FindAll() ([]model.Usuario, error) {
	return s.usuarioRepo.FindAll()
}
