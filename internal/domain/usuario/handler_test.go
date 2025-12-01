package usuario

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// --- Mocks for Handler ---

type MockUsuarioService struct {
	mock.Mock
}

func (m *MockUsuarioService) CriarUsuarioEContrato(usuario *model.Usuario, contrato *model.Contrato, idRequisitante uint) error {
	args := m.Called(usuario, contrato, idRequisitante)
	return args.Error(0)
}

func (m *MockUsuarioService) GetAll(empresaID uint) ([]model.Usuario, error) {
	args := m.Called(empresaID)
	return args.Get(0).([]model.Usuario), args.Error(1)
}

func (m *MockUsuarioService) GetAllPaginated(empresaID uint, page int, limit int) ([]model.Usuario, int64, error) {
	args := m.Called(empresaID, page, limit)
	return args.Get(0).([]model.Usuario), args.Get(1).(int64), args.Error(2)
}

func (m *MockUsuarioService) FindByID(id uint, empresaID uint) (*model.Usuario, error) {
	args := m.Called(id, empresaID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Usuario), args.Error(1)
}

func (m *MockUsuarioService) Update(id uint, empresaID uint, dados map[string]interface{}) error {
	args := m.Called(id, empresaID, dados)
	return args.Error(0)
}

func (m *MockUsuarioService) UpdateWithHierarchy(id uint, empresaID uint, idRequisitante uint, dados map[string]interface{}) error {
	args := m.Called(id, empresaID, idRequisitante, dados)
	return args.Error(0)
}

func (m *MockUsuarioService) Delete(id uint, empresaID uint) error {
	args := m.Called(id, empresaID)
	return args.Error(0)
}

func (m *MockUsuarioService) FindAll() ([]model.Usuario, error) {
	args := m.Called()
	return args.Get(0).([]model.Usuario), args.Error(1)
}

type MockBancoHorasService struct {
	mock.Mock
}

func (m *MockBancoHorasService) GetSaldoAtualUsuario(usuarioID uint, empresaID uint) (int, error) {
	args := m.Called(usuarioID, empresaID)
	return args.Int(0), args.Error(1)
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

func TestUsuarioHandler_GetByIdHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockService := new(MockUsuarioService)
		mockFuncoes := new(MockFuncoes)
		handler := NewUsuarioHandler(mockService, mockFuncoes)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "1"}}

		// Mock expectations
		mockFuncoes.On("GetUintIDFromContext", c, "empresaID").Return(uint(1), nil)
		mockFuncoes.On("StrParaUint", "1").Return(uint(1), nil)

		expectedUser := &model.Usuario{ID: 1, Nome: "Teste"}
		mockService.On("FindByID", uint(1), uint(1)).Return(expectedUser, nil)

		handler.GetByIdHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
		mockFuncoes.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockService := new(MockUsuarioService)
		mockFuncoes := new(MockFuncoes)
		handler := NewUsuarioHandler(mockService, mockFuncoes)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "99"}}

		mockFuncoes.On("GetUintIDFromContext", c, "empresaID").Return(uint(1), nil)
		mockFuncoes.On("StrParaUint", "99").Return(uint(99), nil)

		mockService.On("FindByID", uint(99), uint(1)).Return(nil, gorm.ErrRecordNotFound)

		handler.GetByIdHandler(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
		mockFuncoes.AssertExpectations(t)
	})

	t.Run("InvalidID", func(t *testing.T) {
		mockService := new(MockUsuarioService)
		mockFuncoes := new(MockFuncoes)
		handler := NewUsuarioHandler(mockService, mockFuncoes)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "abc"}}

		mockFuncoes.On("GetUintIDFromContext", c, "empresaID").Return(uint(1), nil)
		mockFuncoes.On("StrParaUint", "abc").Return(uint(0), errors.New("invalid id"))

		handler.GetByIdHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockFuncoes.AssertExpectations(t)
	})

	t.Run("EmpresaIDFromContextError", func(t *testing.T) {
		mockService := new(MockUsuarioService)
		mockFuncoes := new(MockFuncoes)
		handler := NewUsuarioHandler(mockService, mockFuncoes)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "1"}}

		mockFuncoes.On("GetUintIDFromContext", c, "empresaID").Return(uint(0), errors.New("error getting empresaID"))

		handler.GetByIdHandler(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockFuncoes.AssertExpectations(t)
	})
}

