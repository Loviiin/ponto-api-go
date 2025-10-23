package usuario

import (
	"context"
	"errors"
	"fmt"

	"github.com/Loviiin/ponto-api-go/internal/domain/cargo"
	"github.com/Loviiin/ponto-api-go/internal/domain/contrato"
	"github.com/Loviiin/ponto-api-go/internal/domain/empresa"
	"github.com/Loviiin/ponto-api-go/internal/domain/localidade"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/password"
	"gorm.io/gorm"
)

type UsuarioService interface {
	// Agora exige o id do requisitante para validação de hierarquia
	CriarUsuarioEContrato(usuario *model.Usuario, contrato *model.Contrato, idRequisitante uint) error
	GetAll(empresaID uint) ([]model.Usuario, error)
	GetAllPaginated(empresaID uint, page int, limit int) ([]model.Usuario, int64, error)
	FindByID(id uint, empresaID uint) (*model.Usuario, error)
	Update(id uint, empresaID uint, dados map[string]interface{}) error
	Delete(id uint, empresaID uint) error
	FindAll() ([]model.Usuario, error)
}

var criptografaSenha = password.CriptografaSenha

type usuarioService struct {
	Db             *gorm.DB
	usuarioRepo    UsuarioRepository
	cargoRepo      cargo.CargoRepository
	empresaRepo    empresa.EmpresaRepository
	contratoRepo   contrato.ContratoRepository
	localidadeRepo localidade.Repository
}

func NewUsuarioService(db *gorm.DB, repo UsuarioRepository, cargoRepo cargo.CargoRepository, empresaRepo empresa.EmpresaRepository, contratoRepo contrato.ContratoRepository, localidadeRepo localidade.Repository) UsuarioService {
	return &usuarioService{
		Db:             db,
		usuarioRepo:    repo,
		cargoRepo:      cargoRepo,
		empresaRepo:    empresaRepo,
		contratoRepo:   contratoRepo,
		localidadeRepo: localidadeRepo,
	}
}

func (s *usuarioService) GetAll(empresaID uint) ([]model.Usuario, error) {
	return s.usuarioRepo.GetAll(empresaID)
}

func (s *usuarioService) GetAllPaginated(empresaID uint, page int, limit int) ([]model.Usuario, int64, error) {
	return s.usuarioRepo.GetAllPaginated(empresaID, page, limit)
}

func (s *usuarioService) FindByID(id uint, empresaID uint) (*model.Usuario, error) {
	return s.usuarioRepo.FindByID(context.Background(), id, empresaID)
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

	// Validação 1: Email único
	_, err := s.usuarioRepo.FindByEmail(usuario.Email)
	if err == nil {
		tx.Rollback()
		return errors.New("email já cadastrado")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return err
	}

	// Validação 2: CPF único
	_, err = s.usuarioRepo.FindByCPF(usuario.CPF)
	if err == nil {
		tx.Rollback()
		return errors.New("cpf já cadastrado")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return err
	}

	// Validação 3: Cargo existe e pertence à empresa
	cargoAlvo, err := s.cargoRepo.FindByID(context.Background(), contrato.CargoID, contrato.EmpresaID)
	if err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("cargo não encontrado ou não pertence a esta empresa")
		}
		return errors.New("erro ao buscar cargo")
	}

	// Validação 4: Localidade existe e pertence à empresa
	var localidadeComEmpresa model.Localidade
	if err := s.Db.Where("id = ? AND empresa_id = ?", contrato.LocalidadeID, contrato.EmpresaID).First(&localidadeComEmpresa).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("localidade não encontrada ou não pertence a esta empresa")
		}
		return errors.New("erro ao validar localidade")
	}

	// Validação 5: Buscar o cargo do requisitante para regras de hierarquia
	requisitante, err := s.usuarioRepo.FindByID(context.Background(), idRequisitante, contrato.EmpresaID)
	if err != nil {
		tx.Rollback()
		return errors.New("usuário requisitante não encontrado para validar permissão")
	}
	cargoRequisitante := requisitante.Contrato.Cargo

	// Validação 6: Regra de hierarquia - requisitante não pode criar cargo superior ao seu
	if cargoRequisitante.NivelHierarquia < cargoAlvo.NivelHierarquia {
		tx.Rollback()
		return errors.New("acesso negado: você não pode atribuir um cargo com nível hierárquico superior ao seu")
	}

	// Validação 7: Salário dentro da faixa permitida
	if cargoAlvo.SalarioMaximo > 0 && (contrato.Salario < cargoAlvo.SalarioMinimo || contrato.Salario > cargoAlvo.SalarioMaximo) {
		tx.Rollback()
		return fmt.Errorf("o salário R$%.2f está fora da faixa permitida (R$%.2f - R$%.2f) para o cargo %s",
			contrato.Salario, cargoAlvo.SalarioMinimo, cargoAlvo.SalarioMaximo, cargoAlvo.Nome)
	}

	// Criptografar senha
	senhaHash, err := criptografaSenha(usuario.Senha)
	if err != nil {
		tx.Rollback()
		return errors.New("erro ao criptografar senha")
	}
	usuario.Senha = senhaHash

	// Criar usuário
	if err := s.usuarioRepo.WithTransaction(tx).Save(usuario); err != nil {
		tx.Rollback()
		return errors.New("erro ao criar usuário")
	}

	// Criar contrato
	contrato.UsuarioID = usuario.ID
	if err := s.contratoRepo.WithTransaction(tx).Save(contrato); err != nil {
		tx.Rollback()
		return errors.New("erro ao criar contrato")
	}

	// Commit da transação
	if err := tx.Commit().Error; err != nil {
		return errors.New("erro ao finalizar criação do usuário")
	}

	// Buscar usuário criado com todos os relacionamentos para retornar
	usuarioCriado, err := s.usuarioRepo.FindByID(context.Background(), usuario.ID, contrato.EmpresaID)
	if err != nil {
		// Usuário foi criado, mas não conseguimos buscar os detalhes completos
		return nil
	}

	// Copiar dados completos para o ponteiro original
	*usuario = *usuarioCriado

	return nil
}

func (s *usuarioService) Update(id uint, empresaID uint, dados map[string]interface{}) error {
	// Verificar se usuário existe
	usuarioAtual, err := s.usuarioRepo.FindByID(context.Background(), id, empresaID)
	if err != nil {
		return err
	}

	// Se email está sendo atualizado, verificar se já está em uso
	if novoEmail, ok := dados["email"].(string); ok && novoEmail != "" {
		// Buscar usuário com este email
		usuarioComEmail, err := s.usuarioRepo.FindByEmail(novoEmail)
		if err == nil {
			// Email encontrado, verificar se é de outro usuário
			if usuarioComEmail.ID != usuarioAtual.ID {
				return errors.New("email já está em uso por outro usuário")
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			// Erro diferente de "não encontrado"
			return err
		}
	}

	return s.usuarioRepo.Update(id, dados)
}

func (s *usuarioService) Delete(id uint, empresaID uint) error {
	_, err := s.usuarioRepo.FindByID(context.Background(), id, empresaID)
	if err != nil {
		return err
	}
	return s.usuarioRepo.Delete(id)
}

func (s *usuarioService) FindAll() ([]model.Usuario, error) {
	return s.usuarioRepo.FindAll()
}
