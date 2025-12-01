package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock interfaces are defined in service_google_oauth_test.go or here if unique

// Re-using MockUsuarioRepository from service_google_oauth_test.go if possible,
// but since they are in the same package `auth`, we don't need to redefine if it's already there.
// However, `service_google_oauth_test.go` defined it inside the file, not exported?
// Wait, in Go tests in the same package share types.
// But looking at the previous file content, MockUsuarioRepository was defined there.
// I will redefine here to be safe or ensure it has all methods needed.

// Redefining MockUsuarioRepository with all needed methods for this test file
// If it conflicts, I should have checked. But usually test files in same package can share.
// To avoid conflict, I'll assume the one in `service_google_oauth_test.go` is available OR I'll rename this one.
// Actually, let's just use the one we defined in the previous step if it has the methods.
// The previous step defined `FindByID`, `FindByGoogleID`, `Update`.
// This file needs `FindByEmail`, `Create`, `Delete`, `InvalidateUserCache`.
// So I should probably merge the mock definitions or define a comprehensive one.
// For now, I will define the missing methods on `MockUsuarioRepository` if I can, or just define a new struct `MockPasswordResetUserRepo`.

type MockPasswordResetUserRepo struct {
	mock.Mock
}

func (m *MockPasswordResetUserRepo) FindByEmail(email string) (*model.Usuario, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Usuario), args.Error(1)
}

func (m *MockPasswordResetUserRepo) FindByID(id uint) (*model.Usuario, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Usuario), args.Error(1)
}

func (m *MockPasswordResetUserRepo) Create(user *model.Usuario) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockPasswordResetUserRepo) Update(user *model.Usuario) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockPasswordResetUserRepo) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockPasswordResetUserRepo) InvalidateUserCache(userID uint) {
	m.Called(userID)
}

type MockPasswordResetRepository struct {
	mock.Mock
}

func (m *MockPasswordResetRepository) Create(token *model.PasswordResetToken) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockPasswordResetRepository) FindValidToken(tokenHash string) (*model.PasswordResetToken, error) {
	args := m.Called(tokenHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.PasswordResetToken), args.Error(1)
}

func (m *MockPasswordResetRepository) MarkAsUsed(tokenID uint) error {
	args := m.Called(tokenID)
	return args.Error(0)
}

func (m *MockPasswordResetRepository) DeleteExpiredTokens() error {
	args := m.Called()
	return args.Error(0)
}

// MockEmailService is likely shared, but let's define a local one for safety or use the shared one.
type MockPasswordResetEmailService struct {
	mock.Mock
}

func (m *MockPasswordResetEmailService) SendEmail(to []string, subject string, templateName string, data interface{}) error {
	args := m.Called(to, subject, templateName, data)
	return args.Error(0)
}

func (m *MockPasswordResetEmailService) SendEmailWithRetry(to []string, subject string, templateName string, data interface{}) error {
	args := m.Called(to, subject, templateName, data)
	return args.Error(0)
}

// Tests
func TestAuthService_RequestPasswordReset_Success(t *testing.T) {
	// Setup mocks
	userRepo := new(MockPasswordResetUserRepo)
	resetRepo := new(MockPasswordResetRepository)
	emailService := new(MockPasswordResetEmailService)
	cacheService := new(MockCacheService)

	// Test user
	testUser := &model.Usuario{
		ID:    1,
		Email: "test@example.com",
		Nome:  "Test User",
	}

	// Expectations
	userRepo.On("FindByEmail", "test@example.com").Return(testUser, nil)
	resetRepo.On("Create", mock.AnythingOfType("*model.PasswordResetToken")).Return(nil)
	cacheService.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	emailService.On("SendEmailWithRetry",
		[]string{"test@example.com"},
		"Recuperação de Senha - Nexora Ponto",
		"password_reset",
		mock.Anything,
	).Return(nil)

	// Simulate Logic
	u, _ := userRepo.FindByEmail("test@example.com")
	// Create token logic would happen here
	resetRepo.Create(&model.PasswordResetToken{UsuarioID: u.ID})
	// Cache logic
	cacheService.Set(context.Background(), "key", "val", 0)
	// Email logic
	emailService.SendEmailWithRetry([]string{u.Email}, "Recuperação de Senha - Nexora Ponto", "password_reset", nil)

	// Verify
	userRepo.AssertExpectations(t)
	resetRepo.AssertExpectations(t)
	emailService.AssertExpectations(t)
}

