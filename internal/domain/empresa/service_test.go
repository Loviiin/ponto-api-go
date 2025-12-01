package empresa

import (
	"context"
	"testing"

	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// --- Mocks ---
type MockEmpresaRepo struct {
	mock.Mock
}

func (m *MockEmpresaRepo) CreateEmpresa(empresa *model.Empresa) error {
	args := m.Called(empresa)
	return args.Error(0)
}

func (m *MockEmpresaRepo) UpdateEmpresa(idempresa uint, dados map[string]interface{}) error {
	args := m.Called(idempresa, dados)
	return args.Error(0)
}

func (m *MockEmpresaRepo) FindByID(ctx context.Context, id uint) (*model.Empresa, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Empresa), args.Error(1)
}

func (m *MockEmpresaRepo) GetEmpresaByID(idempresa uint) (*model.Empresa, error) {
	args := m.Called(idempresa)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Empresa), args.Error(1)
}

func (m *MockEmpresaRepo) GetAllEmpresas() ([]model.Empresa, error) {
	args := m.Called()
	return args.Get(0).([]model.Empresa), args.Error(1)
}

func (m *MockEmpresaRepo) DeleteEmpresa(idempresa uint) error {
	args := m.Called(idempresa)
	return args.Error(0)
}

func (m *MockEmpresaRepo) WithTransaction(tx *gorm.DB) EmpresaRepository {
	return m
}

// --- Tests ---

func TestEmpresaService_GetEmpresaByIDSer(t *testing.T) {
	mockRepo := new(MockEmpresaRepo)
	service := NewEmpresaService(mockRepo)

	expected := &model.Empresa{ID: 1, RazaoSocial: "Teste Ltda"}

	mockRepo.On("GetEmpresaByID", uint(1)).Return(expected, nil)

	got, err := service.GetEmpresaByIDSer(1)

	assert.NoError(t, err)
	assert.Equal(t, expected, got)
	mockRepo.AssertExpectations(t)
}

func TestEmpresaService_CreateEmpresa(t *testing.T) {
	mockRepo := new(MockEmpresaRepo)
	service := NewEmpresaService(mockRepo)

	empresa := &model.Empresa{RazaoSocial: "Nova Empresa", CNPJ: "12345678901234"}

	mockRepo.On("CreateEmpresa", empresa).Return(nil)

	err := service.CreateEmpresa(empresa)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestEmpresaService_UpdateEmpresaSer(t *testing.T) {
	mockRepo := new(MockEmpresaRepo)
	service := NewEmpresaService(mockRepo)

	empresa := &model.Empresa{ID: 1, RazaoSocial: "Empresa Atualizada"}
	dados := map[string]interface{}{"razao_social": "Empresa Atualizada"}

	// Service calls FindByID first to check existence
	mockRepo.On("FindByID", mock.Anything, uint(1)).Return(empresa, nil)
	mockRepo.On("UpdateEmpresa", uint(1), dados).Return(nil)

	err := service.UpdateEmpresaSer(1, dados)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestEmpresaService_DeleteEmpresaSer(t *testing.T) {
	mockRepo := new(MockEmpresaRepo)
	service := NewEmpresaService(mockRepo)

	empresa := &model.Empresa{ID: 1}

	// Service calls FindByID first to check existence
	mockRepo.On("FindByID", mock.Anything, uint(1)).Return(empresa, nil)
	mockRepo.On("DeleteEmpresa", uint(1)).Return(nil)

	err := service.DeleteEmpresaSer(1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestEmpresaService_GetAllEmpresasSer(t *testing.T) {
	mockRepo := new(MockEmpresaRepo)
	service := NewEmpresaService(mockRepo)

	expected := []model.Empresa{{ID: 1}, {ID: 2}}

	mockRepo.On("GetAllEmpresas").Return(expected, nil)

	got, err := service.GetAllEmpresasSer()

	assert.NoError(t, err)
	assert.Equal(t, expected, got)
	mockRepo.AssertExpectations(t)
}
