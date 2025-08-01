package auth

import (
	"errors"
	"github.com/Loviiin/ponto-api-go/internal/usuario"
	"github.com/Loviiin/ponto-api-go/pkg/jwt"
	password2 "github.com/Loviiin/ponto-api-go/pkg/password"
	"gorm.io/gorm"
)

type AuthService interface {
	Authenticate(email string, password string) (string, error)
}

type authService struct {
	usuarioRepo usuario.UsuarioRepository
	jwtService  *jwt.JWTService
}

func NewAuthService(usuarioRepo usuario.UsuarioRepository, jwtService *jwt.JWTService) AuthService {
	return &authService{
		usuarioRepo: usuarioRepo,
		jwtService:  jwtService,
	}
}

func (s *authService) Authenticate(email string, password string) (string, error) {

	usuario, err := s.usuarioRepo.FindByEmail(email)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("credenciais inválidas")
		}
		return "", err
	}

	if !password2.VerificaHashSenha(password, usuario.Senha) {
		return "", errors.New("credenciais inválidas")
	}

	token, err := s.jwtService.GenerateToken(usuario.ID)
	if err != nil {
		return "", err
	}
	return token, nil
}
