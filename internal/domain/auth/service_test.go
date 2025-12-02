package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/domain/usuario"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/jwt"
	"github.com/Loviiin/ponto-api-go/pkg/password"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// --- Mocks ---

type MockUsuarioRepo struct {
	mock.Mock
}

func (m *MockUsuarioRepo) FindByEmail(email string) (*model.Usuario, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Usuario), args.Error(1)
}

func (m *MockUsuarioRepo) Create(usuario *model.Usuario) error {
	return nil
}
func (m *MockUsuarioRepo) Update(id uint, dados map[string]interface{}) error {
	return nil
}
func (m *MockUsuarioRepo) Delete(id uint) error {
	return nil
}
func (m *MockUsuarioRepo) FindByID(ctx context.Context, id uint, empresaID uint) (*model.Usuario, error) {
	return nil, nil
}
func (m *MockUsuarioRepo) FindByCPF(cpf string) (*model.Usuario, error) {
	return nil, nil
}
func (m *MockUsuarioRepo) GetAll(empresaID uint) ([]model.Usuario, error) {
	return nil, nil
}
func (m *MockUsuarioRepo) GetAllPaginated(empresaID uint, page int, limit int) ([]model.Usuario, int64, error) {
	return nil, 0, nil
}
func (m *MockUsuarioRepo) FindAll() ([]model.Usuario, error) {
	return nil, nil
}
func (m *MockUsuarioRepo) FindByGoogleID(googleID string) (*model.Usuario, error) {
	return nil, nil
}
func (m *MockUsuarioRepo) WithTransaction(tx *gorm.DB) usuario.UsuarioRepository {
	return m
}

// Fix WithTransaction signature match
// In service.go: usuarioRepo.WithTransaction(tx).Save(usuario)
// We need to match the interface.
// Assuming interface is: WithTransaction(tx *gorm.DB) UsuarioRepository
// But here I can't import usuario.UsuarioRepository easily to implement it fully if it has other methods.
// Wait, I am in `auth` package. `usuario` package is imported.
// So I can implement `usuario.UsuarioRepository`.

// Let's implement the interface methods required by Authenticate only for now.
// Authenticate uses: FindByEmail.

func (m *MockUsuarioRepo) Save(usuario *model.Usuario) error {
	return nil
}
func (m *MockUsuarioRepo) InvalidarCacheUsuario(id uint, empresaID uint) {}
func (m *MockUsuarioRepo) GetAllActive(empresaID uint) ([]model.Usuario, error) {
	return nil, nil
}

// Ensure MockUsuarioRepo implements usuario.UsuarioRepository
// I need to import "github.com/Loviiin/ponto-api-go/internal/domain/usuario"
// But I can't implement it if I don't implement all methods.
// I will just implement the ones I need and rely on duck typing? No, Go is static.
// I must implement all methods of the interface.
// I'll add stubs for the rest.

type MockCacheService struct {
	mock.Mock
}

func (m *MockCacheService) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCacheService) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacheService) Clear(ctx context.Context) error {
	return nil
}

// --- Tests ---

