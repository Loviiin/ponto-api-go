package auth

import (
	"errors"
	"fmt"

	"github.com/Loviiin/ponto-api-go/internal/config"
	"github.com/Loviiin/ponto-api-go/internal/domain/cargo"
	"github.com/Loviiin/ponto-api-go/internal/domain/contrato" // NOVO IMPORT
	"github.com/Loviiin/ponto-api-go/internal/domain/empresa"
	"github.com/Loviiin/ponto-api-go/internal/domain/localidade" // NOVO IMPORT
	"github.com/Loviiin/ponto-api-go/internal/domain/usuario"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/geolocation"
	"github.com/Loviiin/ponto-api-go/pkg/jwt"
	"github.com/Loviiin/ponto-api-go/pkg/password"
	"gorm.io/gorm"
)

type AuthService interface {
	Authenticate(email string, password string) (string, error)
	SignUp(
		empresaReq *model.Empresa,
		localidadeParcial *model.Localidade,
		usuarioReq *model.Usuario,
		dadosContrato *model.Contrato,
	) (*model.Usuario, string, error)
}

type authService struct {
	usuarioRepo    usuario.UsuarioRepository
	empresaRepo    empresa.EmpresaRepository
	cargoRepo      cargo.CargoRepository
	contratoRepo   contrato.ContratoRepository
	localidadeRepo localidade.Repository
	geolocationSvc geolocation.Service
	jwtService     *jwt.JWTService
	db             *gorm.DB
}

func NewAuthService(
	usuarioRepo usuario.UsuarioRepository,
	empresaRepo empresa.EmpresaRepository,
	cargoRepo cargo.CargoRepository,
	contratoRepo contrato.ContratoRepository,
	localidadeRepo localidade.Repository,
	geolocationSvc geolocation.Service,
	jwtService *jwt.JWTService,
	db *gorm.DB,
) AuthService {
	return &authService{
		usuarioRepo:    usuarioRepo,
		empresaRepo:    empresaRepo,
		cargoRepo:      cargoRepo,
		contratoRepo:   contratoRepo,
		localidadeRepo: localidadeRepo,
		geolocationSvc: geolocationSvc, // NOVA ATRIBUIÇÃO
		jwtService:     jwtService,
		db:             db,
	}
}

func (s *authService) Authenticate(email string, passwordStr string) (string, error) {
	// A busca por email já carrega o Contrato graças às nossas correções anteriores
	usuari, err := s.usuarioRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("credenciais inválidas")
		}
		return "", err
	}

	if usuari.Contrato.ID == 0 {
		return "", errors.New("utilizador não possui um contrato ativo")
	}

	if !password.VerificaHashSenha(passwordStr, usuari.Senha) {
		return "", errors.New("credenciais inválidas")
	}

	// O token agora é gerado com o EmpresaID do Contrato
	token, err := s.jwtService.GenerateToken(usuari.ID, usuari.Contrato.EmpresaID)
	// --- FIM DA CORREÇÃO ---
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *authService) SignUp(
	empresaReq *model.Empresa,
	localidadeParcial *model.Localidade,
	usuarioReq *model.Usuario,
	dadosContrato *model.Contrato,
) (*model.Usuario, string, error) {
	_, err := s.usuarioRepo.FindByEmail(usuarioReq.Email)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", errors.New("e-mail já cadastrado")
	}

	// Adicionar verificação de CNPJ aqui se necessário

	var novoUsuario *model.Usuario
	var token string

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Criar a Empresa
		if err := s.empresaRepo.WithTransaction(tx).CreateEmpresa(empresaReq); err != nil {
			return fmt.Errorf("falha ao criar empresa: %w", err)
		}

		// 2. Obter dados completos da localidade a partir do CEP e criar
		localidadeCompleta, err := s.geolocationSvc.GetLocationFromCEP(localidadeParcial.CEP)
		if err != nil {
			return fmt.Errorf("falha ao obter dados do CEP: %w", err)
		}
		localidadeCompleta.Nome = localidadeParcial.Nome
		localidadeCompleta.RaioGeofenceMetros = localidadeParcial.RaioGeofenceMetros
		localidadeCompleta.EmpresaID = empresaReq.ID

		if err := s.localidadeRepo.WithTransaction(tx).Save(localidadeCompleta); err != nil {
			return fmt.Errorf("falha ao criar localidade: %w", err)
		}

		// 3. Configurar Cargos e Permissões e obter os cargos padrão
		permissoes := config.SeedPermissions(tx)
		donoCargo, _, _ := config.SetupDefaultRolesAndPermissions(tx, empresaReq.ID, permissoes)

		// 4. Preparar os dados do utilizador antes de o criar
		// Nota: Empresa e Cargo agora pertencem ao Contrato; o usuário não possui mais esses campos.

		senhaHash, err := password.CriptografaSenha(usuarioReq.Senha)
		if err != nil {
			return err
		}
		usuarioReq.Senha = senhaHash
		if err := s.usuarioRepo.WithTransaction(tx).Save(usuarioReq); err != nil {
			return fmt.Errorf("falha ao criar utilizador: %w", err)
		}

		// 5. Criar o Contrato
		novoContrato := &model.Contrato{
			UsuarioID:    usuarioReq.ID,
			EmpresaID:    empresaReq.ID,
			LocalidadeID: localidadeCompleta.ID,
			CargoID:      donoCargo.ID,
			Salario:      dadosContrato.Salario,
			DataAdmissao: dadosContrato.DataAdmissao,
		}
		if err := s.contratoRepo.WithTransaction(tx).Save(novoContrato); err != nil {
			return fmt.Errorf("falha ao criar contrato: %w", err)
		}

		// 6. Gerar Token
		token, err = s.jwtService.GenerateToken(usuarioReq.ID, empresaReq.ID)
		if err != nil {
			return err
		}
		novoUsuario = usuarioReq
		novoUsuario.Contrato = *novoContrato

		return nil
	})

	return novoUsuario, token, err
}
