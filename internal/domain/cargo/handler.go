package cargo

import (
	"errors"
	"net/http"

	"github.com/Loviiin/ponto-api-go/internal/model"
	"github.com/Loviiin/ponto-api-go/pkg/funcoes"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CargoHandler struct {
	service   CargoService
	converter funcoes.FuncoesInterface
}

func NewCargoHandler(s CargoService, f funcoes.FuncoesInterface) *CargoHandler {
	return &CargoHandler{
		service:   s,
		converter: f,
	}
}

func (h *CargoHandler) CreateCargo(c *gin.Context) {
	type createRequest struct {
		Nome      string `json:"nome" binding:"required"`
		EmpresaID uint   `json:"empresa_id" binding:"required"`
	}
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O corpo da requisição é inválido. 'nome' e 'empresa_id' são obrigatórios."})
		return
	}

	cargo := model.Cargo{
		Nome:      req.Nome,
		EmpresaID: req.EmpresaID,
	}

	if err := h.service.Create(&cargo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao criar o cargo."})
		return
	}

	c.JSON(http.StatusCreated, cargo)
}

func (h *CargoHandler) GetAllCargos(c *gin.Context) {
	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	cargos, err := h.service.GetAllByEmpresaID(empresaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao buscar os cargos."})
		return
	}

	c.JSON(http.StatusOK, cargos)
}

func (h *CargoHandler) UpdateCargo(c *gin.Context) {
	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	cargoID, err := h.converter.StrParaUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do cargo inválido."})
		return
	}

	var dados map[string]interface{}
	if err := c.ShouldBindJSON(&dados); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Corpo da requisição (JSON) inválido."})
		return
	}

	err = h.service.Update(cargoID, empresaID, dados)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cargo não encontrado nesta empresa."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao atualizar o cargo."})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *CargoHandler) DeleteCargo(c *gin.Context) {
	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	cargoID, err := h.converter.StrParaUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do cargo inválido."})
		return
	}

	err = h.service.Delete(cargoID, empresaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cargo não encontrado nesta empresa."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao deletar o cargo."})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *CargoHandler) AddPermissionToCargo(c *gin.Context) {
	// Precisamos de obter a empresaID do token para garantir a segurança.
	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	cargoID, err := h.converter.StrParaUint(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID do cargo inválido."})
		return
	}

	permissaoID, err := h.converter.StrParaUint(c.Param("permissaoId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da permissão inválido."})
		return
	}

	// AQUI ESTÁ A CORREÇÃO: Enviamos os 3 argumentos que o serviço espera.
	err = h.service.AddPermissionToCargo(cargoID, permissaoID, empresaID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cargo ou Permissão não encontrado."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao adicionar permissão ao cargo."})
		return
	}

	c.Status(http.StatusNoContent)
}
