package usuario

import (
	"context"
	"errors"
	"testing"

	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// --- Mocks ---

type MockUsuarioRepo struct {
	mock.Mock
}

func (m *MockUsuarioRepo) Create(usuario *model.Usuario) error {
	args := m.Called(usuario)
	return args.Error(0)
}

func (m *MockUsuarioRepo) Update(id uint, dados map[string]interface{}) error {
	args := m.Called(id, dados)
	return args.Error(0)
}

func (m *MockUsuarioRepo) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUsuarioRepo) FindByID(ctx context.Context, id uint, empresaID uint) (*model.Usuario, error) {
	args := m.Called(ctx, id, empresaID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Usuario), args.Error(1)
}

func (m *MockUsuarioRepo) FindByEmail(email string) (*model.Usuario, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Usuario), args.Error(1)
}

func (m *MockUsuarioRepo) FindByCPF(cpf string) (*model.Usuario, error) {
	args := m.Called(cpf)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Usuario), args.Error(1)
}

func (m *MockUsuarioRepo) GetAll(empresaID uint) ([]model.Usuario, error) {
	args := m.Called(empresaID)
	return args.Get(0).([]model.Usuario), args.Error(1)
}

func (m *MockUsuarioRepo) GetAllPaginated(empresaID uint, page int, limit int) ([]model.Usuario, int64, error) {
	args := m.Called(empresaID, page, limit)
	return args.Get(0).([]model.Usuario), args.Get(1).(int64), args.Error(2)
}

func (m *MockUsuarioRepo) FindAll() ([]model.Usuario, error) {
	args := m.Called()
	return args.Get(0).([]model.Usuario), args.Error(1)
}

func (m *MockUsuarioRepo) FindByGoogleID(googleID string) (*model.Usuario, error) {
	args := m.Called(googleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Usuario), args.Error(1)
}

func (m *MockUsuarioRepo) WithTransaction(tx *gorm.DB) UsuarioRepository {
	return m
}

func (m *MockUsuarioRepo) Save(usuario *model.Usuario) error {
	args := m.Called(usuario)
	return args.Error(0)
}

func (m *MockUsuarioRepo) InvalidarCacheUsuario(id uint, empresaID uint) {
	m.Called(id, empresaID)
}

func (m *MockUsuarioRepo) GetAllActive(empresaID uint) ([]model.Usuario, error) {
	args := m.Called(empresaID)
	return args.Get(0).([]model.Usuario), args.Error(1)
}

type MockCargoRepo struct {
	mock.Mock
}

func (m *MockCargoRepo) Create(cargo *model.Cargo) error {
	args := m.Called(cargo)
	return args.Error(0)
}

func (m *MockCargoRepo) Update(cargo *model.Cargo) error {
	args := m.Called(cargo)
	return args.Error(0)
}

func (m *MockCargoRepo) Delete(id uint, empresaID uint) error {
	args := m.Called(id, empresaID)
	return args.Error(0)
}

func (m *MockCargoRepo) FindByID(ctx context.Context, id uint, empresaID uint) (*model.Cargo, error) {
	args := m.Called(ctx, id, empresaID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Cargo), args.Error(1)
}

func (m *MockCargoRepo) GetAll(empresaID uint) ([]model.Cargo, error) {
	args := m.Called(empresaID)
	return args.Get(0).([]model.Cargo), args.Error(1)
}

type MockEmpresaRepo struct {
	mock.Mock
}

func (m *MockEmpresaRepo) Create(empresa *model.Empresa) error {
	args := m.Called(empresa)
	return args.Error(0)
}

func (m *MockEmpresaRepo) Update(empresa *model.Empresa) error {
	args := m.Called(empresa)
	return args.Error(0)
}

func (m *MockEmpresaRepo) FindByID(id uint) (*model.Empresa, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Empresa), args.Error(1)
}

type MockContratoRepo struct {
	mock.Mock
}

func (m *MockContratoRepo) Create(contrato *model.Contrato) error {
	args := m.Called(contrato)
	return args.Error(0)
}

func (m *MockContratoRepo) Update(contrato *model.Contrato) error {
	args := m.Called(contrato)
	return args.Error(0)
}

func (m *MockContratoRepo) FindByID(id uint) (*model.Contrato, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Contrato), args.Error(1)
}

func (m *MockContratoRepo) WithTransaction(tx *gorm.DB) interface{} { // Ajustar assinatura se necessário
	return m
}

// ContratoRepository interface might define WithTransaction returning ContratoRepository
// Let's assume interface matches.
func (m *MockContratoRepo) Save(contrato *model.Contrato) error {
	args := m.Called(contrato)
	return args.Error(0)
}

type MockLocalidadeRepo struct {
	mock.Mock
}

func (m *MockLocalidadeRepo) Create(localidade *model.Localidade) error {
	args := m.Called(localidade)
	return args.Error(0)
}

func (m *MockLocalidadeRepo) Update(localidade *model.Localidade) error {
	args := m.Called(localidade)
	return args.Error(0)
}

func (m *MockLocalidadeRepo) Delete(id uint, empresaID uint) error {
	args := m.Called(id, empresaID)
	return args.Error(0)
}

func (m *MockLocalidadeRepo) FindByID(ctx context.Context, id uint, empresaID uint) (*model.Localidade, error) {
	args := m.Called(ctx, id, empresaID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Localidade), args.Error(1)
}

