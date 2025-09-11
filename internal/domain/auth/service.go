package auth

import (
	"errors"

	"github.com/Loviiin/ponto-api-go/internal/config"
	"github.com/Loviiin/ponto-api-go/internal/domain/cargo"
	"github.com/Loviiin/ponto-api-go/internal/domain/empresa"
	"github.com/Loviiin/ponto-api-go/internal/domain/usuario"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/jwt"
	"github.com/Loviiin/ponto-api-go/pkg/password"
	"gorm.io/gorm"
)

type AuthService interface {
	Authenticate(email string, password string) (string, error)
	SignUp(empresaReq *model.Empresa, usuarioReq *model.Usuario) (*model.Usuario, string, error)
}

type authService struct {
	usuarioRepo usuario.UsuarioRepository
	empresaRepo empresa.EmpresaRepository
	cargoRepo   cargo.CargoRepository
	jwtService  *jwt.JWTService
	db          *gorm.DB
}

func NewAuthService(
	usuarioRepo usuario.UsuarioRepository,
	empresaRepo empresa.EmpresaRepository,
	cargoRepo cargo.CargoRepository,
	jwtService *jwt.JWTService,
	db *gorm.DB,
) AuthService {
	return &authService{
		usuarioRepo: usuarioRepo,
		empresaRepo: empresaRepo,
		cargoRepo:   cargoRepo,
		jwtService:  jwtService,
		db:          db,
	}
}

func (s *authService) Authenticate(email string, passwordStr string) (string, error) {

	usuari, err := s.usuarioRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("credenciais inválidas")
		}
		return "", err
	}

	if !password.VerificaHashSenha(passwordStr, usuari.Senha) {
		return "", errors.New("credenciais inválidas")
	}

	token, err := s.jwtService.GenerateToken(usuari.ID, usuari.EmpresaID)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *authService) SignUp(empresaReq *model.Empresa, usuarioReq *model.Usuario) (*model.Usuario, string, error) {

	var novoUsuario *model.Usuario
	var token string

	err := s.db.Transaction(func(tx *gorm.DB) error {
		empresaRepoTx := s.empresaRepo.WithTransaction(tx)
		if err := empresaRepoTx.CreateEmpresa(empresaReq); err != nil {
			return err
		}

		permissoes := config.SeedPermissions(s.db)
		config.SetupDefaultRolesAndPermissions(tx, empresaReq.ID, permissoes)

		cargoRepoTx := s.cargoRepo.WithTransaction(tx)
		adminCargo, err := cargoRepoTx.FindByName("Admin", empresaReq.ID)
		if err != nil {
			// Se não encontrarmos o cargo, algo correu muito mal.
			return errors.New("falha ao encontrar o cargo de Admin padrão")
		}

		// 4. Preparar os dados do utilizador antes de o criar
		usuarioReq.EmpresaID = empresaReq.ID
		usuarioReq.CargoID = adminCargo.ID

		senhaHash, err := password.CriptografaSenha(usuarioReq.Senha)
		if err != nil {
			return err // Falha na criptografia? Rollback.
		}
		usuarioReq.Senha = senhaHash

		// 5. Criar o Utilizador
		usuarioRepoTx := s.usuarioRepo.WithTransaction(tx)
		if err := usuarioRepoTx.Save(usuarioReq); err != nil {
			return err // Falhou? Rollback.
		}

		// 6. Se tudo correu bem, guardamos os resultados nas variáveis
		novoUsuario = usuarioReq
		token, err = s.jwtService.GenerateToken(novoUsuario.ID, novoUsuario.EmpresaID)
		if err != nil {
			return err // Falha a gerar o token? Rollback.
		}

		// 7. Retornar nil para dar COMMIT na transação
		return nil
	})

	// Se 'err' não for nulo aqui, a transação falhou.
	// Se for nulo, retornamos o utilizador e o token.
	return novoUsuario, token, err
}
