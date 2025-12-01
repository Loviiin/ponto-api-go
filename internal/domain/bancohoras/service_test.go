package bancohoras

import (
	"context"
	"testing"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/domain/logbancohoras"
	"github.com/Loviiin/ponto-api-go/internal/domain/ponto"
	"github.com/Loviiin/ponto-api-go/internal/domain/usuario"
	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// --- Mocks ---

type MockPontoRepo struct {
	mock.Mock
}

func (m *MockPontoRepo) FindPontosByUserIDAndDate(ctx context.Context, userID uint, date time.Time) ([]model.RegistroPonto, error) {
	args := m.Called(ctx, userID, date)
	return args.Get(0).([]model.RegistroPonto), args.Error(1)
}

func (m *MockPontoRepo) FindPontosByUserIDAndDateRange(ctx context.Context, userID uint, start, end time.Time) ([]model.RegistroPonto, error) {
	args := m.Called(ctx, userID, start, end)
	return args.Get(0).([]model.RegistroPonto), args.Error(1)
}

func (m *MockPontoRepo) SavePonto(ctx context.Context, ponto *model.RegistroPonto) error {
	return nil
}
func (m *MockPontoRepo) FindPontoByID(ctx context.Context, id uint, empresaID uint) (*model.RegistroPonto, error) {
	return nil, nil
}
func (m *MockPontoRepo) UpdatePonto(ctx context.Context, ponto *model.RegistroPonto) error {
	return nil
}
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
func (m *MockPontoRepo) WithTransaction(tx *gorm.DB) ponto.RegistroPontoRepository {
	return m
}

type MockUsuarioRepo struct {
	mock.Mock
}

func (m *MockUsuarioRepo) FindByID(ctx context.Context, id uint, empresaID uint) (*model.Usuario, error) {
	args := m.Called(ctx, id, empresaID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Usuario), args.Error(1)
}

func (m *MockUsuarioRepo) Create(usuario *model.Usuario) error                { return nil }
func (m *MockUsuarioRepo) Update(id uint, dados map[string]interface{}) error { return nil }
func (m *MockUsuarioRepo) Delete(id uint) error                               { return nil }
func (m *MockUsuarioRepo) FindByEmail(email string) (*model.Usuario, error)   { return nil, nil }
func (m *MockUsuarioRepo) FindByCPF(cpf string) (*model.Usuario, error)       { return nil, nil }
func (m *MockUsuarioRepo) GetAll(empresaID uint) ([]model.Usuario, error)     { return nil, nil }
func (m *MockUsuarioRepo) GetAllPaginated(empresaID uint, page int, limit int) ([]model.Usuario, int64, error) {
	return nil, 0, nil
}
func (m *MockUsuarioRepo) FindAll() ([]model.Usuario, error)                      { return nil, nil }
func (m *MockUsuarioRepo) FindByGoogleID(googleID string) (*model.Usuario, error) { return nil, nil }
func (m *MockUsuarioRepo) WithTransaction(tx *gorm.DB) usuario.UsuarioRepository  { return m }
func (m *MockUsuarioRepo) Save(usuario *model.Usuario) error                      { return nil }
func (m *MockUsuarioRepo) InvalidarCacheUsuario(id uint, empresaID uint)          {}
func (m *MockUsuarioRepo) GetAllActive(empresaID uint) ([]model.Usuario, error)   { return nil, nil }

type MockLogRepo struct {
	mock.Mock
}

func (m *MockLogRepo) Create(log *model.LogBancoHoras) error {
	args := m.Called(log)
	return args.Error(0)
}

func (m *MockLogRepo) GetAllByUsuarioAndEmpresa(usuarioID uint, empresaID uint) ([]model.LogBancoHoras, error) {
	args := m.Called(usuarioID, empresaID)
	return args.Get(0).([]model.LogBancoHoras), args.Error(1)
}

func (m *MockLogRepo) WithTransaction(tx *gorm.DB) logbancohoras.Repository {
	return m
}

// --- Tests ---

func TestBancoHorasService_GetSaldoAtualUsuario(t *testing.T) {
	mockPontoRepo := new(MockPontoRepo)
	mockUserRepo := new(MockUsuarioRepo)
	mockLogRepo := new(MockLogRepo)

	service := NewBancoHorasService(mockPontoRepo, mockUserRepo, mockLogRepo, nil)

	// Setup User with Contract
	user := &model.Usuario{
		ID: 1,
		Contrato: model.Contrato{
			ID:                     1,
			SaldoBancoHorasMinutos: 100, // 100 minutes balance
			Cargo: model.Cargo{
				ID:                        1,
				CargaHorariaDiariaMinutos: 480, // 8 hours
			},
		},
	}

	// Expectation: Find User
	mockUserRepo.On("FindByID", mock.Anything, uint(1), uint(1)).Return(user, nil)

	// Expectation: Find Pontos for today (to calculate today's balance)
	mockPontoRepo.On("FindPontosByUserIDAndDate", mock.Anything, uint(1), mock.Anything).Return([]model.RegistroPonto{}, nil)

	got, err := service.GetSaldoAtualUsuario(1, 1)

	assert.NoError(t, err)
	// 100 (previous) - 480 (today absent) = -380
	assert.Equal(t, -380, got)

	mockUserRepo.AssertExpectations(t)
	mockPontoRepo.AssertExpectations(t)
}