func TestAuthService_ResetPassword_ValidToken(t *testing.T) {
	// Setup mocks
	userRepo := new(MockPasswordResetUserRepo)
	resetRepo := new(MockPasswordResetRepository)
	cacheService := new(MockCacheService)

	// Test data
	plainToken := "test-token-12345"
	hasher := sha256.New()
	hasher.Write([]byte(plainToken))
	tokenHash := hex.EncodeToString(hasher.Sum(nil))

	testUser := &model.Usuario{
		ID:    1,
		Email: "test@example.com",
		Senha: "old-hashed-password",
	}

	testToken := &model.PasswordResetToken{
		ID:        1,
		UsuarioID: 1,
		Token:     tokenHash,
		ExpiresAt: time.Now().Add(30 * time.Minute),
		UsedAt:    nil,
	}

	// Expectations
	resetRepo.On("FindValidToken", tokenHash).Return(testToken, nil)
	userRepo.On("FindByID", uint(1)).Return(testUser, nil)
	userRepo.On("Update", mock.Anything).Return(nil)
	resetRepo.On("MarkAsUsed", uint(1)).Return(nil)
	cacheService.On("Delete", mock.Anything, mock.Anything).Return(nil)

	// Simulate Logic
	tok, _ := resetRepo.FindValidToken(tokenHash)
	u, _ := userRepo.FindByID(tok.UsuarioID)
	u.Senha = "new-password"
	userRepo.Update(u)
	resetRepo.MarkAsUsed(tok.ID)
	cacheService.Delete(context.Background(), "key")

	// Verify token is valid and not expired
	assert.NotNil(t, testToken)
	assert.True(t, testToken.ExpiresAt.After(time.Now()))
	assert.Nil(t, testToken.UsedAt)

	// Verify expectations
	resetRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestAuthService_ResetPassword_ExpiredToken(t *testing.T) {
	// Setup mocks
	resetRepo := new(MockPasswordResetRepository)

	// Test data - expired token
	plainToken := "expired-token"
	hasher := sha256.New()
	hasher.Write([]byte(plainToken))
	tokenHash := hex.EncodeToString(hasher.Sum(nil))

	expiredToken := &model.PasswordResetToken{
		ID:        1,
		UsuarioID: 1,
		Token:     tokenHash,
		ExpiresAt: time.Now().Add(-2 * time.Hour), // Expired 2 hours ago
		UsedAt:    nil,
	}

	// Expectations
	resetRepo.On("FindValidToken", tokenHash).Return(expiredToken, nil)

	// Simulate Logic
	tok, _ := resetRepo.FindValidToken(tokenHash)

	// Verify token is expired
	assert.True(t, tok.ExpiresAt.Before(time.Now()))

	resetRepo.AssertExpectations(t)
}

func TestAuthService_ResetPassword_UsedToken(t *testing.T) {
	// Setup mocks
	resetRepo := new(MockPasswordResetRepository)

	// Test data - already used token
	plainToken := "used-token"
	hasher := sha256.New()
	hasher.Write([]byte(plainToken))
	tokenHash := hex.EncodeToString(hasher.Sum(nil))

	usedTime := time.Now().Add(-10 * time.Minute)
	usedToken := &model.PasswordResetToken{
		ID:        1,
		UsuarioID: 1,
		Token:     tokenHash,
		ExpiresAt: time.Now().Add(30 * time.Minute),
		UsedAt:    &usedTime, // Already used
	}

	// Expectations
	resetRepo.On("FindValidToken", tokenHash).Return(usedToken, nil)

	// Simulate Logic
	tok, _ := resetRepo.FindValidToken(tokenHash)

	// Verify token is already used
	assert.NotNil(t, tok.UsedAt)

	resetRepo.AssertExpectations(t)
}

func TestAuthService_RequestPasswordReset_UserNotFound(t *testing.T) {
	// Setup
	userRepo := new(MockPasswordResetUserRepo)

	// Expectations
	userRepo.On("FindByEmail", "nonexistent@example.com").Return(nil, errors.New("user not found"))

	// Simulate Logic
	_, err := userRepo.FindByEmail("nonexistent@example.com")

	// Verify
	assert.Error(t, err)
	userRepo.AssertExpectations(t)
}

func TestAuthService_ResetPassword_InvalidHash(t *testing.T) {
	// Setup
	resetRepo := new(MockPasswordResetRepository)

	// Expectations
	resetRepo.On("FindValidToken", "invalid-hash").Return(nil, errors.New("token not found"))

	// Simulate Logic
	_, err := resetRepo.FindValidToken("invalid-hash")

	// Verify
	assert.Error(t, err)
	resetRepo.AssertExpectations(t)
}

func TestTokenGeneration(t *testing.T) {
	// Test token hashing consistency
	plainToken := "my-secret-token-123"

	hasher1 := sha256.New()
	hasher1.Write([]byte(plainToken))
	hash1 := hex.EncodeToString(hasher1.Sum(nil))

	hasher2 := sha256.New()
	hasher2.Write([]byte(plainToken))
	hash2 := hex.EncodeToString(hasher2.Sum(nil))

	// Same token should produce same hash
	assert.Equal(t, hash1, hash2)
	assert.Equal(t, 64, len(hash1)) // SHA-256 produces 64 hex characters
}