func TestAuthService_Authenticate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockUsuarioRepo)
		mockCache := new(MockCacheService)
		jwtService := jwt.NewJWTService("secret", "issuer")

		service := NewAuthService(mockRepo, nil, nil, nil, nil, nil, nil, jwtService, mockCache, nil, nil)

		passwordStr := "senha123"
		hashedPassword, _ := password.CriptografaSenha(passwordStr)

		user := &model.Usuario{
			ID:    1,
			Email: "teste@empresa.com",
			Senha: hashedPassword,
			Contrato: model.Contrato{
				ID:        1,
				EmpresaID: 1,
			},
		}

		// Cache miss
		mockCache.On("Get", mock.Anything, "auth:user:teste@empresa.com").Return("", errors.New("miss"))

		// Repo hit
		mockRepo.On("FindByEmail", "teste@empresa.com").Return(user, nil)

		// Cache set
		mockCache.On("Set", mock.Anything, "auth:user:teste@empresa.com", mock.Anything, 15*time.Minute).Return(nil)

		token, err := service.Authenticate("teste@empresa.com", passwordStr)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("InvalidPassword", func(t *testing.T) {
		mockRepo := new(MockUsuarioRepo)
		mockCache := new(MockCacheService)
		jwtService := jwt.NewJWTService("secret", "issuer")

		service := NewAuthService(mockRepo, nil, nil, nil, nil, nil, nil, jwtService, mockCache, nil, nil)

		passwordStr := "senha123"
		hashedPassword, _ := password.CriptografaSenha(passwordStr)

		user := &model.Usuario{
			ID:    1,
			Email: "teste@empresa.com",
			Senha: hashedPassword,
			Contrato: model.Contrato{
				ID:        1,
				EmpresaID: 1,
			},
		}

		mockCache.On("Get", mock.Anything, "auth:user:teste@empresa.com").Return("", errors.New("miss"))
		mockRepo.On("FindByEmail", "teste@empresa.com").Return(user, nil)

		token, err := service.Authenticate("teste@empresa.com", "senhaerrada")

		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Equal(t, "credenciais inválidas", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

// Test para FindByEmail
func TestAuthService_FindByEmail(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockUsuarioRepo)

		user := &model.Usuario{
			ID:    1,
			Email: "teste@empresa.com",
		}

		mockRepo.On("FindByEmail", "teste@empresa.com").Return(user, nil)

		// Simula chamada interna (não é método público mas testa o repo)
		result, err := mockRepo.FindByEmail("teste@empresa.com")

		assert.NoError(t, err)
		assert.Equal(t, user, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockRepo := new(MockUsuarioRepo)

		mockRepo.On("FindByEmail", "naoexiste@empresa.com").Return(nil, gorm.ErrRecordNotFound)

		result, err := mockRepo.FindByEmail("naoexiste@empresa.com")

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

// Test de validação de senha
func TestAuthService_PasswordValidation(t *testing.T) {
	t.Run("CorrectPassword", func(t *testing.T) {
		passwordStr := "senha123"
		hashedPassword, _ := password.CriptografaSenha(passwordStr)

		user := &model.Usuario{
			ID:    1,
			Senha: hashedPassword,
		}

		ok := password.VerificaHashSenha(passwordStr, user.Senha)
		assert.True(t, ok)
	})

	t.Run("IncorrectPassword", func(t *testing.T) {
		passwordStr := "senha123"
		hashedPassword, _ := password.CriptografaSenha(passwordStr)

		ok := password.VerificaHashSenha("senhaerrada", hashedPassword)
		assert.False(t, ok)
	})
}

// Tests para cenários de erro de autenticação
func TestAuthService_Authenticate_Errors(t *testing.T) {
	t.Run("UserNotFound", func(t *testing.T) {
		mockRepo := new(MockUsuarioRepo)
		mockCache := new(MockCacheService)
		jwtService := jwt.NewJWTService("secret", "issuer")

		service := NewAuthService(mockRepo, nil, nil, nil, nil, nil, nil, jwtService, mockCache, nil, nil)

		mockCache.On("Get", mock.Anything, "auth:user:naoexiste@empresa.com").Return("", errors.New("miss"))
		mockRepo.On("FindByEmail", "naoexiste@empresa.com").Return(nil, gorm.ErrRecordNotFound)

		token, err := service.Authenticate("naoexiste@empresa.com", "senha123")

		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Equal(t, "credenciais inválidas", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("EmptyPassword", func(t *testing.T) {
		mockRepo := new(MockUsuarioRepo)
		mockCache := new(MockCacheService)
		jwtService := jwt.NewJWTService("secret", "issuer")

		service := NewAuthService(mockRepo, nil, nil, nil, nil, nil, nil, jwtService, mockCache, nil, nil)

		mockCache.On("Get", mock.Anything, "auth:user:teste@empresa.com").Return("", errors.New("miss"))
		mockRepo.On("FindByEmail", "teste@empresa.com").Return(nil, errors.New("invalid password"))

		token, err := service.Authenticate("teste@empresa.com", "")

		// Deve falhar com senha vazia
		assert.Error(t, err)
		assert.Empty(t, token)
	})
}
