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

type createRequest struct {
	Nome string `json:"nome" binding:"required"`
}

// @Summary      Cria um novo cargo
// @Description  Cria um novo cargo para a empresa do usuário logado. Requer permissão 'GERENCIAR_CARGOS'.
// @Tags         Cargos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        cargo  body      createRequest  true  "Dados do novo cargo"
// @Success      201    {object}  model.Cargo
// @Failure      400    {object}  map[string]string
// @Failure      403    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /cargos [post]
func (h *CargoHandler) CreateCargo(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "O corpo da requisição é inválido. O campo 'nome' é obrigatório."})
		return
	}

	empresaID, err := h.converter.GetUintIDFromContext(c, "empresaId")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Não foi possível identificar a empresa do usuário."})
		return
	}

	cargo := model.Cargo{
		Nome:      req.Nome,
		EmpresaID: empresaID,
	}

	if err := h.service.Create(&cargo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao criar o cargo."})
		return
	}

	c.JSON(http.StatusCreated, cargo)
}

// @Summary      Lista os cargos da empresa
// @Description  Retorna uma lista de todos os cargos da empresa do usuário logado.
// @Tags         Cargos
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   model.Cargo
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /cargos [get]
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

// @Summary      Atualiza um cargo
// @Description  Atualiza os dados de um cargo. Requer permissão 'GERENCIAR_CARGOS'.
// @Tags         Cargos
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id     path      int                     true  "ID do Cargo"
// @Param        dados  body      map[string]interface{}  true  "Dados para atualização"
// @Success      204    "No Content"
// @Failure      400    {object}  map[string]string
// @Failure      403    {object}  map[string]string
// @Failure      404    {object}  map[string]string
// @Router       /cargos/{id} [put]
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

// @Summary      Deleta um cargo
// @Description  Deleta um cargo da empresa. Requer permissão 'GERENCIAR_CARGOS'.
// @Tags         Cargos
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "ID do Cargo"
// @Success      204 "No Content"
// @Failure      400 {object} map[string]string
// @Failure      403 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /cargos/{id} [delete]
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

// @Summary      Adiciona permissão a um cargo
// @Description  Associa uma permissão existente a um cargo. Requer permissão 'GERENCIAR_CARGOS'.
// @Tags         Cargos
// @Produce      json
// @Security     BearerAuth
// @Param        id           path  int  true  "ID do Cargo"
// @Param        permissaoId  path  int  true  "ID da Permissão"
// @Success      204          "No Content"
// @Failure      400          {object}  map[string]string
// @Failure      403          {object}  map[string]string
// @Failure      404          {object}  map[string]string
// @Router       /cargos/{id}/permissoes/{permissaoId} [post]
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
