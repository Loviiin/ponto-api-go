package ponto

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mocks ---

type MockPontoService struct {
	mock.Mock
}

func (m *MockPontoService) BaterPonto(ctx context.Context, usuarioID uint, empresaID uint, latitude, longitude float64) (*model.RegistroPonto, error) {
	args := m.Called(ctx, usuarioID, empresaID, latitude, longitude)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.RegistroPonto), args.Error(1)
}

func (m *MockPontoService) GetPontosDoDia(ctx context.Context, usuarioID uint, dia time.Time) ([]model.RegistroPonto, error) {
	args := m.Called(ctx, usuarioID, dia)
	return args.Get(0).([]model.RegistroPonto), args.Error(1)
}

func (m *MockPontoService) AjustarPonto(ctx context.Context, usuarioID, empresaID, adminID uint, timestamp time.Time, justificativaID *uint) (*model.RegistroPonto, error) {
	args := m.Called(ctx, usuarioID, empresaID, adminID, timestamp, justificativaID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.RegistroPonto), args.Error(1)
}

func (m *MockPontoService) EditarPonto(ctx context.Context, pontoID, empresaID uint, novoTimestamp time.Time, justificativaID *uint) (*model.RegistroPonto, error) {
	args := m.Called(ctx, pontoID, empresaID, novoTimestamp, justificativaID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.RegistroPonto), args.Error(1)
}

func (m *MockPontoService) FindPontoByID(ctx context.Context, pontoID, empresaID uint) (*model.RegistroPonto, error) {
	args := m.Called(ctx, pontoID, empresaID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.RegistroPonto), args.Error(1)
}

func (m *MockPontoService) GerarRelatorio(ctx context.Context, userID, empresaID uint, inicio, fim time.Time, formato string) ([]byte, string, string, error) {
	args := m.Called(ctx, userID, empresaID, inicio, fim, formato)
	return args.Get(0).([]byte), args.String(1), args.String(2), args.Error(3)
}

type MockJustificativaService struct {
	mock.Mock
}

func (m *MockJustificativaService) SolicitarAjuste(j *model.Justificativa) error {
	args := m.Called(j)
	return args.Error(0)
}

type MockBancoHorasService struct {
	mock.Mock
}

func (m *MockBancoHorasService) InvalidarCacheDia(usuarioID uint, empresaID uint, dia time.Time) {
	m.Called(usuarioID, empresaID, dia)
}

func (m *MockBancoHorasService) InvalidarCacheUsuario(usuarioID uint, empresaID uint) {
	m.Called(usuarioID, empresaID)
}

type MockFuncoes struct {
	mock.Mock
}

func (m *MockFuncoes) GetUintIDFromContext(c *gin.Context, key string) (uint, error) {
	args := m.Called(c, key)
	return args.Get(0).(uint), args.Error(1)
}

func (m *MockFuncoes) StrParaUint(s string) (uint, error) {
	args := m.Called(s)
	return args.Get(0).(uint), args.Error(1)
}

// --- Tests ---

func TestPontoHandler_BaterPonto(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockService := new(MockPontoService)
		mockJust := new(MockJustificativaService)
		mockBH := new(MockBancoHorasService)
		mockFuncoes := new(MockFuncoes)
		handler := NewPontoHandler(mockService, mockJust, mockBH, mockFuncoes)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Setup context values
		c.Set("userID", "1")
		c.Set("empresaID", "1")

		requestBody := BaterPontoRequest{
			Latitude:  -15.0,
			Longitude: -47.0,
		}
		jsonBody, _ := json.Marshal(requestBody)
		c.Request, _ = http.NewRequest("POST", "/pontos", bytes.NewBuffer(jsonBody))

		expectedPonto := &model.RegistroPonto{ID: 1, UsuarioID: 1}
		mockService.On("BaterPonto", mock.Anything, uint(1), uint(1), -15.0, -47.0).Return(expectedPonto, nil)

		handler.BaterPonto(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("InvalidUserID", func(t *testing.T) {
		mockService := new(MockPontoService)
		handler := NewPontoHandler(mockService, nil, nil, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", "abc") // Invalid format

		handler.BaterPonto(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestPontoHandler_GetMeusRegistos(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockService := new(MockPontoService)
		handler := NewPontoHandler(mockService, nil, nil, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", "1")
		c.Request, _ = http.NewRequest("GET", "/pontos/meus-registros", nil)

		expectedPontos := []model.RegistroPonto{{ID: 1}}
		mockService.On("GetPontosDoDia", mock.Anything, uint(1), mock.Anything).Return(expectedPontos, nil)

		handler.GetMeusRegistos(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}