func (m *MockLocalidadeRepo) GetAll(empresaID uint) ([]model.Localidade, error) {
	args := m.Called(empresaID)
	return args.Get(0).([]model.Localidade), args.Error(1)
}

// --- Tests ---

func TestUsuarioService_FindByID(t *testing.T) {
	mockRepo := new(MockUsuarioRepo)
	service := NewUsuarioService(nil, mockRepo, nil, nil, nil, nil)

	expectedUser := &model.Usuario{ID: 1, Nome: "Teste"}

	mockRepo.On("FindByID", mock.Anything, uint(1), uint(1)).Return(expectedUser, nil)

	user, err := service.FindByID(1, 1)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	mockRepo.AssertExpectations(t)
}

func TestUsuarioService_FindByID_NotFound(t *testing.T) {
	mockRepo := new(MockUsuarioRepo)
	service := NewUsuarioService(nil, mockRepo, nil, nil, nil, nil)

	mockRepo.On("FindByID", mock.Anything, uint(1), uint(1)).Return(nil, errors.New("not found"))

	user, err := service.FindByID(1, 1)

	assert.Error(t, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestUsuarioService_Update(t *testing.T) {
	mockRepo := new(MockUsuarioRepo)
	service := NewUsuarioService(nil, mockRepo, nil, nil, nil, nil)

	// Usuario não tem EmpresaID direto, depende do contexto ou contrato
	existingUser := &model.Usuario{ID: 1, Email: "old@test.com"}
	dados := map[string]interface{}{"nome": "Novo Nome"}

	mockRepo.On("FindByID", mock.Anything, uint(1), uint(1)).Return(existingUser, nil)
	mockRepo.On("Update", uint(1), dados).Return(nil)

	err := service.Update(1, 1, dados)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUsuarioService_Update_EmailConflict(t *testing.T) {
	mockRepo := new(MockUsuarioRepo)
	service := NewUsuarioService(nil, mockRepo, nil, nil, nil, nil)

	existingUser := &model.Usuario{ID: 1, Email: "old@test.com"}
	otherUser := &model.Usuario{ID: 2, Email: "new@test.com"}
	dados := map[string]interface{}{"email": "new@test.com"}

	mockRepo.On("FindByID", mock.Anything, uint(1), uint(1)).Return(existingUser, nil)
	mockRepo.On("FindByEmail", "new@test.com").Return(otherUser, nil)

	err := service.Update(1, 1, dados)

	assert.Error(t, err)
	assert.Equal(t, "email já está em uso por outro usuário", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestUsuarioService_Delete(t *testing.T) {
	mockRepo := new(MockUsuarioRepo)
	service := NewUsuarioService(nil, mockRepo, nil, nil, nil, nil)

	existingUser := &model.Usuario{ID: 1}

	mockRepo.On("FindByID", mock.Anything, uint(1), uint(1)).Return(existingUser, nil)
	mockRepo.On("Delete", uint(1)).Return(nil)

	err := service.Delete(1, 1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUsuarioService_GetAll(t *testing.T) {
	mockRepo := new(MockUsuarioRepo)
	service := NewUsuarioService(nil, mockRepo, nil, nil, nil, nil)

	usuarios := []model.Usuario{
		{ID: 1, Nome: "Usuario 1"},
		{ID: 2, Nome: "Usuario 2"},
	}

	mockRepo.On("GetAll", uint(1)).Return(usuarios, nil)

	result, err := service.GetAll(1)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	mockRepo.AssertExpectations(t)
}

func TestUsuarioService_GetAllPaginated(t *testing.T) {
	mockRepo := new(MockUsuarioRepo)
	service := NewUsuarioService(nil, mockRepo, nil, nil, nil, nil)

	usuarios := []model.Usuario{
		{ID: 1, Nome: "Usuario 1"},
		{ID: 2, Nome: "Usuario 2"},
	}

	mockRepo.On("GetAllPaginated", uint(1), 1, 10).Return(usuarios, int64(2), nil)

	result, total, err := service.GetAllPaginated(1, 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, int64(2), total)
	mockRepo.AssertExpectations(t)
}

func TestUsuarioService_FindAll(t *testing.T) {
	mockRepo := new(MockUsuarioRepo)
	service := NewUsuarioService(nil, mockRepo, nil, nil, nil, nil)

	usuarios := []model.Usuario{
		{ID: 1, Nome: "Usuario 1"},
		{ID: 2, Nome: "Usuario 2"},
		{ID: 3, Nome: "Usuario 3"},
	}

	mockRepo.On("FindAll").Return(usuarios, nil)

	result, err := service.FindAll()

	assert.NoError(t, err)
	assert.Equal(t, 3, len(result))
	mockRepo.AssertExpectations(t)
}

func TestUsuarioService_Delete_NotFound(t *testing.T) {
	mockRepo := new(MockUsuarioRepo)
	service := NewUsuarioService(nil, mockRepo, nil, nil, nil, nil)

	mockRepo.On("FindByID", mock.Anything, uint(999), uint(1)).Return(nil, errors.New("not found"))

	err := service.Delete(999, 1)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUsuarioService_Update_NotFound(t *testing.T) {
	mockRepo := new(MockUsuarioRepo)
	service := NewUsuarioService(nil, mockRepo, nil, nil, nil, nil)

	dados := map[string]interface{}{"nome": "Novo Nome"}

	mockRepo.On("FindByID", mock.Anything, uint(999), uint(1)).Return(nil, errors.New("not found"))

	err := service.Update(999, 1, dados)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}
