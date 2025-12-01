package justificativa

import (
	"context"
	"testing"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/domain/ponto"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// --- Mocks ---
type MockJustificativaRepo struct {
	mock.Mock
}

func (m *MockJustificativaRepo) Create(justificativa *model.Justificativa) error {
	args := m.Called(justificativa)
	return args.Error(0)
}

func (m *MockJustificativaRepo) Update(justificativa *model.Justificativa) error {
	args := m.Called(justificativa)
	return args.Error(0)
}

func (m *MockJustificativaRepo) FindByID(id uint, empresaID uint) (*model.Justificativa, error) {
	args := m.Called(id, empresaID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Justificativa), args.Error(1)
}

func (m *MockJustificativaRepo) FindByUsuarioID(usuarioID uint, empresaID uint) ([]model.Justificativa, error) {
	args := m.Called(usuarioID, empresaID)
	return args.Get(0).([]model.Justificativa), args.Error(1)
}

func (m *MockJustificativaRepo) FindByStatus(empresaID uint, status string) ([]model.Justificativa, error) {
	args := m.Called(empresaID, status)
	return args.Get(0).([]model.Justificativa), args.Error(1)
}

func (m *MockJustificativaRepo) FindByUsuarioIDAndPeriodo(usuarioID uint, empresaID uint, inicio, fim time.Time) ([]model.Justificativa, error) {
	args := m.Called(usuarioID, empresaID, inicio, fim)
	return args.Get(0).([]model.Justificativa), args.Error(1)
}

func (m *MockJustificativaRepo) InvalidarCacheEmpresa(empresaID uint) {
	m.Called(empresaID)
}

func (m *MockJustificativaRepo) InvalidarCacheUsuario(usuarioID uint, empresaID uint) {
	m.Called(usuarioID, empresaID)
}

func (m *MockJustificativaRepo) WithTransaction(tx *gorm.DB) Repository {
	return m
}

type MockPontoRepo struct {
	mock.Mock
}

func (m *MockPontoRepo) SavePonto(ctx context.Context, ponto *model.RegistroPonto) error {
	args := m.Called(ctx, ponto)
	return args.Error(0)
}

func (m *MockPontoRepo) FindPontoByID(ctx context.Context, id uint, empresaID uint) (*model.RegistroPonto, error) {
	args := m.Called(ctx, id, empresaID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.RegistroPonto), args.Error(1)
}

func (m *MockPontoRepo) UpdatePonto(ctx context.Context, ponto *model.RegistroPonto) error {
	args := m.Called(ctx, ponto)
	return args.Error(0)
}

func (m *MockPontoRepo) WithTransaction(tx *gorm.DB) ponto.RegistroPontoRepository {
	return m
}

func (m *MockPontoRepo) FindPontosByUserIDAndDate(ctx context.Context, userID uint, date time.Time) ([]model.RegistroPonto, error) {
	args := m.Called(ctx, userID, date)
	return args.Get(0).([]model.RegistroPonto), args.Error(1)
}

func (m *MockPontoRepo) FindPontosByUserIDAndDateRange(ctx context.Context, userID uint, start, end time.Time) ([]model.RegistroPonto, error) {
	args := m.Called(ctx, userID, start, end)
	return args.Get(0).([]model.RegistroPonto), args.Error(1)
}

// Add other methods of PontoRepository if needed to satisfy interface
func (m *MockPontoRepo) FindPontosByPeriod(ctx context.Context, usuarioID uint, inicio, fim string) ([]model.RegistroPonto, error) {
	return nil, nil
}
func (m *MockPontoRepo) FindLastPonto(ctx context.Context, usuarioID uint) (*model.RegistroPonto, error) {
	return nil, nil
}
func (m *MockPontoRepo) CountPontosByDay(ctx context.Context, usuarioID uint, data string) (int64, error) {
	return 0, nil
}
func (m *MockPontoRepo) FindPontosByDay(ctx context.Context, usuarioID uint, data string) ([]model.RegistroPonto, error) {
	return nil, nil
}

// --- Tests ---

func TestJustificativaService_SolicitarAjuste(t *testing.T) {
	mockRepo := new(MockJustificativaRepo)
	mockPontoRepo := new(MockPontoRepo)
	service := NewService(mockRepo, mockPontoRepo, &gorm.DB{})

	justificativa := &model.Justificativa{UsuarioID: 1, Descricao: "Médico", Tipo: "PONTO_FALTANTE"}

	mockRepo.On("Create", justificativa).Return(nil)

	err := service.SolicitarAjuste(justificativa)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
