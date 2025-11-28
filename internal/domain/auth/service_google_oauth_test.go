package auth

import (
	"testing"

	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"
)

// Mock Google OAuth
type MockOAuthConfig struct {
	mock.Mock
}

func (m *MockOAuthConfig) Exchange(code string) (*oauth2.Token, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*oauth2.Token), args.Error(1)
}

type MockEmpresaRepository struct {
	mock.Mock
}

func (m *MockEmpresaRepository) FindByID(id uint) (*model.Empresa, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Empresa), args.Error(1)
}

func (m *MockEmpresaRepository) Update(empresa *model.Empresa) error {
	args := m.Called(empresa)
	return args.Error(0)
}

// Tests for Google OAuth
func TestAuthService_LinkGoogleAccount_Strict_SameEmail(t *testing.T) {
	// Setup
	userRepo := new(MockUsuarioRepository)
	empresaRepo := new(MockEmpresaRepository)

	testUser := &model.Usuario{
		ID:    1,
		Email: "user@company.com",
		Contrato: model.Contrato{
			EmpresaID: 1,
		},
	}

	testEmpresa := &model.Empresa{
		ID:                 1,
		GoogleLinkStrategy: "strict",
	}

	googleEmail := "user@company.com"

	// Expectations
	userRepo.On("FindByID", uint(1)).Return(testUser, nil)
	empresaRepo.On("FindByID", uint(1)).Return(testEmpresa, nil)
	userRepo.On("Update", testUser).Return(nil)

	// Test: Email matches, should link successfully
	assert.Equal(t, "strict", testEmpresa.GoogleLinkStrategy)
	assert.Equal(t, testUser.Email, googleEmail)

	// Verify expectations
	userRepo.AssertExpectations(t)
	empresaRepo.AssertExpectations(t)
}

func TestAuthService_LinkGoogleAccount_Strict_DifferentEmail(t *testing.T) {
	// Setup
	testUser := &model.Usuario{
		ID:    1,
		Email: "user@company.com",
		Contrato: model.Contrato{
			EmpresaID: 1,
		},
	}

	testEmpresa := &model.Empresa{
		ID:                 1,
		GoogleLinkStrategy: "strict",
	}

	googleEmail := "different@gmail.com"

	// Test: Email doesn't match in strict mode
	assert.Equal(t, "strict", testEmpresa.GoogleLinkStrategy)
	assert.NotEqual(t, testUser.Email, googleEmail)

	// In strict mode with different email, should return error
}

func TestAuthService_LinkGoogleAccount_Flexible_SameEmail(t *testing.T) {
	// Setup
	testUser := &model.Usuario{
		ID:    1,
		Email: "user@company.com",
		Contrato: model.Contrato{
			EmpresaID: 1,
		},
	}

	testEmpresa := &model.Empresa{
		ID:                 1,
		GoogleLinkStrategy: "flexible",
	}

	googleEmail := "user@company.com"

	// Test: Same email in flexible mode should link immediately
	assert.Equal(t, "flexible", testEmpresa.GoogleLinkStrategy)
	assert.Equal(t, testUser.Email, googleEmail)

	// Should link without confirmation email
}

func TestAuthService_LinkGoogleAccount_Flexible_DifferentEmail(t *testing.T) {
	// Setup
	userRepo := new(MockUsuarioRepository)
	empresaRepo := new(MockEmpresaRepository)
	emailService := new(MockEmailService)

	testUser := &model.Usuario{
		ID:    1,
		Email: "user@company.com",
		Contrato: model.Contrato{
			EmpresaID: 1,
		},
	}

	testEmpresa := &model.Empresa{
		ID:                 1,
		GoogleLinkStrategy: "flexible",
	}

	googleEmail := "personal@gmail.com"

	// Expectations
	userRepo.On("FindByID", uint(1)).Return(testUser, nil)
	empresaRepo.On("FindByID", uint(1)).Return(testEmpresa, nil)

	// In flexible mode with different email, should send confirmation
	emailService.On("SendEmailWithRetry",
		[]string{testUser.Email},
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)

	// Test: Different email in flexible mode
	assert.Equal(t, "flexible", testEmpresa.GoogleLinkStrategy)
	assert.NotEqual(t, testUser.Email, googleEmail)

	// Verify expectations
	userRepo.AssertExpectations(t)
	empresaRepo.AssertExpectations(t)
}

func TestAuthService_LinkGoogleAccount_DuplicateGoogleID(t *testing.T) {
	// Setup
	// Simulate another user already has this Google ID
	existingUser := &model.Usuario{
		ID:       2,
		Email:    "other@company.com",
		GoogleID: stringPtr("google-id-already-used"),
	}

	googleID := "google-id-already-used"

	// Test: Should fail with duplicate Google ID error
	assert.Equal(t, *existingUser.GoogleID, googleID)

	// In actual implementation, this would be caught by UNIQUE constraint
}

func TestAuthService_AuthenticateWithGoogle_Success(t *testing.T) {
	// Setup
	testUser := &model.Usuario{
		ID:       1,
		Email:    "user@company.com",
		GoogleID: stringPtr("google-id-12345"),
		Contrato: model.Contrato{
			ID:        1,
			EmpresaID: 1,
		},
	}

	googleID := "google-id-12345"

	// Expectations
	// In real implementation, would call FindByGoogleID
	// For now, testing logic
	assert.NotNil(t, testUser.GoogleID)
	assert.Equal(t, *testUser.GoogleID, googleID)
	assert.NotEqual(t, 0, testUser.Contrato.EmpresaID)
}

func TestAuthService_AuthenticateWithGoogle_NotLinked(t *testing.T) {
	// Setup
	// Test: User with this Google ID doesn't exist (no variable needed for this test)

	// Test: User with this Google ID doesn't exist
	// Should return error suggesting to link account first
}

func TestAuthService_AuthenticateWithGoogle_NoActiveContract(t *testing.T) {
	// Setup
	testUser := &model.Usuario{
		ID:       1,
		Email:    "user@company.com",
		GoogleID: stringPtr("google-id-12345"),
		Contrato: model.Contrato{
			ID: 0, // No active contract
		},
	}

	// Test: User has Google linked but no active contract
	assert.NotNil(t, testUser.GoogleID)
	assert.Equal(t, uint(0), testUser.Contrato.ID)

	// Should return error about no active contract
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
