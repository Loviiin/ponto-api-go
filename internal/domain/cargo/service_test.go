package cargo

import (
	"context"
	"testing"

	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// --- Mocks ---
type MockCargoRepo struct {
	mock.Mock
}

func (m *MockCargoRepo) Create(cargo *model.Cargo) error {
	args := m.Called(cargo)
	return args.Error(0)
}

func (m *MockCargoRepo) Update(id uint, empresaID uint, dados map[string]interface{}) error {
	args := m.Called(id, empresaID, dados)
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

func (m *MockCargoRepo) GetAllByEmpresaID(empresaID uint) ([]model.Cargo, error) {
	args := m.Called(empresaID)
	return args.Get(0).([]model.Cargo), args.Error(1)
}

func (m *MockCargoRepo) AddPermissionToCargo(cargoID uint, permissaoID uint) error {
	args := m.Called(cargoID, permissaoID)
	return args.Error(0)
}

func (m *MockCargoRepo) RemovePermissionFromCargo(cargoID uint, permissaoID uint) error {
	args := m.Called(cargoID, permissaoID)
	return args.Error(0)
}

func (m *MockCargoRepo) GetPermissionsByCargo(cargoID uint, empresaID uint) ([]model.Permissao, error) {
	args := m.Called(cargoID, empresaID)
	return args.Get(0).([]model.Permissao), args.Error(1)
}

func (m *MockCargoRepo) FindByName(nome string, empresaID uint) (*model.Cargo, error) {
	args := m.Called(nome, empresaID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Cargo), args.Error(1)
}

func (m *MockCargoRepo) HasUsuarios(cargoID uint, empresaID uint) (bool, error) {
	args := m.Called(cargoID, empresaID)
	return args.Bool(0), args.Error(1)
}

func (m *MockCargoRepo) WithTransaction(tx *gorm.DB) CargoRepository {
	return m
}

// --- Tests ---

func TestCargoService_Create(t *testing.T) {
	mockRepo := new(MockCargoRepo)
	service := NewCargoService(mockRepo)

	cargo := &model.Cargo{Nome: "Desenvolvedor", EmpresaID: 1}

	mockRepo.On("Create", cargo).Return(nil)

	err := service.Create(cargo)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCargoService_Update(t *testing.T) {
	mockRepo := new(MockCargoRepo)
	service := NewCargoService(mockRepo)

	cargo := &model.Cargo{ID: 1, Nome: "Senior Dev", EmpresaID: 1}
	dados := map[string]interface{}{"nome": "Senior Dev"}

	// Verify existence first
	mockRepo.On("FindByID", mock.Anything, uint(1), uint(1)).Return(cargo, nil)
	mockRepo.On("Update", uint(1), uint(1), dados).Return(nil)

	err := service.Update(1, 1, dados)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCargoService_Delete(t *testing.T) {
	mockRepo := new(MockCargoRepo)
	service := NewCargoService(mockRepo)

	cargo := &model.Cargo{ID: 1, EmpresaID: 1}

	// Verify existence first
	mockRepo.On("FindByID", mock.Anything, uint(1), uint(1)).Return(cargo, nil)
	mockRepo.On("Delete", uint(1), uint(1)).Return(nil)

	err := service.Delete(1, 1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCargoService_GetAll(t *testing.T) {
	mockRepo := new(MockCargoRepo)
	service := NewCargoService(mockRepo)

	expected := []model.Cargo{{ID: 1, Nome: "Dev"}, {ID: 2, Nome: "QA"}}

	mockRepo.On("GetAllByEmpresaID", uint(1)).Return(expected, nil)

	got, err := service.GetAllByEmpresaID(1)

	assert.NoError(t, err)
	assert.Equal(t, expected, got)
	mockRepo.AssertExpectations(t)
}
