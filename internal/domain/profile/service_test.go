package profile

import (
	"context"
	"testing"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/password"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// --- Mocks ---

type MockProfileRepo struct {
	mock.Mock
}

func (m *MockProfileRepo) GetUserByID(userID uint) (*model.Usuario, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Usuario), args.Error(1)
}

func (m *MockProfileRepo) GetLocalidadesByEmpresa(empresaID uint) ([]model.Localidade, error) {
	args := m.Called(empresaID)
	return args.Get(0).([]model.Localidade), args.Error(1)
}

func (m *MockProfileRepo) GetPermissoesByUserID(userID uint) ([]model.Permissao, error) {
	args := m.Called(userID)
	return args.Get(0).([]model.Permissao), args.Error(1)
}

func (m *MockProfileRepo) InvalidateUserCache(userID uint) {
	m.Called(userID)
}

func (m *MockProfileRepo) UpdateUser(user *model.Usuario) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockProfileRepo) GetLatestBancoHoras(userID, empresaID uint) (*model.LogBancoHoras, error) {
	args := m.Called(userID, empresaID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.LogBancoHoras), args.Error(1)
}

func (m *MockProfileRepo) GetPontoStatsByUserAndMonth(userID uint, month, year int) (*PontoMonthStats, error) {
	args := m.Called(userID, month, year)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*PontoMonthStats), args.Error(1)
}

