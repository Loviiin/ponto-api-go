package cargo

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// --- Mocks ---

type MockCargoService struct {
	mock.Mock
}

func (m *MockCargoService) Create(cargo *model.Cargo) error {
	args := m.Called(cargo)
	return args.Error(0)
}

func (m *MockCargoService) FindByID(id uint, empresaID uint) (*model.Cargo, error) {
	args := m.Called(id, empresaID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Cargo), args.Error(1)
}

func (m *MockCargoService) FindByName(nome string, empresaID uint) (*model.Cargo, error) {
	args := m.Called(nome, empresaID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Cargo), args.Error(1)
}

func (m *MockCargoService) GetAllByEmpresaID(empresaID uint) ([]model.Cargo, error) {
	args := m.Called(empresaID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Cargo), args.Error(1)
}

func (m *MockCargoService) Update(id uint, empresaID uint, dados map[string]interface{}) error {
	args := m.Called(id, empresaID, dados)
	return args.Error(0)
}

func (m *MockCargoService) Delete(id uint, empresaID uint) error {
	args := m.Called(id, empresaID)
	return args.Error(0)
}

func (m *MockCargoService) AddPermissionToCargo(cargoID uint, permissaoID uint, empresaID uint) error {
	args := m.Called(cargoID, permissaoID, empresaID)
	return args.Error(0)
}

func (m *MockCargoService) RemovePermissionFromCargo(cargoID uint, permissaoID uint, empresaID uint) error {
	args := m.Called(cargoID, permissaoID, empresaID)
	return args.Error(0)
}

func (m *MockCargoService) GetPermissionsByCargo(cargoID uint, empresaID uint) ([]model.Permissao, error) {
	args := m.Called(cargoID, empresaID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Permissao), args.Error(1)
}

func (m *MockCargoService) HasUsuarios(cargoID uint, empresaID uint) (bool, error) {
	args := m.Called(cargoID, empresaID)
	return args.Bool(0), args.Error(1)
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

func TestCargoHandler_CreateCargo(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockService := new(MockCargoService)
		mockFuncoes := new(MockFuncoes)
		handler := NewCargoHandler(mockService, mockFuncoes)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		requestBody := createRequest{
			Nome:            "Novo Cargo",
			NivelHierarquia: 10,
		}
		jsonBody, _ := json.Marshal(requestBody)
		c.Request, _ = http.NewRequest("POST", "/cargos", bytes.NewBuffer(jsonBody))

		mockFuncoes.On("GetUintIDFromContext", c, "empresaID").Return(uint(1), nil)
		mockService.On("FindByName", "Novo Cargo", uint(1)).Return(nil, nil)
		mockService.On("Create", mock.AnythingOfType("*model.Cargo")).Return(nil)

		handler.CreateCargo(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
		mockFuncoes.AssertExpectations(t)
	})

	t.Run("Conflict", func(t *testing.T) {
		mockService := new(MockCargoService)
		mockFuncoes := new(MockFuncoes)
		handler := NewCargoHandler(mockService, mockFuncoes)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		requestBody := createRequest{
			Nome:            "Cargo Existente",
			NivelHierarquia: 10,
		}
		jsonBody, _ := json.Marshal(requestBody)
		c.Request, _ = http.NewRequest("POST", "/cargos", bytes.NewBuffer(jsonBody))

		mockFuncoes.On("GetUintIDFromContext", c, "empresaID").Return(uint(1), nil)
		existingCargo := &model.Cargo{ID: 1, Nome: "Cargo Existente"}
		mockService.On("FindByName", "Cargo Existente", uint(1)).Return(existingCargo, nil)

		handler.CreateCargo(c)

		assert.Equal(t, http.StatusConflict, w.Code)
		mockService.AssertExpectations(t)
		mockFuncoes.AssertExpectations(t)
	})
}

func TestCargoHandler_GetCargoByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockService := new(MockCargoService)
		mockFuncoes := new(MockFuncoes)
		handler := NewCargoHandler(mockService, mockFuncoes)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "1"}}

		mockFuncoes.On("GetUintIDFromContext", c, "empresaID").Return(uint(1), nil)
		mockFuncoes.On("StrParaUint", "1").Return(uint(1), nil)
		expectedCargo := &model.Cargo{ID: 1, Nome: "Cargo 1"}
		mockService.On("FindByID", uint(1), uint(1)).Return(expectedCargo, nil)

		handler.GetCargoByID(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
		mockFuncoes.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockService := new(MockCargoService)
		mockFuncoes := new(MockFuncoes)
		handler := NewCargoHandler(mockService, mockFuncoes)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "99"}}

		mockFuncoes.On("GetUintIDFromContext", c, "empresaID").Return(uint(1), nil)
		mockFuncoes.On("StrParaUint", "99").Return(uint(99), nil)
		mockService.On("FindByID", uint(99), uint(1)).Return(nil, gorm.ErrRecordNotFound)

		handler.GetCargoByID(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
		mockFuncoes.AssertExpectations(t)
	})
}

func TestCargoHandler_GetAllCargos(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockService := new(MockCargoService)
		mockFuncoes := new(MockFuncoes)
		handler := NewCargoHandler(mockService, mockFuncoes)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		mockFuncoes.On("GetUintIDFromContext", c, "empresaID").Return(uint(1), nil)
		cargos := []model.Cargo{
			{ID: 1, Nome: "Cargo 1"},
			{ID: 2, Nome: "Cargo 2"},
		}
		mockService.On("GetAllByEmpresaID", uint(1)).Return(cargos, nil)

		handler.GetAllCargos(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
		mockFuncoes.AssertExpectations(t)
	})
}

func TestCargoHandler_UpdateCargo(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockService := new(MockCargoService)
		mockFuncoes := new(MockFuncoes)
		handler := NewCargoHandler(mockService, mockFuncoes)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "1"}}

		requestBody := map[string]interface{}{"nome": "Cargo Atualizado"}
		jsonBody, _ := json.Marshal(requestBody)
		c.Request, _ = http.NewRequest("PUT", "/cargos/1", bytes.NewBuffer(jsonBody))

		mockFuncoes.On("GetUintIDFromContext", c, "empresaID").Return(uint(1), nil)
		mockFuncoes.On("StrParaUint", "1").Return(uint(1), nil)
		mockService.On("FindByID", uint(1), uint(1)).Return(&model.Cargo{ID: 1, Nome: "Cargo Antigo", NivelHierarquia: 5}, nil)
		mockService.On("FindByName", "Cargo Atualizado", uint(1)).Return(nil, nil)
		mockService.On("Update", uint(1), uint(1), mock.Anything).Return(nil)

		handler.UpdateCargo(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
		mockFuncoes.AssertExpectations(t)
	})
}

func TestCargoHandler_DeleteCargo(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success", func(t *testing.T) {
		mockService := new(MockCargoService)
		mockFuncoes := new(MockFuncoes)
		handler := NewCargoHandler(mockService, mockFuncoes)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "1"}}

		mockFuncoes.On("GetUintIDFromContext", c, "empresaID").Return(uint(1), nil)
		mockFuncoes.On("StrParaUint", "1").Return(uint(1), nil)
		mockService.On("HasUsuarios", uint(1), uint(1)).Return(false, nil)
		mockService.On("Delete", uint(1), uint(1)).Return(nil)

		handler.DeleteCargo(c)

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockService.AssertExpectations(t)
		mockFuncoes.AssertExpectations(t)
	})
}