func TestUsuarioHandler_CriarUsuarioHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockService := new(MockUsuarioService)
		mockFuncoes := new(MockFuncoes)
		handler := NewUsuarioHandler(mockService, mockFuncoes)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		requestBody := CriarUsuarioRequest{
			Nome:         "Novo Usuario",
			CPF:          "12345678901",
			Email:        "novo@empresa.com",
			Senha:        "senha123",
			LocalidadeID: 1,
			CargoID:      1,
			Salario:      3000,
			DataAdmissao: time.Now(),
		}
		jsonBody, _ := json.Marshal(requestBody)
		c.Request, _ = http.NewRequest("POST", "/usuarios", bytes.NewBuffer(jsonBody))

		mockFuncoes.On("GetUintIDFromContext", c, "userID").Return(uint(1), nil)
		mockFuncoes.On("GetUintIDFromContext", c, "empresaID").Return(uint(1), nil)
		mockService.On("CriarUsuarioEContrato", mock.AnythingOfType("*model.Usuario"), mock.AnythingOfType("*model.Contrato"), uint(1)).Return(nil)

		handler.CriarUsuarioHandler(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
		mockFuncoes.AssertExpectations(t)
	})
}

func TestUsuarioHandler_DeleteHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockService := new(MockUsuarioService)
		mockFuncoes := new(MockFuncoes)
		handler := NewUsuarioHandler(mockService, mockFuncoes)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "2"}}

		mockFuncoes.On("GetUintIDFromContext", c, "empresaID").Return(uint(1), nil)
		mockFuncoes.On("GetUintIDFromContext", c, "userID").Return(uint(1), nil)
		mockFuncoes.On("StrParaUint", "2").Return(uint(2), nil)

		// Requester permissions setup
		requester := &model.Usuario{
			ID: 1,
			Contrato: model.Contrato{
				ID: 1,
				Cargo: model.Cargo{
					ID: 1,
					Permissoes: []model.Permissao{
						{Nome: "DELETAR_USUARIO"},
					},
				},
			},
		}
		mockService.On("FindByID", uint(1), uint(1)).Return(requester, nil)
		mockService.On("Delete", uint(2), uint(1)).Return(nil)

		handler.DeleteHandler(c)

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockService.AssertExpectations(t)
		mockFuncoes.AssertExpectations(t)
	})
}

func TestUsuarioHandler_UpdateUsuarioHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockService := new(MockUsuarioService)
		mockFuncoes := new(MockFuncoes)
		handler := NewUsuarioHandler(mockService, mockFuncoes)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "2"}}

		requestBody := UpdateUsuarioRequest{
			Nome:  "Nome Atualizado",
			Email: "atualizado@empresa.com",
		}
		jsonBody, _ := json.Marshal(requestBody)
		c.Request, _ = http.NewRequest("PUT", "/usuarios/2", bytes.NewBuffer(jsonBody))

		mockFuncoes.On("GetUintIDFromContext", c, "empresaID").Return(uint(1), nil)
		mockFuncoes.On("GetUintIDFromContext", c, "userID").Return(uint(1), nil)
		mockFuncoes.On("StrParaUint", "2").Return(uint(2), nil)

		// Requester permissions setup
		requester := &model.Usuario{
			ID: 1,
			Contrato: model.Contrato{
				ID: 1,
				Cargo: model.Cargo{
					ID: 1,
					Permissoes: []model.Permissao{
						{Nome: "EDITAR_USUARIO"},
					},
				},
			},
		}
		mockService.On("FindByID", uint(1), uint(1)).Return(requester, nil)

		// Update expectation
		mockService.On("Update", uint(2), uint(1), mock.MatchedBy(func(dados map[string]interface{}) bool {
			return dados["nome"] == "Nome Atualizado" && dados["email"] == "atualizado@empresa.com"
		})).Return(nil)

		// FindByID after update expectation
		updatedUser := &model.Usuario{ID: 2, Nome: "Nome Atualizado", Email: "atualizado@empresa.com"}
		mockService.On("FindByID", uint(2), uint(1)).Return(updatedUser, nil)

		handler.UpdateUsuarioHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
		mockFuncoes.AssertExpectations(t)
	})
}