func (m *MockProfileRepo) GetTotalPontosCount(userID uint) (int64, error) {
	args := m.Called(userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockProfileRepo) GetJustificativasCount(userID, empresaID uint) (int64, error) {
	args := m.Called(userID, empresaID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockProfileRepo) GetJustificativasPendentesCount(userID, empresaID uint) (int64, error) {
	args := m.Called(userID, empresaID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockProfileRepo) GetRecentPontos(userID uint, limit int) ([]model.RegistroPonto, error) {
	args := m.Called(userID, limit)
	return args.Get(0).([]model.RegistroPonto), args.Error(1)
}

func (m *MockProfileRepo) GetPontosByMonth(userID uint, month, year int) ([]model.RegistroPonto, error) {
	args := m.Called(userID, month, year)
	return args.Get(0).([]model.RegistroPonto), args.Error(1)
}

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

type MockAuditService struct {
	mock.Mock
}

func (m *MockAuditService) LogAction(userID *uint, empresaID *uint, action string, targetType string, targetID uint, oldData, newData interface{}, ip, userAgent string) error {
	args := m.Called(userID, empresaID, action, targetType, targetID, oldData, newData, ip, userAgent)
	return args.Error(0)
}

func (m *MockAuditService) GetUserAuditLogs(usuarioID uint, limit int) ([]model.AuditLog, error) {
	args := m.Called(usuarioID, limit)
	return args.Get(0).([]model.AuditLog), args.Error(1)
}

func (m *MockAuditService) GetEntityAuditLogs(entidade string, entidadeID uint, limit int) ([]model.AuditLog, error) {
	args := m.Called(entidade, entidadeID, limit)
	return args.Get(0).([]model.AuditLog), args.Error(1)
}

// --- Tests ---

func TestProfileService_GetMyProfile(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockProfileRepo)
		service := NewService(mockRepo, nil, nil, nil)

		user := &model.Usuario{
			ID:    1,
			Nome:  "Teste",
			Email: "teste@empresa.com",
			Contrato: model.Contrato{
				ID:        1,
				EmpresaID: 1,
				Empresa: model.Empresa{
					ID:           1,
					NomeFantasia: "Empresa Teste",
				},
				Cargo: model.Cargo{
					ID:   1,
					Nome: "Cargo Teste",
				},
			},
		}

		mockRepo.On("GetUserByID", uint(1)).Return(user, nil)
		mockRepo.On("GetLocalidadesByEmpresa", uint(1)).Return([]model.Localidade{}, nil)
		mockRepo.On("GetPermissoesByUserID", uint(1)).Return([]model.Permissao{}, nil)

		profile, err := service.GetMyProfile(1)

		assert.NoError(t, err)
		assert.NotNil(t, profile)
		assert.Equal(t, "Teste", profile.Nome)
		mockRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockRepo := new(MockProfileRepo)
		service := NewService(mockRepo, nil, nil, nil)

		mockRepo.On("GetUserByID", uint(99)).Return(nil, gorm.ErrRecordNotFound)

		profile, err := service.GetMyProfile(99)

		assert.Error(t, err)
		assert.Nil(t, profile)
		assert.Equal(t, "usuário não encontrado", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestProfileService_ChangePassword(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockProfileRepo)
		mockAudit := new(MockAuditService)
		service := NewService(mockRepo, nil, nil, mockAudit)

		currentPassword := "SenhaAtual123!"
		hashedCurrent, _ := password.CriptografaSenha(currentPassword)

		user := &model.Usuario{
			ID:    1,
			Senha: hashedCurrent,
			Contrato: model.Contrato{
				ID:        1,
				EmpresaID: 1,
			},
		}

		req := ChangePasswordRequest{
			SenhaAtual:     currentPassword,
			NovaSenha:      "NovaSenha123!",
			ConfirmarSenha: "NovaSenha123!",
		}

		mockRepo.On("GetUserByID", uint(1)).Return(user, nil)
		mockRepo.On("UpdateUser", mock.AnythingOfType("*model.Usuario")).Return(nil)

		// Audit log expectation
		mockAudit.On("LogAction", mock.Anything, mock.Anything, "CHANGE_PASSWORD", "usuario", uint(1), nil, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		err := service.ChangePassword(1, req, "127.0.0.1", "TestAgent")

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockAudit.AssertExpectations(t)
	})

	t.Run("PasswordMismatch", func(t *testing.T) {
		mockRepo := new(MockProfileRepo)
		service := NewService(mockRepo, nil, nil, nil)

		req := ChangePasswordRequest{
			SenhaAtual:     "SenhaAtual123!",
			NovaSenha:      "NovaSenha123!",
			ConfirmarSenha: "OutraSenha123!",
		}

		err := service.ChangePassword(1, req, "127.0.0.1", "TestAgent")

		assert.Error(t, err)
		assert.Equal(t, "nova senha e confirmação não coincidem", err.Error())
	})
}

// Test de validação de permissões do usuário
func TestProfileService_GetMyPermissions(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := new(MockProfileRepo)
		service := NewService(mockRepo, nil, nil, nil)

		permissoes := []model.Permissao{
			{ID: 1, Nome: "VISUALIZAR_PONTO"},
			{ID: 2, Nome: "EDITAR_PONTO"},
		}

		mockRepo.On("GetPermissoesByUserID", uint(1)).Return(permissoes, nil)

		permissions, err := service.GetMyPermissions(1)

		assert.NoError(t, err)
		assert.NotNil(t, permissions)
		assert.Equal(t, 2, len(permissions.Permissoes))
		mockRepo.AssertExpectations(t)
	})
}

// Test para ChangePassword com senha incorreta
func TestProfileService_ChangePassword_WrongPassword(t *testing.T) {
	mockRepo := new(MockProfileRepo)
	service := NewService(mockRepo, nil, nil, nil)

	currentPassword := "SenhaAtual123!"
	hashedCurrent, _ := password.CriptografaSenha(currentPassword)

	user := &model.Usuario{
		ID:    1,
		Senha: hashedCurrent,
		Contrato: model.Contrato{
			ID:        1,
			EmpresaID: 1,
		},
	}

	req := ChangePasswordRequest{
		SenhaAtual:     "SenhaErrada!",
		NovaSenha:      "NovaSenha123!",
		ConfirmarSenha: "NovaSenha123!",
	}

	mockRepo.On("GetUserByID", uint(1)).Return(user, nil)

	err := service.ChangePassword(1, req, "127.0.0.1", "TestAgent")

	assert.Error(t, err)
	assert.Equal(t, "senha atual incorreta", err.Error())
	mockRepo.AssertExpectations(t)
}
